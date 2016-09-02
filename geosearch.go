package regionagogo

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"

	"github.com/Workiva/go-datastructures/augmentedtree"
	"github.com/boltdb/bolt"
	"github.com/golang/geo/s2"
	"github.com/golang/protobuf/proto"
	"github.com/hashicorp/golang-lru"
	"github.com/kpawlik/geojson"
)

const (
	mininumViableLevel = 3 // the minimum cell level we accept
	loopBucket         = "loop"
	coverBucket        = "cover"
)

// GeoSearch provides in memory index and query engine for fences lookup
type GeoSearch struct {
	augmentedtree.Tree
	*bolt.DB
	cache *lru.Cache
	Debug bool
}

// GeoSearchOption used to pass options to NewGeoSearch
type GeoSearchOption func(*geoSearchOptions)

type geoSearchOptions struct {
	maxCachedEntries uint
}

// WithCachedEntries enable an LRU cache default is disabled
func WithCachedEntries(maxCachedEntries uint) GeoSearchOption {
	return func(o *geoSearchOptions) {
		o.maxCachedEntries = maxCachedEntries
	}
}

// NewGeoSearch creates a new geo database, needs a writable path for BoltDB
func NewGeoSearch(dbpath string, opts ...GeoSearchOption) (*GeoSearch, error) {
	db, err := bolt.Open(dbpath, 0600, nil)
	if err != nil {
		return nil, err
	}
	err = db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucket([]byte(loopBucket))
		if err != nil {
			return fmt.Errorf("create bucket: %s", err)
		}
		_, err = tx.CreateBucket([]byte(coverBucket))
		if err != nil {
			return fmt.Errorf("create bucket: %s", err)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	var geoOpts geoSearchOptions

	for _, opt := range opts {
		opt(&geoOpts)
	}

	gs := &GeoSearch{
		Tree: augmentedtree.New(1),
		DB:   db,
	}

	if geoOpts.maxCachedEntries != 0 {
		cache, err := lru.New(int(geoOpts.maxCachedEntries))
		if err != nil {
			return nil, err
		}
		gs.cache = cache
	}

	return gs, nil
}

// ImportGeoData open the DB and load all cells into the segment tree
func (gs *GeoSearch) ImportGeoData() error {
	var count int
	err := gs.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(coverBucket))
		cur := b.Cursor()

		// load the cell ranges into the tree
		var rc RegionCover
		for k, v := cur.First(); k != nil; k, v = cur.Next() {
			count++
			err := proto.Unmarshal(v, &rc)
			if err != nil {
				return err
			}
			if gs.Debug {
				log.Println("read", rc.Cellunion)
			}

			// read back the loopID from the key
			var id uint64
			buf := bytes.NewBuffer(k)
			err = binary.Read(buf, binary.BigEndian, &id)
			if err != nil {
				return err
			}

			for _, cell := range rc.Cellunion {
				s2interval := &S2Interval{CellID: s2.CellID(cell)}
				intervals := gs.Query(s2interval)
				found := false

				if len(intervals) != 0 {
					for _, existInterval := range intervals {
						if existInterval.LowAtDimension(1) == s2interval.LowAtDimension(1) &&
							existInterval.HighAtDimension(1) == s2interval.HighAtDimension(1) {
							if gs.Debug {
								log.Println("added to existing interval", s2interval, id)
							}
							s2interval.LoopIDs = append(s2interval.LoopIDs, id)
							found = true
							break
						}
					}
				}

				if !found {
					// create new interval with current loop
					s2interval.LoopIDs = []uint64{id}
					gs.Add(s2interval)
					if gs.Debug {
						log.Println("added new interval", s2interval, id)
					}
				}
			}
		}

		return nil
	})
	if err != nil {
		return err
	}

	log.Println("loaded", count, "regions")

	return nil
}

// RegionByID returns a region from DB by its id
func (gs *GeoSearch) RegionByID(loopID uint64) *Region {
	// TODO: refactor as Fence ?
	if gs.cache != nil {
		if val, ok := gs.cache.Get(loopID); ok {
			r, _ := val.(*Region)
			return r
		}
	}
	var rs *RegionStorage
	err := gs.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(loopBucket))
		v := b.Get(itob(loopID))

		var frs RegionStorage
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
	r := NewRegionFromStorage(rs)
	if gs.cache != nil && r != nil {
		gs.cache.Add(loopID, r)
	}
	return r
}

// StubbingQuery returns the region for the corresponding lat, lng point
func (gs *GeoSearch) StubbingQuery(lat, lng float64) *Region {
	q := s2.CellIDFromLatLng(s2.LatLngFromDegrees(lat, lng))
	i := &S2Interval{CellID: q}

	if gs.Debug {
		log.Println("lookup", lat, lng, q)
	}
	r := gs.Tree.Query(i)

	var foundRegion *Region

	for _, itv := range r {
		sitv := itv.(*S2Interval)
		if gs.Debug {
			log.Println("found", sitv, sitv.LoopIDs)
		}

		// a region can include a smaller region
		// return only the one that is contained in the other
		for _, loopID := range sitv.LoopIDs {
			region := gs.RegionByID(loopID)
			if region != nil && region.Loop.ContainsPoint(q.Point()) {
				if foundRegion == nil {
					foundRegion = region
					continue
				}
				// we take the 1st vertex of the region.Loop if it is contained in previousLoop
				// region loop  is more precise
				if foundRegion.Loop.ContainsPoint(region.Loop.Vertex(0)) {
					foundRegion = region
				}
			}
		}
	}

	return foundRegion
}

// RectQuery perform rectangular query ur upper right bl bottom left
// return a list of loopID
func (gs *GeoSearch) RectQuery(urlat, urlng, bllat, bllng float64, limit int) (Regions, error) {
	rect := s2.RectFromLatLng(s2.LatLngFromDegrees(bllat, bllng))
	rect = rect.AddPoint(s2.LatLngFromDegrees(urlat, urlng))

	rc := &s2.RegionCoverer{MaxLevel: 30, MaxCells: 1}
	covering := rc.Covering(rect)
	if len(covering) != 1 {
		return nil, errors.New("impossible covering")
	}
	i := &S2Interval{CellID: covering[0]}
	r := gs.Tree.Query(i)

	regions := make(map[uint64]*Region)

	for _, itv := range r {
		sitv := itv.(*S2Interval)
		for _, loopID := range sitv.LoopIDs {
			var region *Region
			if v, ok := regions[loopID]; ok {
				region = v
			} else {
				region = gs.RegionByID(loopID)
			}
			// testing the found loop is actually inside the rect
			// (since we are using only one large cover it may be outside)
			if rect.Contains(region.Loop.RectBound()) {
				regions[loopID] = region
			}
		}
	}

	var res []*Region
	for _, v := range regions {
		res = append(res, v)
	}
	return Regions(res), nil
}

// ImportGeoJSONFile will load a geo json and save the polygons into
// the database for later lookup f
// fields are the properties fields you want to be associated with each fences
func (gs *GeoSearch) ImportGeoJSONFile(r io.Reader, fields []string) error {
	var geo geojson.FeatureCollection

	d := json.NewDecoder(r)
	if err := d.Decode(&geo); err != nil {
		return err
	}

	for _, f := range geo.Features {
		geom, err := f.GetGeometry()
		if err != nil {
			return err
		}

		rc := &s2.RegionCoverer{MinLevel: 1, MaxLevel: 30, MaxCells: 8}

		switch geom.GetType() {
		case "Polygon":
			mp := geom.(*geojson.Polygon)
			for _, p := range mp.Coordinates {
				err := gs.insertPolygon(f, p, rc, fields)
				if err != nil {
					return err
				}
			}
		case "MultiPolygon":
			mp := geom.(*geojson.MultiPolygon)
			// multipolygon
			for _, m := range mp.Coordinates {
				// coordinates polygon
				p := m[0]

				err := gs.insertPolygon(f, p, rc, fields)
				if err != nil {
					return err
				}
			}
		default:
			return errors.New("unknown type")
		}
	}

	return nil
}

func (gs *GeoSearch) insertPolygon(f *geojson.Feature, p geojson.Coordinates, rc *s2.RegionCoverer, fields []string) error {
	if isClockwisePolygon(p) {
		reversePolygon(p)
	}
	// polygon
	// do not add last point in storage (first point is last point)
	points := make([]s2.Point, len(p)-1)

	// For type "MultiPolygon", the "coordinates" member must be an array of Polygon coordinate arrays.
	// "Polygon", the "coordinates" member must be an array of LinearRing coordinate arrays.
	// For Polygons with multiple rings, the first must be the exterior ring and any others must be interior rings or holes.

	for i := 0; i < len(p)-1; i++ {
		ll := s2.LatLngFromDegrees(float64(p[i][1]), float64(p[i][0]))
		points[i] = s2.PointFromLatLng(ll)
	}

	// TODO: parallelized region cover
	l := LoopRegionFromPoints(points)

	if l.IsEmpty() || l.IsFull() {
		log.Println("invalid loop", f.Properties)
		return nil
	}

	covering := rc.Covering(l)

	data := make(map[string]string)
	for _, field := range fields {
		if v, ok := f.Properties[field].(string); !ok {
			log.Println("can't find field on", f.Properties)
		} else {
			data[field] = v
		}
	}

	cu := make([]uint64, len(covering))
	var invalidLoop bool

	for i, v := range covering {
		// added a security there if the level is too high it probably means the polygon is bogus
		// this to avoid a large cell to cover everything
		if v.Level() < mininumViableLevel {
			log.Print("cell level too big", v.Level(), data)
			invalidLoop = true
		}
		cu[i] = uint64(v)
	}

	// do not insert big loop
	if invalidLoop {
		return nil
	}

	var cpoints []*CPoint

	for _, p := range points {
		ll := s2.LatLngFromPoint(p)
		cpoints = append(cpoints, &CPoint{Lat: float32(ll.Lat.Degrees()), Lng: float32(ll.Lng.Degrees())})
	}

	rs := &RegionStorage{
		Points: cpoints,
		Data:   data,
	}

	return gs.Update(func(tx *bolt.Tx) error {
		loopB := tx.Bucket([]byte(loopBucket))
		coverBucket := tx.Bucket([]byte(coverBucket))

		id, err := loopB.NextSequence()
		if err != nil {
			return err
		}

		buf, err := proto.Marshal(rs)
		if err != nil {
			return err
		}

		if gs.Debug {
			log.Println("inserted", id, data, cu)
		}

		// convert our loopID to bigendian to be used as key
		k := itob(id)

		err = loopB.Put(k, buf)
		if err != nil {
			return err
		}

		// inserting into cover index using the same key
		bufc, err := proto.Marshal(&RegionCover{Cellunion: cu})
		if err != nil {
			return err
		}

		return coverBucket.Put(k, bufc)
	})
}

func isClockwisePolygon(p geojson.Coordinates) bool {
	sum := 0.0
	for i, coord := range p[:len(p)-1] {
		next := p[i+1]
		sum += float64((next[0] - coord[0]) * (next[1] + coord[1]))
	}
	if sum == 0 {
		return true
	}
	return sum > 0
}

func reversePolygon(p geojson.Coordinates) {
	for i := len(p)/2 - 1; i >= 0; i-- {
		opp := len(p) - 1 - i
		p[i], p[opp] = p[opp], p[i]
	}
}

// itob returns an 8-byte big endian representation of v.
func itob(v uint64) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, v)
	return b
}
