package geojson

import (
	"encoding/json"
	"testing"
)

func BenchmarkMarshalPoint(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Marshal(NewPoint(Coordinate{1.1, 2.2}))
	}
}

func BenchmarkUnmarshalPoint(b *testing.B) {
	b.StopTimer()
	var p Point
	res := `{"type":"Point","coordinates":[1.1,2.2]}`
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		json.Unmarshal([]byte(res), &p)
	}
}

func BenchmarkUnmarshalLineString(b *testing.B) {
	b.StopTimer()
	var ls LineString
	res := `{"type":"LineString","coordinates":[[1,2], [3.0, 4.1], [1,2], [3.0, 4.1], [1,2], [3.0, 4.1], [1,2], [3.0, 4.1], [1,2], [3.0, 4.1],
	[1,2], [3.0, 4.1], [1,2], [3.0, 4.1], [1,2], [3.0, 4.1], [1,2], [3.0, 4.1], [1,2], [3.0, 4.1], [1,2], [3.0, 4.1],
	[1,2], [3.0, 4.1], [1,2], [3.0, 4.1], [1,2], [3.0, 4.1], [1,2], [3.0, 4.1], [1,2], [3.0, 4.1], [1,2], [3.0, 4.1],
	[1,2], [3.0, 4.1], [1,2], [3.0, 4.1], [1,2], [3.0, 4.1], [1,2], [3.0, 4.1], [1,2], [3.0, 4.1], [1,2], [3.0, 4.1]]}`
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		json.Unmarshal([]byte(res), &ls)
	}
}

func BenchmarkUnmarshalFeatureAndMultiPoint1(b *testing.B) {
	b.StopTimer()
	res := `{"type":"Feature","geometry":{"type":"MultiPoint","coordinates":[[1.1,2],[2,3.0]]},"properties":null}`
	var uf Feature
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		json.Unmarshal([]byte(res), &uf)
	}

}

func BenchmarkUnmarshalFeatureAndMultiPoint2(b *testing.B) {
	b.StopTimer()
	res := `{"type":"Feature","geometry":{"type":"MultiPoint","coordinates":[[1.1,2],[2,3.0]]},"properties":null}`
	var uf Feature
	json.Unmarshal([]byte(res), &uf)
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		g, _ := uf.GetGeometry()
		_ = g.(*MultiPoint)
	}

}
func BenchmarkUnmarshalFeatureAndMultiPoint3(b *testing.B) {
	b.StopTimer()
	res := `{"type":"Feature","geometry":{"type":"MultiPoint","coordinates":[[1.1,2],[2,3.0]]},"properties":null}`
	var uf Feature
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		json.Unmarshal([]byte(res), &uf)
		g, _ := uf.GetGeometry()
		_ = g.(*MultiPoint)
	}

}
