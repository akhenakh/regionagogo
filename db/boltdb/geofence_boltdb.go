package boltdb

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"log"
	"math"
	"sort"

	"github.com/Workiva/go-datastructures/augmentedtree"
	region "github.com/akhenakh/regionagogo"
	"github.com/akhenakh/regionagogo/geostore"
	"github.com/boltdb/bolt"
	"github.com/golang/geo/s2"
	"github.com/golang/protobuf/proto"
	lru "github.com/hashicorp/golang-lru"
)

const (
	defaultLoopBucket       = "loop"
	defaultCoverBucket      = "cover"
	earthCircumferenceMeter = 40075017
)

var (
	defaultCoverer = s2.RegionCoverer{MinLevel: 1, MaxLevel: 30, MaxCells: 8}
)

// GeoFenceBoltDB provides an in memory index and boltdb query engine for fences lookup
type GeoFenceBoltDB struct {
	augmentedtree.Tree
	*bolt.DB
	cache       *lru.Cache
	loopBucket  []byte
	coverBucket []byte
	debug       bool
	ro          bool
}

// GeoSearchOption used to pass options to NewGeoSearch
type GeoFenceBoltDBOption func(*geoFenceBoltDBOptions)

type geoFenceBoltDBOptions struct {
	maxCachedEntries uint
	debug            bool
	loopBucket       []byte
	coverBucket      []byte
	ro               bool
}

// WithLoopBucket set the loop bucket name
func WithLoopBucket(loopBucket string) GeoFenceBoltDBOption {
	return func(o *geoFenceBoltDBOptions) {
		o.loopBucket = []byte(loopBucket)
	}
}

// WithCoverBucket set the cover bucket name
func WithCoverBucket(coverBucket string) GeoFenceBoltDBOption {
	return func(o *geoFenceBoltDBOptions) {
		o.coverBucket = []byte(coverBucket)
	}
}

// WithCachedEntries enable an LRU cache default is disabled
func WithCachedEntries(maxCachedEntries uint) GeoFenceBoltDBOption {
	return func(o *geoFenceBoltDBOptions) {
		o.maxCachedEntries = maxCachedEntries
	}
}

// WithDebug enable debug
func WithDebug(debug bool) GeoFenceBoltDBOption {
	return func(o *geoFenceBoltDBOptions) {
		o.debug = debug
	}
}

// WithReadOnly enable read only db
func WithReadOnly(ro bool) GeoFenceBoltDBOption {
	return func(o *geoFenceBoltDBOptions) {
		o.ro = ro
	}
}

// NewGeoFenceBoltDB creates or reopen a bolt geo database
func NewGeoFenceBoltDB(dbpath string, opts ...GeoFenceBoltDBOption) (*GeoFenceBoltDB, error) {
	var geoOpts geoFenceBoltDBOptions

	for _, opt := range opts {
		opt(&geoOpts)
	}

	db, err := bolt.Open(dbpath, 0600, &bolt.Options{ReadOnly: geoOpts.ro})
	if err != nil {
		return nil, err
	}

	return NewGeoFenceIdx(db, opts...)
}

// NewGeoFenceIdx a geo index over a BoltDB storage
func NewGeoFenceIdx(db *bolt.DB, opts ...GeoFenceBoltDBOption) (*GeoFenceBoltDB, error) {
	var geoOpts geoFenceBoltDBOptions

	for _, opt := range opts {
		opt(&geoOpts)
	}

	gs := &GeoFenceBoltDB{
		Tree:        augmentedtree.New(1),
		DB:          db,
		debug:       geoOpts.debug,
		ro:          geoOpts.ro,
		loopBucket:  geoOpts.loopBucket,
		coverBucket: geoOpts.coverBucket,
	}

	if geoOpts.maxCachedEntries != 0 {
		cache, err := lru.New(int(geoOpts.maxCachedEntries))
		if err != nil {
			return nil, err
		}
		gs.cache = cache
	}

	if len(gs.loopBucket) == 0 {
		gs.loopBucket = []byte(defaultLoopBucket)
	}

	if len(gs.coverBucket) == 0 {
		gs.coverBucket = []byte(defaultCoverBucket)
	}

	// create bucket if we have write permission
	if !geoOpts.ro {
		if errdb := db.Update(func(tx *bolt.Tx) error {
			if _, errtx := tx.CreateBucketIfNotExists(gs.loopBucket); errtx != nil {
				return fmt.Errorf("create bucket: %s", errtx)
			}
			if _, errtx := tx.CreateBucketIfNotExists(gs.coverBucket); errtx != nil {
				return fmt.Errorf("create bucket: %s", errtx)
			}
			return nil
		}); errdb != nil {
			return nil, errdb
		}
	}

	if err := gs.importGeoData(); err != nil {
		return nil, err
	}

	return gs, nil
}

// index indexes each cells of the cover and set its loopID
func (gs *GeoFenceBoltDB) index(fc *geostore.FenceCover, loopID uint64) {
	for _, cell := range fc.Cellunion {
		s2interval := &region.S2Interval{CellID: s2.CellID(cell)}
		intervals := gs.Query(s2interval)
		found := false

		if len(intervals) != 0 {
			for _, existInterval := range intervals {
				if existInterval.LowAtDimension(1) == s2interval.LowAtDimension(1) &&
					existInterval.HighAtDimension(1) == s2interval.HighAtDimension(1) {
					// update existing interval
					existS2Interval := existInterval.(*region.S2Interval)
					if gs.debug {
						log.Printf("added %d to existing interval %s containing %v", loopID, existS2Interval, existS2Interval.LoopIDs)
					}

					existS2Interval.LoopIDs = append(existS2Interval.LoopIDs, loopID)
					found = true
					break
				}
			}
		}

		if !found {
			// create new interval with current loop
			s2interval.LoopIDs = []uint64{loopID}
			gs.Add(s2interval)
			if gs.debug {
				log.Printf("added %v to new interval %s", s2interval.LoopIDs, s2interval)
			}
		}
	}
}

// importGeoData loads all existing cells into the segment tree
func (gs *GeoFenceBoltDB) importGeoData() error {
	var count int
	err := gs.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(gs.coverBucket)
		cur := b.Cursor()

		// load the cell ranges into the tree
		var fc geostore.FenceCover
		for k, v := cur.First(); k != nil; k, v = cur.Next() {
			count++
			err := proto.Unmarshal(v, &fc)
			if err != nil {
				return err
			}
			if gs.debug {
				log.Println("read", fc.Cellunion)
			}

			// read back the loopID from the key
			var loopID uint64
			buf := bytes.NewBuffer(k)
			err = binary.Read(buf, binary.BigEndian, &loopID)
			if err != nil {
				return err
			}

			gs.index(&fc, loopID)
		}

		return nil
	})
	if err != nil {
		return err
	}

	if count != 0 {
		log.Println("loaded", count, "existing fences")
	} else {
		log.Println("initialized database")
	}

	return nil
}

// FenceByID returns a region from DB by its id
func (gs *GeoFenceBoltDB) FenceByID(loopID uint64) *region.Fence {
	// TODO: refactor as Fence ?
	if gs.cache != nil {
		if val, ok := gs.cache.Get(loopID); ok {
			r, _ := val.(*region.Fence)
			return r
		}
	}
	var rs *geostore.FenceStorage
	err := gs.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(gs.loopBucket)
		v := b.Get(itob(loopID))

		if len(v) == 0 {
			return nil
		}

		var frs geostore.FenceStorage
		err := proto.Unmarshal(v, &frs)
		if err != nil {
			return err
		}
		rs = &frs
		return nil
	})
	if err != nil {
		return nil
	}
	r := region.NewFenceFromStorage(rs)
	if gs.cache != nil && r != nil {
		gs.cache.Add(loopID, r)
	}
	return r
}

// StubbingQuery returns the fence for the corresponding lat, lng point
func (gs *GeoFenceBoltDB) StubbingQuery(lat, lng float64, opts ...region.QueryOptionsFunc) (region.Fences, error) {
	// the CellID at L30
	q := s2.CellIDFromLatLng(s2.LatLngFromDegrees(lat, lng))

	i := &region.S2Interval{CellID: q}

	if gs.debug {
		log.Println("lookup", lat, lng, q)
	}
	r := gs.Tree.Query(i)

	var foundFence *region.Fence

	var queryOpts region.QueryOptions
	for _, opt := range opts {
		opt(&queryOpts)
	}

	var res []*region.Fence

	for _, itv := range r {
		sitv := itv.(*region.S2Interval)
		if gs.debug {
			log.Println("found possible solution", sitv, sitv.LoopIDs)
		}

		for _, loopID := range sitv.LoopIDs {
			fence := gs.FenceByID(loopID)
			if fence != nil && fence.Loop.ContainsPoint(q.Point()) {
				res = append(res, fence)
				if foundFence == nil {
					if gs.debug {
						log.Println("found matching fence", sitv, sitv.LoopIDs, fence.Loop.NumEdges())
					}
					foundFence = fence
					continue
				}

				// a fence can include a smaller fence
				// return only the one that is contained in the other if asked
				if !queryOpts.MultipleFences {
					// we take the 1st vertex of the fence.Loop if it is contained in previousLoop
					// region loop  is more precise
					if foundFence.Loop.ContainsPoint(fence.Loop.Vertex(0)) {
						foundFence = fence
					}
				}
			}

		}

	}

	if !queryOpts.MultipleFences && foundFence != nil {
		return []*region.Fence{foundFence}, nil
	}

	// Order fences by their size
	sort.Sort(sort.Reverse(region.BySize(res)))

	return res, nil
}

// RectQuery perform rectangular query ur upper right bl bottom left
func (gs *GeoFenceBoltDB) RectQuery(urlat, urlng, bllat, bllng float64, opts ...region.QueryOptionsFunc) (region.Fences, error) {
	rect := s2.RectFromLatLng(s2.LatLngFromDegrees(bllat, bllng))
	rect = rect.AddPoint(s2.LatLngFromDegrees(urlat, urlng))

	rc := &s2.RegionCoverer{MaxLevel: 30, MaxCells: 1}
	covering := rc.Covering(rect)
	if len(covering) != 1 {
		return nil, errors.New("impossible covering")
	}
	i := &region.S2Interval{CellID: covering[0]}
	r := gs.Tree.Query(i)

	fences := make(map[uint64]*region.Fence)

	for _, itv := range r {
		sitv := itv.(*region.S2Interval)
		for _, loopID := range sitv.LoopIDs {
			var region *region.Fence
			if v, ok := fences[loopID]; ok {
				region = v
			} else {
				region = gs.FenceByID(loopID)
				// testing the found loop is actually inside the rect
				// (since we are using only one large cover it may be outside)
				if rect.Contains(region.Loop.RectBound()) {
					fences[loopID] = region
				}
			}

		}
	}

	var res []*region.Fence
	for _, v := range fences {
		res = append(res, v)
	}
	return region.Fences(res), nil
}

// RadiusQuery is performing a radius query
func (gs *GeoFenceBoltDB) RadiusQuery(lat, lng, radius float64, opts ...region.QueryOptionsFunc) (region.Fences, error) {
	center := s2.PointFromLatLng(s2.LatLngFromDegrees(lat, lng))
	cap := s2.CapFromCenterArea(center, s2RadialAreaMeters(radius))
	covering := defaultCoverer.Covering(cap)

	var res []*region.Fence

	fencesIds := make(map[uint64]struct{})

	for _, cellID := range covering {
		i := &region.S2Interval{CellID: cellID}
		r := gs.Tree.Query(i)

		for _, itv := range r {
			sitv := itv.(*region.S2Interval)
			for _, loopID := range sitv.LoopIDs {
				fencesIds[loopID] = struct{}{}
			}
		}
	}

	for k := range fencesIds {
		fence := gs.FenceByID(k)
		if fence != nil {
			res = append(res, fence)
		}
	}

	return res, nil
}

func s2RadialAreaMeters(radius float64) float64 {
	r := (radius / earthCircumferenceMeter) * math.Pi * 2
	return (math.Pi * r * r)
}

// StoreFence stores a fence into the database and load its index in memory
func (gs *GeoFenceBoltDB) StoreFence(fs *geostore.FenceStorage, cover []uint64) error {
	if gs.ro {
		return errors.New("db is in read only mode")
	}
	return gs.Update(func(tx *bolt.Tx) error {
		loopB := tx.Bucket(gs.loopBucket)
		coverBucket := tx.Bucket(gs.coverBucket)

		loopID, err := loopB.NextSequence()
		if err != nil {
			return err
		}

		buf, err := proto.Marshal(fs)
		if err != nil {
			return err
		}

		if gs.debug {
			log.Println("inserted", loopID, fs.Data, cover)
		}

		// convert our loopID to bigendian to be used as key
		k := itob(loopID)

		err = loopB.Put(k, buf)
		if err != nil {
			return err
		}

		// inserting into cover index using the same key
		fc := &geostore.FenceCover{Cellunion: cover}
		bufc, err := proto.Marshal(fc)
		if err != nil {
			return err
		}

		// also load into memory
		gs.index(fc, loopID)

		return coverBucket.Put(k, bufc)
	})
}

// itob returns an 8-byte big endian representation of v.
func itob(v uint64) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, v)
	return b
}
