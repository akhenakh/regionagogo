package geojson

import (
	"errors"
	"fmt"
)

// lowest and highest values fo coordinates
type BoundingBox []float64

// A GeoJSON object with the type "Feature" is a feature object.
// - A feature object must have a member with the name "geometry". 
//	 The value of the geometry member is a geometry object or a JSON null value.
// - A feature object must have a member with the name "properties". 
//   The value of the properties member is an object (any JSON object 
//   or a JSON null value).
// - If a feature has a commonly used identifier, that identifier should be 
//   included as a member of the feature object with the name "id".
type Feature struct {
	Type       string                 `json:"type" bson:"type"`
	Id         interface{}            `json:"id,omitempty" bson:"id,omitempty"`
	Geometry   interface{}            `json:"geometry" bson:"geometry"`
	Properties map[string]interface{} `json:"properties" bson:"properties"`
	Bbox       BoundingBox            `json:"bbox,omitempty" bson:"bbox,omitempty"`
	Crs        *CRS                   `json:"crs,omitempty" bson:"crs,omitempty"`
}

func (t *Feature) GetGeometry() (Geometry, error) {
	gi := t.Geometry
	return parseGeometry(gi)
}

// Factory constructor method 
func NewFeature(geom Geometry, properties map[string]interface{},
	id interface{}) *Feature {
	return &Feature{Type: "Feature",
		Geometry:   geom,
		Properties: properties,
		Id:         id}
}

// An object of type "FeatureCollection" must have a member with the name 
// "features". The value corresponding to "features" is an array. 
// Each element in the array is a Feature object.
type FeatureCollection struct {
	Type     string      `json:"type" bson:"type"`
	Features []*Feature  `json:"features" bson:"features"`
	Bbox     BoundingBox `json:"bbox,omitempty" bson:"bbox,omitempty"`
	Crs      *CRS        `json:"crs,omitempty" bson:"crs,omitempty"`
}

func (t *FeatureCollection) AddFeatures(f ...*Feature) {
	if f == nil {
		t.Features = make([]*Feature, 0, 100)
	}
	t.Features = append(t.Features, f...)
}

// factory funcion
func NewFeatureCollection(features []*Feature) *FeatureCollection {
	return &FeatureCollection{Type: "FeatureCollection", Features: features}
}

// The coordinate reference system (CRS) of a GeoJSON object 
// is determined by its "crs" member. 
type CRS struct {
	Type       string            `json:"type" bson:"type"`
	Properties map[string]string `json:"properties" bson:"properties"`
}

//Example:
//"crs": {
//  "type": "name",
//  "properties": {
//    "name": "urn:ogc:def:crs:OGC:1.3:CRS84"
//    }
//  }
func NewNamedCRS(name string) *CRS {
	return &CRS{Type: "name", Properties: map[string]string{"name": name}}
}

// Exaples:
//"crs": {
//  "type": "link",
//  "properties": {
//    "href": "http://example.com/crs/42",
//    "type": "proj4"
//    }
//  }
//
//"crs": {
//  "type": "link",
//  "properties": {
//    "href": "data.crs",
//    "type": "ogcwkt"
//    }
//  }
func NewLinkedCRS(href, typ string) *CRS {
	return &CRS{Type: "link",
		Properties: map[string]string{"href": href, "type": typ}}
}

func parseCoordinate(c interface{}) (coord Coordinate, err error) {
	defer func() {
		if e := recover(); e != nil {
			err = errors.New(fmt.Sprintf("%v", e))
			return
		}
	}()
	coordinate, ok := c.([]interface{})
	if !ok || len(coordinate) != 2 {
		err = errors.New(fmt.Sprintf("Error unmarshal %v to coordinates", c))
		return
	}
	x := Coord(coordinate[0])
	y := Coord(coordinate[1])
	coord = Coordinate{x, y}
	return
}

func parseCoordinates(obj interface{}) (coords Coordinates, err error) {
	if c, ok := obj.([]interface{}); !ok || len(c) < 1 {
		err = errors.New(fmt.Sprintf("ParseErrr: Coordinates parse error, %v", obj))
		return
	} else {
		coords = make(Coordinates, len(c), len(c))
		for i, v := range c {
			if coords[i], err = parseCoordinate(v); err != nil {
				return
			}
		}
	}
	return
}

func parseMultiLine(obj interface{}) (coords MultiLine, err error) {
	if c, ok := obj.([]interface{}); !ok || len(c) < 1 {
		err = errors.New(fmt.Sprintf("ParseErrr: MultiLine parse error, %v", obj))
		return
	} else {
		coords = make(MultiLine, len(c), len(c))
		for i, v := range c {
			if coords[i], err = parseCoordinates(v); err != nil {
				return
			}
		}
	}
	return
}

func parsePoint(obj interface{}) (p *Point, err error) {
	var c Coordinate
	if c, err = parseCoordinate(obj); err != nil {
		return
	} else {
		p = NewPoint(c)
	}
	return
}

func parseLineString(obj interface{}) (ls *LineString, err error) {
	var cc Coordinates
	if cc, err = parseCoordinates(obj); err != nil {
		return
	}
	ls = NewLineString(cc)
	return
}

func parseMultiPoint(obj interface{}) (mp *MultiPoint, err error) {
	var cc Coordinates
	if cc, err = parseCoordinates(obj); err != nil {
		return
	}
	mp = NewMultiPoint(cc)
	return
}

func parseMultiLineString(obj interface{}) (mls *MultiLineString, err error) {
	var ml MultiLine
	if ml, err = parseMultiLine(obj); err != nil {
		return
	}
	mls = NewMultiLineString(ml)
	return
}

func parsePolygon(obj interface{}) (pol *Polygon, err error) {
	var pl MultiLine
	if pl, err = parseMultiLine(obj); err != nil {
		return
	}
	pol = NewPolygon(pl)
	return
}

func parseMultiPolygon(obj interface{}) (mpol *MultiPolygon, err error) {
	var ml []MultiLine
	if si, ok := obj.([]interface{}); !ok {
		err = errors.New("Parse Error: parse multi polygon error")
		return
	} else {
		ml = make([]MultiLine, len(si), len(si))
		for i, slice := range si {
			if ml[i], err = parseMultiLine(slice); err != nil {
				return
			}
		}
	}
	mpol = NewMultiPolygon(ml)
	return
}

func parseGeometry(gi interface{}) (geom Geometry, err error) {
	switch g := gi.(type) {
	case map[string]interface{}:
		coord := g["coordinates"]
		switch typ := g["type"]; typ {
		case "Point":
			return parsePoint(coord)
		case "LineString":
			return parseLineString(coord)
		case "MultiPoint":
			return parseMultiPoint(coord)
		case "MultiLineString":
			return parseMultiLineString(coord)
		case "Polygon":
			return parsePolygon(coord)
		case "MultiPolygon":
			return parseMultiPolygon(coord)
		case "GeometryCollection":
			return parseGeometryCollection(coord)
		default:
			err = errors.New(fmt.Sprintf("ParseError: Unknown geometry type %s", typ))
			break
		}
	}
	return geom, err
}

func parseGeometryCollection(obj interface{}) (gc *GeometryCollection, err error) {
	gc = NewGeometryCollection([]interface{}{})
	if si, ok := obj.([]interface{}); !ok {
		err = errors.New("ParseError: Error durring parse geometry collection")
		return
	} else {
		for i := 0; i < len(si); i++ {
			gc.AddGeometry(si[i])
		}
	}
	return
}
