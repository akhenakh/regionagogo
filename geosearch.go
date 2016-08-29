package regionagogo

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"

	"encoding/binary"

	"github.com/Workiva/go-datastructures/augmentedtree"
	"github.com/boltdb/bolt"
	"github.com/golang/geo/s2"
	"github.com/golang/protobuf/proto"
	"github.com/kpawlik/geojson"
)

const (
	mininumViableLevel = 3 // the minimum cell level we accept
	loopBucket         = "loop"
)

// GeoSearch provides in memory index and query engine for fences lookup
type GeoSearch struct {
	augmentedtree.Tree
	*bolt.DB
	Debug bool
}

// Regions a slice of *Region (type used mainly to return one GeoJSON of the regions)
type Regions []*Region

// Region is region for memory use
// it contains an S2 loop and the associated metadata
type Region struct {
	Data map[string]string `json:"data"`
	Loop *s2.Loop          `json:"-"`
}

func NewRegionFromStorage(rs *RegionStorage) *Region {
	if rs == nil {
		return nil
	}

	points := make([]s2.Point, len(rs.Points))
	for i, c := range rs.Points {
		// Points in Storage are lat lng points
		ll := s2.LatLngFromDegrees(float64(c.Lat), float64(c.Lng))
		point := s2.PointFromLatLng(ll)
		points[i] = point
	}

	l := LoopRegionFromPoints(points)

	return &Region{Data: rs.Data, Loop: l.Loop}
}

// NewGeoSearch
func NewGeoSearch(dbpath string) (*GeoSearch, error) {
	db, err := bolt.Open(dbpath, 0600, nil)
	if err != nil {
		return nil, err
	}
	db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucket([]byte(loopBucket))
		if err != nil {
			return fmt.Errorf("create bucket: %s", err)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	gs := &GeoSearch{
		Tree: augmentedtree.New(1),
		DB:   db,
	}

	return gs, nil
}

// ImportGeoData open the DB and load all cells into the segment tree
func (gs *GeoSearch) ImportGeoData() error {
	var count int
	err := gs.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(loopBucket))
		cur := b.Cursor()

		// load the cell ranges into the tree
		var rs RegionStorage
		for k, v := cur.First(); k != nil; k, v = cur.Next() {
			count++
			err := proto.Unmarshal(v, &rs)
			if err != nil {
				return err
			}
			if gs.Debug {
				log.Println("read", rs.Id, rs.Data, rs.Cellunion)
			}
			for _, cell := range rs.Cellunion {
				s2interval := &S2Interval{CellID: s2.CellID(cell)}
				intervals := gs.Query(s2interval)
				if len(intervals) == 0 {
					// create new interval with current loop
					s2interval.LoopIDs = []uint64{rs.Id}
					gs.Add(s2interval)
					if gs.Debug {
						log.Println("added new interval", s2interval)
					}
				} else {
					for _, existInterval := range intervals {
						if existInterval.LowAtDimension(1) == s2interval.LowAtDimension(1) &&
							existInterval.HighAtDimension(1) == s2interval.HighAtDimension(1) {
							if gs.Debug {
								log.Println("added to existing interval", s2interval, rs.Id)
							}
							s2interval.LoopIDs = append(s2interval.LoopIDs, rs.Id)
						}
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

// TODO: refactor as Fence ?
func (gs *GeoSearch) RegionByID(loopID uint64) *Region {
	var rs *RegionStorage
	err := gs.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(loopBucket))
		v := b.Get(itob(loopID))
		var frs RegionStorage
		err := proto.Unmarshal(v, &frs)
		if err != nil {
			return err
		}
		log.Println("DEBUG", itob(loopID), frs)
		rs = &frs
		return nil
	})
	if err != nil {
		return nil
	}

	return NewRegionFromStorage(rs)
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
				log.Println("testing region", loopID)
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

// importGeoJSONFile will load a geo json and save the polygons into
// a msgpack file named geodata
// fields to lookup for in GeoJSON
func (gs *GeoSearch) ImportGeoJSONFile(r io.Reader, fields []string) error {
	var geo geojson.FeatureCollection

	d := json.NewDecoder(r)
	if err := d.Decode(&geo); err != nil {
		return err
	}

	var geoData GeoDataStorage

	cl := make(map[s2.CellID][]int64)

	for _, f := range geo.Features {
		geom, err := f.GetGeometry()
		if err != nil {
			return err
		}

		rc := &s2.RegionCoverer{MinLevel: 1, MaxLevel: 30, MaxCells: 8}

		switch geom.GetType() {
		case "MultiPolygon":
			mp := geom.(*geojson.MultiPolygon)
			// multipolygon
			for _, m := range mp.Coordinates {
				// polygon
				var points []s2.Point
				var cpoints []*CPoint
				// For type "MultiPolygon", the "coordinates" member must be an array of Polygon coordinate arrays.
				// "Polygon", the "coordinates" member must be an array of LinearRing coordinate arrays.
				// For Polygons with multiple rings, the first must be the exterior ring and any others must be interior rings or holes.

				if len(m) < 1 {
					continue
				}

				p := m[0]
				// coordinates

				// reverse the slice
				for i := len(p)/2 - 1; i >= 0; i-- {
					opp := len(p) - 1 - i
					p[i], p[opp] = p[opp], p[i]
				}

				for i, c := range p {
					ll := s2.LatLngFromDegrees(float64(c[1]), float64(c[0]))
					point := s2.PointFromLatLng(ll)
					points = append(points, point)
					// do not add cpoint on storage (first point is last point)
					if i == len(p)-1 {
						break
					}
					cpoints = append(cpoints, &CPoint{Lat: float32(c[1]), Lng: float32(c[0])})
				}

				// TODO: parallelized region cover
				l := LoopRegionFromPoints(points)

				if l.IsEmpty() || l.IsFull() {
					log.Println("invalid loop")
					continue
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
						log.Print("cell level too big", v.Level(), mp)
						invalidLoop = true
					}
					cu[i] = uint64(v)
				}

				// do not insert big loop
				if invalidLoop {
					break
				}

				rs := &RegionStorage{
					Points:    cpoints,
					Cellunion: cu,
					Data:      data,
				}

				gs.Update(func(tx *bolt.Tx) error {
					b := tx.Bucket([]byte(loopBucket))

					id, err := b.NextSequence()
					if err != nil {
						return err
					}
					rs.Id = id
					if gs.Debug {
						log.Println("inserted", rs.Id, data, covering)
					}

					buf, err := proto.Marshal(rs)
					if err != nil {
						return err
					}

					log.Println("STORE", itob(rs.Id), rs)
					return b.Put(itob(rs.Id), buf)
				})

				geoData.Rs = append(geoData.Rs, rs)
			}
		default:
			return errors.New("unknown type")
		}

	}

	for k, v := range cl {
		geoData.Cl = append(geoData.Cl, &CellIDLoopStorage{Cell: uint64(k), Loops: v})
	}

	log.Println("imported", len(geoData.Rs), "regions")

	b, err := proto.Marshal(&geoData)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile("geodata", b, 0644)

	return nil
}

// itob returns an 8-byte big endian representation of v.
func itob(v uint64) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, v)
	return b
}
