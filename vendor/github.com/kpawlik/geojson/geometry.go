package geojson

import (
	"errors"
	"fmt"
	"reflect"
)

const (
	INIT_GEOM_CAP = 10
)

type CoordType float64

// Function which will try convert interface{} to CoordType object.
// If conversion is not possible then raise Panic.
func Coord(obj interface{}) (ct CoordType) {
	switch num := obj.(type) {
	case float64:
		ct = CoordType(num)
		break
	case int:
		ct = CoordType(num)
		break
	case float32:
		ct = CoordType(num)
		break
	case int64:
		ct = CoordType(num)
		break
	default:
		panic(fmt.Sprintf("Error: Cannot parse object: '%v' type: '%v' to CoordType!", obj, reflect.TypeOf(obj)))
	}
	return
}

// Type to represent one coordinate (x,y).
// To simplify create new instance.
// c := Coordinate{x, y}
type Coordinate [2]CoordType

// Slice of coordinates.
// Simply creation:
// c := Coordinates{{1, 2}, {2,2}}
// or
// c := Coordinates{Coordinate{1, 2}, Coordinate{2,2}}
type Coordinates []Coordinate

// Representation of set of lines
type MultiLine []Coordinates

type Geometry interface {
	GetType() string
	AddGeometry(interface{}) error
	//GetGeometry() interface{}
}

//Point coordinates are in x, y order
//(easting, northing for projected coordinates,
//longitude, latitude for geographic coordinates)
//Out example:
//   { "type": "Point", "coordinates": [100.0, 0.0] }
type Point struct {
	Type        string     `json:"type" bson:"type"`
	Coordinates Coordinate `json:"coordinates" bson:"coordinates"`
	Crs         *CRS       `json:"crs,omitempty" bson:"crs,omitempty"`
}

// Add geometry to coordinates.
// New value will replace existing
func (t *Point) AddGeometry(g interface{}) error {
	if c, ok := g.(Coordinate); ok {
		t.Coordinates = c
	} else {
		return errors.New(fmt.Sprintf("AssertionError: %v to %v",
			g, "Coordinate"))
	}
	return nil
}

func (t Point) GetType() string {
	return t.Type
}

func (t Point) GetGeometry() interface{} {
	return t.Coordinates
}

//Factory function to create new object
func NewPoint(c Coordinate) *Point {
	return &Point{Type: "Point", Coordinates: c}
}

//Coordinates of a MultiPoint are an array of positions:
// Out example:
//    { "type": "MultiPoint",
//		"coordinates": [ [100.0, 0.0], [101.0, 1.0] ]
// 	  }
type MultiPoint struct {
	Type        string      `json:"type" bson:"type"`
	Coordinates Coordinates `json:"coordinates" bson:"coordinates"`
	Crs         *CRS        `json:"crs,omitempty" bson:"crs,omitempty"`
}

// Add geometry to MultiPoint, if paremeter will be append to coordinates
func (t *MultiPoint) AddGeometry(g interface{}) error {
	switch c := g.(type) {
	case Coordinate:
		t.AddCoordinates(c)
		break
	case Coordinates:
		t.AddCoordinates(c...)
		break
	default:
		return errors.New(fmt.Sprintf("AssertionError %v", g))
	}
	return nil
}

// Add new point to MultiPoint object
func (t *MultiPoint) AddCoordinates(p ...Coordinate) {
	t.Coordinates = append(t.Coordinates, p...)
}

func (t MultiPoint) GetType() string {
	return t.Type
}

func (t MultiPoint) GetGeometry() interface{} {
	return t.Coordinates
}

//Factory function to create new object
func NewMultiPoint(coordinates Coordinates) *MultiPoint {
	if coordinates == nil {
		coordinates = make(Coordinates, 0, INIT_GEOM_CAP)
	}
	return &MultiPoint{Type: "MultiPoint", Coordinates: coordinates}
}

//Coordinates of LineString are an array of positions
// Out example:
//    { "type": "LineString",
//      "coordinates": [ [100.0, 0.0], [101.0, 1.0] ]
//    }
type LineString struct {
	Type        string      `json:"type" bson:"type"`
	Coordinates Coordinates `json:"coordinates" bson:"coordinates"`
	Crs         *CRS        `json:"crs,omitempty" bson:"crs,omitempty"`
}

// Add new position to LineString
func (t *LineString) AddCoordinates(c ...Coordinate) {
	t.Coordinates = append(t.Coordinates, c...)
}

// Add new position to LineString
func (t *LineString) AddGeometry(g interface{}) error {
	switch c := g.(type) {
	case Coordinate:
		t.AddCoordinates(c)
		break
	case Coordinates:
		t.AddCoordinates(c...)
		break
	default:
		return errors.New(fmt.Sprintf("AssertionError %v", g))
	}
	return nil
}

func (t LineString) GetType() string {
	return t.Type
}

func (t LineString) GetGeometry() interface{} {
	return t.Coordinates
}

//Factory function to create new object with points
func NewLineString(coordinates Coordinates) *LineString {
	if coordinates == nil {
		coordinates = make(Coordinates, 0, INIT_GEOM_CAP)
	}
	return &LineString{Type: "LineString", Coordinates: coordinates}
}

// For type "MultiLineString", the "coordinates" member must be an array
// of LineString coordinate arrays.
// Out example:
//	 { "type": "MultiLineString",
//	  "coordinates": [
//	      [ [100.0, 0.0], [101.0, 1.0] ],
//	      [ [102.0, 2.0], [103.0, 3.0] ]
//	    ]
//	  }
type MultiLineString struct {
	Type        string    `json:"type" bson:"type"`
	Coordinates MultiLine `json:"coordinates" bson:"coordinates"`
	Crs         *CRS      `json:"crs,omitempty" bson:"crs,omitempty"`
}

// Add new line or line to MultiLineString
func (t *MultiLineString) AddGeometry(g interface{}) error {
	switch c := g.(type) {
	case MultiLine:
		t.AddCoordinates(c...)
		break
	case Coordinates:
		t.AddCoordinates(c)
		break
	default:
		return errors.New(fmt.Sprintf("AssertionError %v", g))
	}
	return nil
}

// Add collection of coordinates to MultiLineString
// new data are append
func (t *MultiLineString) AddCoordinates(coordinates ...Coordinates) {
	t.Coordinates = append(t.Coordinates, coordinates...)
}

func (t MultiLineString) GetType() string {
	return t.Type
}

func (t MultiLineString) GetGeometry() interface{} {
	return t.Coordinates
}

// Factory function for type MultiLineString
func NewMultiLineString(coordinates MultiLine) *MultiLineString {
	if coordinates == nil {
		coordinates = make(MultiLine, 0, INIT_GEOM_CAP)
	}
	return &MultiLineString{Type: "MultiLineString", Coordinates: coordinates}
}

// For type "Polygon", the "coordinates" member must be an array of LinearRing
// coordinate arrays. For Polygons with multiple rings, the first must be
// the exterior ring and any others must be interior rings or holes.
// Out example:
//{ "type": "Polygon",
//  "coordinates": [
//    				[ [100.0, 0.0], [101.0, 0.0], [101.0, 1.0],
//					[100.0, 1.0], [100.0, 0.0] ]
//    ]
// }
type Polygon struct {
	Type        string    `json:"type" bson:"type"`
	Coordinates MultiLine `json:"coordinates,float" bson:"coordinates"`
	Crs         *CRS      `json:"crs,omitempty" bson:"crs,omitempty"`
}

// Add new polygon  or hole to Polygon
func (t *Polygon) AddGeometry(g interface{}) error {
	switch c := g.(type) {
	case MultiLine:
		t.AddCoordinates(c...)
		break
	case Coordinates:
		t.AddCoordinates(c)
		break
	default:
		return errors.New(fmt.Sprintf("AssertionError %v", g))
	}
	return nil
}

// add new polygon or hole.
// new values are append
func (t *Polygon) AddCoordinates(coordinates ...Coordinates) {
	t.Coordinates = append(t.Coordinates, coordinates...)
}

func (t Polygon) GetType() string {
	return t.Type
}

func (t Polygon) GetGeometry() interface{} {
	return t.Coordinates
}

// factory function
func NewPolygon(coordinates MultiLine) *Polygon {
	if coordinates == nil {
		coordinates = make(MultiLine, 0, INIT_GEOM_CAP)
	}
	return &Polygon{Type: "Polygon", Coordinates: coordinates}
}

// For type "MultiPolygon", the "coordinates" member must
// be an array of Polygon coordinate arrays.
// Out example
//{ "type": "MultiPolygon",
//  "coordinates": [
//    [[[102.0, 2.0], [103.0, 2.0], [103.0, 3.0], [102.0, 3.0], [102.0, 2.0]]],
//    [[[100.0, 0.0], [101.0, 0.0], [101.0, 1.0], [100.0, 1.0], [100.0, 0.0]],
//     [[100.2, 0.2], [100.8, 0.2], [100.8, 0.8], [100.2, 0.8], [100.2, 0.2]]]
//    ]
//  }
type MultiPolygon struct {
	Type        string      `json:"type" bson:"type"`
	Coordinates []MultiLine `json:"coordinates" bson:"coordinates"`
	Crs         *CRS        `json:"crs,omitempty" bson:"crs,omitempty"`
}

// add new polygon or hole.
// new values are append
func (t *MultiPolygon) AddCoordinates(lines ...MultiLine) {
	t.Coordinates = append(t.Coordinates, lines...)
}

// Add new polygon  or hole to Polygon
func (t *MultiPolygon) AddGeometry(g interface{}) error {
	switch c := g.(type) {
	case []MultiLine:
		t.AddCoordinates(c...)
		break
	case MultiLine:
		t.AddCoordinates(c)
		break
	default:
		return errors.New(fmt.Sprintf("AssertionError %v", g))
	}
	return nil
}

func (t MultiPolygon) GetType() string {
	return t.Type
}

func (t MultiPolygon) GetGeometry() interface{} {
	return t.Coordinates
}

// factory function
func NewMultiPolygon(coordinates []MultiLine) *MultiPolygon {
	if coordinates == nil {
		coordinates = make([]MultiLine, 0, INIT_GEOM_CAP)
	}
	return &MultiPolygon{Type: "MultiPolygon", Coordinates: coordinates}
}

// A GeoJSON object with type "GeometryCollection" is a geometry object
// which represents a collection of geometry objects.
// A geometry collection must have a member with the name
// "geometries". The value corresponding to "geometries" is an array.
// Each element in this array is a GeoJSON geometry object.
// Out example:
//{ "type": "GeometryCollection",
//  "geometries": [
//    { "type": "Point",
//      "coordinates": [100.0, 0.0]
//      },
//    { "type": "LineString",
//      "coordinates": [ [101.0, 0.0], [102.0, 1.0] ]
//      }
//  ]
//}
type GeometryCollection struct {
	Type       string        `json:"type" bson:"type"`
	Geometries []interface{} `json:"geometries" bson:"geometries"`
	Crs        *CRS          `json:"crs,omitempty" bson:"crs,omitempty"`
}

// new values are append
func (t *GeometryCollection) AddGeometries(g ...interface{}) {
	t.Geometries = append(t.Geometries, g...)
}

// Add new geometry  or hole to GeometryCollection
func (t *GeometryCollection) AddGeometry(g interface{}) error {
	switch c := g.(type) {
	case []interface{}:
		t.AddGeometries(c...)
		break
	case interface{}:
		t.AddGeometries(c)
		break
	default:
		return errors.New(fmt.Sprintf("AssertionError %v", g))
	}
	return nil
}

func (t GeometryCollection) GetType() string {
	return t.Type
}

// factory function
func NewGeometryCollection(g []interface{}) *GeometryCollection {
	if g == nil {
		g = make([]interface{}, 0, 10)
	}
	return &GeometryCollection{Type: "GeometryCollection", Geometries: g}
}
