package regionagogo

import (
	"encoding/json"
	"errors"
	"io"
	"log"

	"github.com/akhenakh/regionagogo/geostore"
	"github.com/golang/geo/s2"
	"github.com/kpawlik/geojson"
)

const (
	mininumViableLevel = 3 // the minimum cell level we accept
)

var (
	defaultCoverer = &s2.RegionCoverer{MinLevel: 1, MaxLevel: 30, MaxCells: 8}
)

// GeoFenceDB is the main interface to store and query your geo database
type GeoFenceDB interface {
	// returns a Fence by it's storage id
	FenceByID(loopID uint64) *Fence

	// returns the fence for the corresponding lat, lng coordinates
	StubbingQuery(lat, lng float64) *Fence

	// RectQuery perform rectangular query ur upper right bl bottom left
	RectQuery(urlat, urlng, bllat, bllng float64, limit int) (Fences, error)

	// Store a Fence into the DB
	StoreFence(rs *geostore.FenceStorage, cover []uint64) error

	// Close the DB
	Close() error
}

// ImportGeoJSONFile will load a geo json and save the polygons into
// the GeoFence for later lookup
// fields are the properties fields names you want to be associated with each fences
func ImportGeoJSONFile(gs GeoFenceDB, r io.Reader, fields []string) error {
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

		switch geom.GetType() {
		case "Polygon":
			mp := geom.(*geojson.Polygon)
			for _, p := range mp.Coordinates {
				rc, cu := preparePolygon(f, p, fields)
				if rc != nil {
					if err := gs.StoreFence(rc, cu); err != nil {
						return err
					}
				}
			}
		case "MultiPolygon":
			mp := geom.(*geojson.MultiPolygon)
			// multipolygon
			for _, m := range mp.Coordinates {
				// coordinates polygon
				p := m[0]

				rc, cu := preparePolygon(f, p, fields)
				if rc != nil {
					if err := gs.StoreFence(rc, cu); err != nil {
						return err
					}
				}
			}
		default:
			return errors.New("unknown type")
		}
	}

	return nil
}

// preparePolygon transform a geojson polygons into FenceStorage
func preparePolygon(f *geojson.Feature, p geojson.Coordinates, fields []string) (*geostore.FenceStorage, []uint64) {
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
	l := LoopFenceFromPoints(points)

	if l.IsEmpty() || l.IsFull() {
		log.Println("invalid loop", f.Properties)
		return nil, nil
	}

	covering := defaultCoverer.Covering(l)

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
		return nil, nil
	}

	var cpoints []*geostore.CPoint

	for _, p := range points {
		ll := s2.LatLngFromPoint(p)
		cpoints = append(cpoints, &geostore.CPoint{Lat: float32(ll.Lat.Degrees()), Lng: float32(ll.Lng.Degrees())})
	}

	rs := &geostore.FenceStorage{
		Points: cpoints,
		Data:   data,
	}
	return rs, cu
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
