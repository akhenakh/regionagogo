package geojson

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
)

func TestCoord(t *testing.T) {
	tt := test{t}
	tt.AssertEq(1.0, float64(Coord(1)), "Coord 1 test")
	tt.AssertEq(1.0, float64(Coord(1.0)), "Coord 1 test")
	func() {
		defer func() {
			err := recover()
			tt.AssertNeq(err, nil, "Panic expected")
		}()
		Coord("asd")
	}()
}

func TestPoint(t *testing.T) {
	var (
		p   *Point
		err error
	)
	tt := test{t}
	res := `{"type":"Point","coordinates":[1.1,2.2]}`
	p = NewPoint(Coordinate{1.1, 2.2})
	tt.AssertMarshal(p, res, "Point 1")
	p = NewPoint(Coordinate{})
	err = p.AddGeometry(nil)
	tt.AssertNeq(err, nil, "Should raise assertion error")
}

func TestUnmarshalPoint(t *testing.T) {
	var (
		err error
		p   Point
	)
	res := `{"type":"Point","coordinates":[1.1,2.2]}`
	tt := test{t}
	err = json.Unmarshal([]byte(res), &p)
	tt.AssertEq(err, nil, "No error expected 1")
	tt.AssertEq("Point", p.Type)
	tt.AssertEq(Coord(1.1), p.Coordinates[0])
	tt.AssertEq(Coord(2.2), p.Coordinates[1])
}

func TestLineString(t *testing.T) {
	tt := test{t}
	res := `{"type":"LineString","coordinates":[]}`
	mp := NewLineString(nil)
	tt.AssertMarshal(mp, res)
}

func TestUnmarshalLineString(t *testing.T) {
	var (
		err error
		ls  LineString
	)
	res := `{"type":"LineString","coordinates":[[1,2], [3.0, 4.1]]}`
	tt := test{t}
	err = json.Unmarshal([]byte(res), &ls)
	tt.AssertEq(err, nil, "No error expected 1")
	tt.AssertEq("LineString", ls.Type)
	tt.AssertEq(len(ls.Coordinates), 2, "Wrong size of coordinates")
	tt.AssertCoordinates(Coordinate{1, 2}, ls.Coordinates[0])
	tt.AssertCoordinates(Coordinate{3, 4.1}, ls.Coordinates[1])
}

func TestMultiPoint(t *testing.T) {
	tt := test{t}
	res := `{"type":"MultiPoint","coordinates":[]}`
	mp := NewMultiPoint(nil)
	tt.AssertMarshal(mp, res)
}

func TestMultiLineString(t *testing.T) {
	tt := test{t}
	res := `{"type":"MultiLineString","coordinates":[]}`
	mp := NewMultiLineString(nil)
	tt.AssertMarshal(mp, res)
}

func TestPolygon(t *testing.T) {
	tt := test{t}
	res := `{"type":"Polygon","coordinates":[]}`
	mp := NewPolygon(nil)
	tt.AssertMarshal(mp, res)
}

func TestPolygonWithHoles(t *testing.T) {
	ps1 := `[[100.01,0],[101,0],[101,1],[100,1],[100,0]]`
	ps2 := `[[100.2,0.2],[100.8,0.2],[100.8,0.8],[100.2,0.8],[100.2,0.2]]`
	ps := `{"type":"Polygon","coordinates":[%s,%s]}`
	res := fmt.Sprintf(ps, ps1, ps2)
	l1 := Coordinates{Coordinate{100.01, 0.0}, Coordinate{101, 0}, Coordinate{101.0, 1.0},
		Coordinate{100.0, 1.0}, Coordinate{100.0, 0.0}}
	l2 := Coordinates{Coordinate{100.2, 0.2}, Coordinate{100.8, 0.2}, Coordinate{100.8, 0.8},
		Coordinate{100.2, 0.8}, Coordinate{100.2, 0.2}}
	ml := MultiLine{l1, l2}
	mp := NewPolygon(ml)
	tt := test{t}
	tt.AssertMarshal(mp, res)

}
func MultiCoordinatesInterfaceTest(factory func(Coordinates) TestInterfaceMultiCoord,
	res string, t test) {
	var (
		mc  TestInterfaceMultiCoord
		err error
	)
	mc = factory(nil)
	mc.AddCoordinates(Coordinate{1.1, 2.2})
	mc.AddCoordinates(Coordinate{3.3, 4.4})
	t.AssertMarshal(mc, res, "1")
	p := Coordinates{{1.1, 2.2}, {3.3, 4.4}}
	mc = factory(p)
	t.AssertMarshal(mc, res, "2")
	mc = factory(nil)
	mc.AddCoordinates(p...)
	t.AssertMarshal(mc, res, "3")
	mc = factory(nil)
	mc.AddCoordinates(Coordinate{1.1, 2.2}, Coordinate{3.3, 4.4})
	t.AssertMarshal(mc, res, "4")
	mc = factory(nil)
	mc.AddGeometry(Coordinate{1.1, 2.2})
	mc.AddGeometry(Coordinate{3.3, 4.4})
	t.AssertMarshal(mc, res, "5")
	mc = factory(nil)
	mc.AddGeometry(Coordinates{{1.1, 2.2}, {3.3, 4.4}})
	t.AssertMarshal(mc, res, "6")
	mc = factory(nil)
	mc.AddGeometry(nil)
	err = mc.AddGeometry(nil)
	t.AssertNeq(err, nil, "Should raise assertion error")
}

func MultiLineCoordinatesInterfaceTest(factory func([]Coordinates) TestInterfaceMultiLineCoords,
	res string, t test) {
	var (
		mc  TestInterfaceMultiLineCoords
		err error
	)
	c1 := Coordinate{1.1, 2.2}
	c2 := Coordinate{3.3, 4.4}
	c := Coordinates{c1, c2}
	cc := MultiLine{c, c}
	mc = factory(nil)
	mc.AddCoordinates(c)
	mc.AddCoordinates(c)
	t.AssertMarshal(mc, res, "1")
	mc = factory(nil)
	mc.AddCoordinates(cc...)
	t.AssertMarshal(mc, res, "2")
	mc = factory(nil)
	mc.AddGeometry(c)
	mc.AddGeometry(c)
	t.AssertMarshal(mc, res, "3")
	mc = factory(nil)
	mc.AddGeometry(cc)
	t.AssertMarshal(mc, res, "4")
	mc = factory(nil)
	mc.AddGeometry(nil)
	err = mc.AddGeometry(nil)
	t.AssertNeq(err, nil, "Should raise assertion error")
}

func TestMultiPointInterface(t *testing.T) {
	tt := test{t}
	res := `{"type":"MultiPoint","coordinates":[[1.1,2.2],[3.3,4.4]]}`
	MultiCoordinatesInterfaceTest(NewMultiPointTest, res, tt)
}

func TestLineStringInterface(t *testing.T) {
	res := `{"type":"LineString","coordinates":[[1.1,2.2],[3.3,4.4]]}`
	tt := test{t}
	MultiCoordinatesInterfaceTest(NewLineStringTest, res, tt)

}
func TestMultiLineStringInterface(t *testing.T) {
	res := (`{"type":"MultiLineString","coordinates":[[[1.1,2.2],[3.3,4.4]],[[1.1,2.2],[3.3,4.4]]]}`)
	tt := test{t}
	MultiLineCoordinatesInterfaceTest(NewMultiLineStringTest, res, tt)
}

func TestPolygonInterface(t *testing.T) {
	res := `{"type":"Polygon","coordinates":[[[1.1,2.2],[3.3,4.4]],[[1.1,2.2],[3.3,4.4]]]}`
	tt := test{t}
	MultiLineCoordinatesInterfaceTest(NewPolygonTest, res, tt)
}

func TestMultiPolygon(t *testing.T) {
	var (
		m   *MultiPolygon
		err error
		res string
	)
	tt := test{t}
	res = `{"type":"MultiPolygon","coordinates":[]}`
	m = NewMultiPolygon(nil)
	tt.AssertMarshal(m, res)
	m = NewMultiPolygon(nil)
	err = m.AddGeometry(nil)
	tt.AssertNeq(err, nil, "Should raise assertion error")
}

func TestGeometryCollection(t *testing.T) {
	var (
		m   *GeometryCollection
		err error
		res string
	)
	tt := test{t}
	ress := []string{`{"type":"GeometryCollection",`,
		`"geometries":[`,
		`{"type":"Point",`,
		`"coordinates":[100,0]`,
		`},`,
		`{"type":"LineString",`,
		`"coordinates":[[101,0.1],[102.2,1]]`,
		`}]}`}
	res = strings.Join(ress, "")
	p := NewPoint(Coordinate{100, 0})
	ls := NewLineString(Coordinates{{101, 0.1}, {102.2, 1}})
	m = NewGeometryCollection([]interface{}{p, ls})
	tt.AssertMarshal(m, res, "1")
	m = NewGeometryCollection(nil)
	m.AddGeometry(p)
	m.AddGeometry(ls)
	tt.AssertMarshal(m, res, "2")
	m = NewGeometryCollection(nil)
	err = m.AddGeometry(nil)
	tt.AssertNeq(err, nil, "Should raise assertion error")
}
