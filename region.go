package regionagogo

import (
	"github.com/golang/geo/s2"
	"github.com/kpawlik/geojson"
)

// Regions a slice of *Region (type used mainly to return one GeoJSON of the regions)
type Regions []*Region

// Region is region for memory use
// it contains an S2 loop and the associated metadata
type Region struct {
	Data map[string]string `json:"data"`
	Loop *s2.Loop          `json:"-"`
}

// NewRegionFromStorage returns a Region from a RegionStorage
// Regions can be extended, RegionStorage is a protocol buffer instance
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

// ToGeoJSON transforms a Region to a valid GeoJSON
func (r *Region) ToGeoJSON() *geojson.FeatureCollection {
	var geo geojson.FeatureCollection

	var cs []geojson.Coordinate

	points := r.Loop.Vertices()
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

	for k, v := range r.Data {
		properties[k] = v
	}

	geo.Features = []*geojson.Feature{
		&geojson.Feature{
			Type:       "Feature",
			Geometry:   poly,
			Properties: properties,
		},
	}
	geo.Type = "FeatureCollection"

	return &geo
}

// ToGeoJSON transforms a set of Region to a valid GeoJSON
func (r *Regions) ToGeoJSON() *geojson.FeatureCollection {
	var geo geojson.FeatureCollection
	var features []*geojson.Feature

	for _, region := range *r {
		var cs []geojson.Coordinate

		points := region.Loop.Vertices()
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

		for k, v := range region.Data {
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
