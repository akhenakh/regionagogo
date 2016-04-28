package regionagogo

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"strings"

	"gopkg.in/vmihailenco/msgpack.v2"

	"github.com/Workiva/go-datastructures/augmentedtree"
	"github.com/golang/geo/s2"
	"github.com/kpawlik/geojson"
)

// GeoSearch provides in memory storage and query engine for regions lookup
type GeoSearch struct {
	augmentedtree.Tree
	rm    map[int]Region
	Debug bool
}

// Region is region for memory use
type Region struct {
	Data map[string]string `json:"data"`
	L    *s2.Loop          `json:"-"`
}

// GeoData is used to pack the data in a msgpack file
type GeoData struct {
	RS []RegionStorage     `msgpack:"rs"`
	CL []CellIDLoopStorage `msgpack:"cl"`
}

// CellIDLoopStorage is a cell with associated loops used for storage
type CellIDLoopStorage struct {
	C     s2.CellID `msgpack:"c"`
	Loops []int     `msgpack:"l"`
}

// RegionStorage is a region representation for storage use
type RegionStorage struct {
	Data         map[string]string `msgpack:"d"`
	Code         string            `msgpack:"i"`
	Points       []CPoint          `msgpack:"p"`
	s2.CellUnion `msgpack:"c"`
}

// CPoint is a []float64 used as coordinates
type CPoint struct {
	Coordinate []float64 `msgpack:"c"`
}

// NewGeoSearch
func NewGeoSearch() *GeoSearch {
	gs := &GeoSearch{
		Tree: augmentedtree.New(1),
		rm:   make(map[int]Region),
	}

	return gs
}

// ImportGeoData loads geodata file into a map loopID->Region
// fills the segment tree for fast lookup
func (gs *GeoSearch) ImportGeoData(b []byte) error {
	var gd GeoData

	err := msgpack.Unmarshal(b, &gd)
	if err != nil {
		return err
	}

	for loopID, r := range gd.RS {
		var points []s2.Point

		for _, c := range r.Points {
			ll := s2.LatLngFromDegrees(c.Coordinate[0], c.Coordinate[1])
			point := s2.PointFromLatLng(ll)
			points = append(points, point)
		}
		// add 1st point as last point to close the shape
		points = append(points, points[0])

		// load the loops into memory
		l := s2.LoopFromPoints(points)
		gs.rm[loopID] = Region{Data: r.Data, L: l}

	}

	// load the cell ranges into the tree
	for _, cLoop := range gd.CL {
		gs.Add(&S2Interval{CellID: cLoop.C, LoopIDs: cLoop.Loops})
	}

	// free some space
	gd.CL = []CellIDLoopStorage{}

	log.Println("loaded", len(gs.rm), "regions")

	return nil
}

// Query returns the country for the corresponding lat, lng point
func (gs *GeoSearch) Query(lat, lng float64) map[string]string {
	q := s2.CellIDFromLatLng(s2.LatLngFromDegrees(lat, lng))
	i := &S2Interval{CellID: q}
	r := gs.Tree.Query(i)

	matchLoopID := -1

	for _, itv := range r {
		sitv := itv.(*S2Interval)
		if gs.Debug {
			fmt.Println("found", sitv, sitv.LoopIDs)
		}

		// a region can include a smaller region
		// return only the one that is contained in the other

		for _, loopID := range sitv.LoopIDs {

			if gs.rm[loopID].L.ContainsPoint(q.Point()) {

				if matchLoopID == -1 {
					matchLoopID = loopID
				} else {
					foundLoop := gs.rm[loopID].L
					previousLoop := gs.rm[matchLoopID].L

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
		return region.Data
	}

	return nil
}

// importGeoJSONFile will load a geo json and save the polygons into
// a msgpack file named geodata
// fields to lookup for in GeoJSON
func ImportGeoJSONFile(filename string, debug bool, fields []string) error {
	var loopID int

	b, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}
	var geo geojson.FeatureCollection

	err = json.Unmarshal(b, &geo)
	if err != nil {
		return err
	}

	var geoData GeoData

	cl := make(map[s2.CellID][]int)

	for _, f := range geo.Features {
		geom, err := f.GetGeometry()
		if err != nil {
			return err
		}

		switch geom.GetType() {
		case "MultiPolygon":
			mp := geom.(*geojson.MultiPolygon)
			// multipolygon
			for _, m := range mp.Coordinates {
				// polygon
				var points []s2.Point
				var cpoints []CPoint
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
					cpoints = append(cpoints, CPoint{Coordinate: []float64{float64(c[1]), float64(c[0])}})
				}

				l := s2.LoopFromPoints(points)

				if l.IsEmpty() || l.IsFull() {
					log.Println("invalid loop")
					continue
				}

				//rb := l.RectBound()
				rc := &s2.RegionCoverer{MinLevel: 1, MaxLevel: 30, MaxCells: 8}
				covering := rc.Covering(l)

				data := make(map[string]string)
				for _, field := range fields {
					if v, ok := f.Properties[field].(string); !ok {
						log.Println("can't find field on", f.Properties)
					} else {
						data[field] = v
					}
				}

				if debug {
					fmt.Println("import", loopID, data)
				}

				r := RegionStorage{
					Data:      data,
					Points:    cpoints,
					CellUnion: covering,
				}

				geoData.RS = append(geoData.RS, r)

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
		geoData.CL = append(geoData.CL, CellIDLoopStorage{C: k, Loops: v})
	}

	log.Println("imported", filename, len(geoData.RS), "regions")

	b, err = msgpack.Marshal(geoData)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile("geodata", b, 0644)
	if err != nil {
		return err
	}

	return nil
}

// FieldFlag reusable parse Value to create import command
type FieldFlag struct {
	Fields []string
}

func (ff *FieldFlag) String() string {
	return fmt.Sprint(ff.Fields)
}

func (ff *FieldFlag) Set(value string) error {
	if len(ff.Fields) > 0 {
		return fmt.Errorf("The field flag is already set")
	}

	ff.Fields = strings.Split(value, ",")
	return nil
}
