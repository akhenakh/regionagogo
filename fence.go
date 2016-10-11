package regionagogo

import (
	"github.com/akhenakh/regionagogo/geostore"
	"github.com/golang/geo/s2"
	"github.com/kpawlik/geojson"
)

// Fences a slice of *Fence (type used mainly to return one GeoJSON of the regions)
type Fences []*Fence

// Fence is an s2 represented FenceStorage
// it contains an S2 loop and the associated metadata
type Fence struct {
	Data map[string]string `json:"data"`
	Loop *s2.Loop          `json:"-"`
}

// NewFenceFromStorage returns a Fence from a FenceStorage
// Fence can be extended, FenceStorage is a protocol buffer instance
func NewFenceFromStorage(rs *geostore.FenceStorage) *Fence {
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

	l := LoopFenceFromPoints(points)

	return &Fence{Data: rs.Data, Loop: l.Loop}
}

// ToGeoJSON transforms a Region to a valid GeoJSON
func (f *Fence) ToGeoJSON() *geojson.FeatureCollection {
	var geo geojson.FeatureCollection

	var cs []geojson.Coordinate

	points := f.Loop.Vertices()
	for i, p := range points {
		ll := s2.LatLngFromPoint(p)
		c := geojson.Coordinate{
			geojson.CoordType(ll.Lng.Degrees()),
			geojson.CoordType(ll.Lat.Degrees()),
		}
		if i == len(points)-1 {
			break
		}
		cs = append(cs, c)
	}

	coordinates := []geojson.Coordinates{cs}

	poly := &geojson.Polygon{
		Type:        "Polygon",
		Coordinates: coordinates,
	}

	properties := make(map[string]interface{})

	for k, v := range f.Data {
		properties[k] = v
	}

	geo.Features = []*geojson.Feature{
		{
			Type:       "Feature",
			Geometry:   poly,
			Properties: properties,
		},
	}
	geo.Type = "FeatureCollection"

	return &geo
}

// ToGeoJSON transforms a set of Fences to a valid GeoJSON
func (f *Fences) ToGeoJSON() *geojson.FeatureCollection {
	var geo geojson.FeatureCollection
	var features []*geojson.Feature

	for _, fence := range *f {
		var cs []geojson.Coordinate

		points := fence.Loop.Vertices()
		for _, p := range points {
			ll := s2.LatLngFromPoint(p)
			c := geojson.Coordinate{
				geojson.CoordType(ll.Lng.Degrees()),
				geojson.CoordType(ll.Lat.Degrees()),
			}
			cs = append(cs, c)
		}

		coordinates := []geojson.Coordinates{cs}

		poly := &geojson.Polygon{
			Type:        "Polygon",
			Coordinates: coordinates,
		}

		properties := make(map[string]interface{})

		for k, v := range fence.Data {
			properties[k] = v
		}

		f := &geojson.Feature{
			Type:       "Feature",
			Geometry:   poly,
			Properties: properties,
		}

		features = append(features, f)
	}

	geo.Features = features
	geo.Type = "FeatureCollection"

	return &geo
}

type BySize []*Fence

func (d BySize) Len() int      { return len(d) }
func (d BySize) Swap(i, j int) { d[i], d[j] = d[j], d[i] }
func (d BySize) Less(i, j int) bool {
	// use approximated area to decide ordering
	return d[i].Loop.RectBound().Area() < d[j].Loop.RectBound().Area()
}
