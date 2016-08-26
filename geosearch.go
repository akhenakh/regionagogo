package regionagogo

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"

	"bufio"
	"os"

	"github.com/Workiva/go-datastructures/augmentedtree"
	"github.com/golang/geo/s2"
	"github.com/golang/protobuf/proto"
	"github.com/kpawlik/geojson"
)

const (
	mininumViableLevel = 3 // the minimum cell level we accept
)

// GeoSearch provides in memory storage and query engine for region lookup
type GeoSearch struct {
	augmentedtree.Tree
	rm    map[int64]Region
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

// NewGeoSearch
func NewGeoSearch() *GeoSearch {
	gs := &GeoSearch{
		Tree: augmentedtree.New(1),
		rm:   make(map[int64]Region),
	}

	return gs
}

// ImportGeoData loads geodata file into a map loopID->Region
// fills the segment tree for fast lookup
func (gs *GeoSearch) ImportGeoData(filename string) error {
	var gd GeoDataStorage

	b, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}

	err = proto.Unmarshal(b, &gd)
	if err != nil {
		return err
	}

	for loopID, r := range gd.Rs {
		var points []s2.Point

		for _, c := range r.Points {
			ll := s2.LatLngFromDegrees(float64(c.Lat), float64(c.Lng))
			point := s2.PointFromLatLng(ll)
			points = append(points, point)
		}
		// add 1st point as last point to close the shape
		points = append(points, points[0])

		// load the loops into memory
		l := s2.LoopFromPoints(points)
		if l.IsEmpty() || l.IsFull() {
			return errors.New("invalid loop")
		}
		gs.rm[int64(loopID)] = Region{Data: r.Data, Loop: l}

	}

	// load the cell ranges into the tree
	for _, cLoop := range gd.Cl {
		gs.Add(&S2Interval{CellID: s2.CellID(cLoop.Cell), LoopIDs: cLoop.Loops})
	}

	// free some space
	gd.Cl = nil
	log.Println("loaded", len(gs.rm), "regions")

	return nil
}

// Query returns the country for the corresponding lat, lng point
func (gs *GeoSearch) StubbingQuery(lat, lng float64) *Region {
	q := s2.CellIDFromLatLng(s2.LatLngFromDegrees(lat, lng))
	i := &S2Interval{CellID: q}
	r := gs.Tree.Query(i)

	matchLoopID := int64(-1)

	for _, itv := range r {
		sitv := itv.(*S2Interval)
		if gs.Debug {
			fmt.Println("found", sitv, sitv.LoopIDs)
		}

		// a region can include a smaller region
		// return only the one that is contained in the other
		for _, loopID := range sitv.LoopIDs {
			if gs.rm[loopID].Loop.ContainsPoint(q.Point()) {

				if matchLoopID == -1 {
					matchLoopID = loopID
				} else {
					foundLoop := gs.rm[loopID].Loop
					previousLoop := gs.rm[matchLoopID].Loop

					// we take the 1st vertex of the foundloop if it is contained in previousLoop
					// foundLoop one is more precise
					if previousLoop.ContainsPoint(foundLoop.Vertex(0)) {
						matchLoopID = loopID
					}
				}
			}
		}
	}

	if matchLoopID != -1 {
		region := gs.rm[matchLoopID]
		return &region
	}

	return nil
}

// importGeoJSONFile will load a geo json and save the polygons into
// a msgpack file named geodata
// fields to lookup for in GeoJSON
func ImportGeoJSONFile(filename string, debug bool, fields []string) error {
	var loopID int64

	fi, err := os.Open(filename)
	defer fi.Close()
	if err != nil {
		return err
	}

	var geo geojson.FeatureCollection
	r := bufio.NewReader(fi)

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
		case "Polygon":
			mp := geom.(*geojson.Polygon)
			// multipolygon
			for _, p := range mp.Coordinates {
				// polygon
				var points []s2.Point
				var cpoints []*CPoint
				// For type "MultiPolygon", the "coordinates" member must be an array of Polygon coordinate arrays.
				// "Polygon", the "coordinates" member must be an array of LinearRing coordinate arrays.
				// For Polygons with multiple rings, the first must be the exterior ring and any others must be interior rings or holes.

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
						log.Print("cell level too big", v.Level(), loopID)
						invalidLoop = true
					}
					cu[i] = uint64(v)
				}

				// do not insert big loop
				if invalidLoop {
					break
				}

				if debug {
					fmt.Println("import", loopID, data)
				}

				rs := &RegionStorage{
					Data:      data,
					Points:    cpoints,
					Cellunion: cu,
				}

				geoData.Rs = append(geoData.Rs, rs)

				for _, cell := range covering {
					cl[cell] = append(cl[cell], loopID)
				}

				loopID = loopID + 1
			}

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
						log.Print("cell level too big", v.Level(), loopID)
						invalidLoop = true
					}
					cu[i] = uint64(v)
				}

				// do not insert big loop
				if invalidLoop {
					break
				}

				if debug {
					fmt.Println("import", loopID, data)
				}

				rs := &RegionStorage{
					Data:      data,
					Points:    cpoints,
					Cellunion: cu,
				}

				geoData.Rs = append(geoData.Rs, rs)

				for _, cell := range covering {
					cl[cell] = append(cl[cell], loopID)
				}

				loopID = loopID + 1
			}
		default:
			return errors.New("unknown type")
		}

	}

	for k, v := range cl {
		geoData.Cl = append(geoData.Cl, &CellIDLoopStorage{Cell: uint64(k), Loops: v})
	}

	log.Println("imported", filename, len(geoData.Rs), "regions")

	b, err := proto.Marshal(&geoData)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile("geodata", b, 0644)

	return nil
}
