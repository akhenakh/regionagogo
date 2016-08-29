/*
Package contains funcionality to simple and fast 
create objects collection to serialize them as GeoJSON data.
*/
package geojson

import (
	"encoding/gob"
	"encoding/json"
)

// Marshal object to json string representation
func Marshal(object interface{}) (data string, err error) {
	if t, err := json.Marshal(object); err != nil {
		data = ""
	} else {
		data = string(t)
	}
	return
}

func Register() {
	gob.Register(Feature{})
	gob.Register(FeatureCollection{})
	gob.Register(GeometryCollection{})
	gob.Register(Point{})
	gob.Register(MultiPoint{})
	gob.Register(LineString{})
	gob.Register(MultiLineString{})
	gob.Register(Polygon{})
	gob.Register(MultiPolygon{})
}
