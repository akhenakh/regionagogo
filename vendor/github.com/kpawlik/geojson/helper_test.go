package geojson

////In case when coordType 1 should be parsed as 1.0
//func (t coordType) MarshalJSON() (json []byte, err error){
//	s := strconv.FormatFloat(float64(t), 'f', -1, 32)
//	if !strings.ContainsRune(s, '.'){
//		s = fmt.Sprintf("%s.0", s)
//	}
//	json = []byte(s)
//	return 
//}

type TestInterfaceMultiCoord interface {
	Geometry
	AddCoordinates(c ...Coordinate)
}

type TestInterfaceMultiLineCoords interface {
	Geometry
	AddCoordinates(c ...Coordinates)
}

//Factory function to create new object 
func NewMultiPointTest(coordinates Coordinates) TestInterfaceMultiCoord {
	return &MultiPoint{Type: "MultiPoint", Coordinates: coordinates}
}

//Factory function to create new object with points
func NewLineStringTest(coordinates Coordinates) TestInterfaceMultiCoord {
	return &LineString{Type: "LineString", Coordinates: coordinates}
}

func NewPolygonTest(coordinates []Coordinates) TestInterfaceMultiLineCoords {
	return &Polygon{Type: "Polygon", Coordinates: coordinates}
}

func NewMultiLineStringTest(coordinates []Coordinates) TestInterfaceMultiLineCoords {
	return &MultiLineString{Type: "MultiLineString", Coordinates: coordinates}
}
