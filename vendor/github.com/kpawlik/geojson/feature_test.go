package geojson

import (
	"encoding/json"
	"testing"
)

func TestFeature(t *testing.T) {
	tt := test{t}
	res := `{"type":"Feature","geometry":{"type":"Point","coordinates":[1,2]},"properties":null}`
	p := NewPoint(Coordinate{1.0, 2.0})
	f := NewFeature(p, nil, nil)
	tt.AssertMarshal(f, res)
}

func TestFeatureWithId(t *testing.T) {
	tt := test{t}
	res := `{"type":"Feature","id":12,"geometry":{"type":"Point","coordinates":[1,2]},"properties":null}`
	p := NewPoint(Coordinate{1.0, 2.0})
	f := NewFeature(p, nil, 12)
	tt.AssertMarshal(f, res)
}

func TestFeatureWithProps(t *testing.T) {
	tt := test{t}
	res := `{"type":"Feature","id":12,"geometry":{"type":"Point","coordinates":[1,2]},"properties":{"foo":"bar","one":1}}`
	props := map[string]interface{}{"foo": "bar", "one": 1}
	p := NewPoint(Coordinate{1.0, 2.0})
	f := NewFeature(p, props, 12)
	tt.AssertMarshal(f, res)
}

func TestFeatureWithCRS(t *testing.T) {
	tt := test{t}
	res := `{"type":"Feature","id":12,"geometry":{"type":"Point","coordinates":[1,2]},"properties":null,"crs":{"type":"name","properties":{"name":"osm.some.com"}}}`
	crs := NewNamedCRS("osm.some.com")
	p := NewPoint(Coordinate{1.0, 2.0})
	f := NewFeature(p, nil, 12)
	f.Crs = crs
	tt.AssertMarshal(f, res)
}

func TestCRS(t *testing.T) {
	var (
		res string
		crs *CRS
	)
	tt := test{t}
	res = `{"type":"name","properties":{"name":"osm.some.com"}}`
	crs = NewNamedCRS("osm.some.com")
	tt.AssertMarshal(crs, res)
	res = `{"type":"link","properties":{"href":"file.crs","type":"osm"}}`
	crs = NewLinkedCRS("file.crs", "osm")
	tt.AssertMarshal(crs, res)
}

func TestUnmarshalFeatureAndPoint(t *testing.T) {
	var (
		uf  Feature
		err error
		p   *Point
		ok  bool
	)
	tt := test{t}
	res := `{"type":"Feature","geometry":{"type":"Point","coordinates":[1,2]},"properties":null}`
	err = json.Unmarshal([]byte(res), &uf)
	tt.AssertEq(err, nil, "No error expected 1")
	g, err := uf.GetGeometry()
	tt.AssertEq(err, nil, "No error expected 2")
	p, ok = g.(*Point)
	tt.AssertEq(ok, true, "No error expected 3")
	tt.AssertEq(p.Type, "Point", "Wrong geometry type")
}

func TestUnmarshalFeatureAndLineString(t *testing.T) {
	var (
		uf  Feature
		err error
		ls  *LineString
		ok  bool
	)
	tt := test{t}
	res := `{"type":"Feature","geometry":{"type":"LineString","coordinates":[[1,2],[2,3.0]]},"properties":null}`
	err = json.Unmarshal([]byte(res), &uf)
	tt.AssertEq(err, nil, "No error expected 1")
	g, err := uf.GetGeometry()
	tt.AssertEq(err, nil, "No error expected 2")
	ls, ok = g.(*LineString)
	tt.AssertEq(ok, true, "No error expected 3")
	tt.AssertEq(ls.Type, "LineString", "Wrong geometry type")
	tt.AssertCoordinates(ls.Coordinates[0], Coordinate{1, 2}, "Coordinate test 1")
	tt.AssertCoordinates(ls.Coordinates[1], Coordinate{2, 3}, "Coordinate test 2")
}

func TestUnmarshalFeatureAndMultiPoint(t *testing.T) {
	var (
		uf  Feature
		err error
		mp  *MultiPoint
		ok  bool
	)
	tt := test{t}
	res := `{"type":"Feature","geometry":{"type":"MultiPoint","coordinates":[[1.1,2],[2,3.0]]},"properties":null}`
	err = json.Unmarshal([]byte(res), &uf)
	tt.AssertEq(err, nil, "No error expected 1")
	g, err := uf.GetGeometry()
	tt.AssertEq(err, nil, "No error expected 2")
	mp, ok = g.(*MultiPoint)
	tt.AssertEq(ok, true, "No error expected 3")
	tt.AssertEq(mp.Type, "MultiPoint", "Wrong geometry type")
	tt.AssertCoordinates(mp.Coordinates[0], Coordinate{1.1, 2}, "Coordinate test 1")
	tt.AssertCoordinates(mp.Coordinates[1], Coordinate{2, 3}, "Coordinate test 2")
}

func TestUnmarshalFeatureAndMultiLineString(t *testing.T) {
	var (
		uf  Feature
		err error
		mls *MultiLineString
		ok  bool
	)
	tt := test{t}
	res := `{"type":"Feature","geometry":{"type":"MultiLineString","coordinates":[[[1.1,2],[2,3.0]], [[2.1,3],[3,4.0]]]},"properties":null}`
	err = json.Unmarshal([]byte(res), &uf)
	tt.AssertEq(err, nil, "No error expected 1")
	g, err := uf.GetGeometry()
	_ = g
	tt.AssertEq(err, nil, "No error expected 2")
	mls, ok = g.(*MultiLineString)
	_ = mls
	_ = ok
	tt.AssertEq(ok, true, "No error expected 3")
	tt.AssertEq(mls.Type, "MultiLineString", "Wrong geometry type")
	tt.AssertCoordinates(mls.Coordinates[0][0], Coordinate{1.1, 2}, "Coordinate test 1")
	tt.AssertCoordinates(mls.Coordinates[0][1], Coordinate{2, 3}, "Coordinate test 2")
}

func TestParseCoordinate(t *testing.T) {
	var (
		icoord interface{}
		coord  Coordinate
		err    error
	)
	tt := test{t}
	icoord = []interface{}{1, 2.0}
	coord, _ = parseCoordinate(icoord)
	_ = coord
	tt.AssertEq(coord[0], Coord(1), "Parse coordinate test 1")
	tt.AssertEq(coord[1], Coord(2.0), "Parse coordinate test 2")
	icoord = []interface{}{1, "Ala ma kota"}
	_, err = parseCoordinate(icoord)
	tt.AssertNeq(err, nil, "Error expected")
}

func TestParseCoordinates(t *testing.T) {
	var (
		icoords interface{}
		err     error
	)
	icoords = []interface{}{[]interface{}{1, 0}, []interface{}{1.1, 2.0}}
	tt := test{t}
	if coords, err := parseCoordinates(icoords); err != nil {
		t.Error(err)
	} else {
		tt.AssertEq(coords[0], Coordinate{1, 0}, "Parse first coordinate error")
		tt.AssertEq(coords[1], Coordinate{1.1, 2}, "Parse second coordinate error")
	}
	_, err = parseCoordinates([][]interface{}{{"1", 1}, {"asd", 0}})
	tt.AssertNeq(err, nil, "Error expected")
}

func TestParseMultiline(t *testing.T) {
	var (
		icoords interface{}
		err     error
	)
	icoords = []interface{}{[]interface{}{[]interface{}{1, 1}, []interface{}{2, 2}},
		[]interface{}{[]interface{}{3, 3}, []interface{}{4, 4}}}
	tt := test{t}
	if ml, err := parseMultiLine(icoords); err != nil {
		t.Error(err)
	} else {
		tt.AssertEq(ml[0][0], Coordinate{1, 1}, "Parse first coordinate error")
		tt.AssertEq(ml[1][1], Coordinate{4, 4}, "Parse last coordinate error")
	}
	_, err = parseCoordinates([]interface{}{1, 2})
	tt.AssertNeq(err, nil, "Error expected")
}
