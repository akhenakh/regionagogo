package boltdb

import (
	"bufio"
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/akhenakh/regionagogo"
	"github.com/akhenakh/regionagogo/geostore"
	"github.com/golang/geo/s2"
	"github.com/stretchr/testify/require"
)

var cities = []struct {
	c    []float64
	code string
	name string
}{
	{[]float64{47.339608, -3.164062}, "FR", "Bretagne"},         // Le Palais
	{[]float64{47.204059, -1.549072}, "FR", "Pays de la Loire"}, // Nantes
	{[]float64{48.857205, 2.345581}, "FR", "le-de-France"},      // Paris
	{[]float64{37.935757, -122.347748}, "US", "California"},     // Richmond, CA
	{[]float64{19.542915, -155.665857}, "US", "Hawaii"},
	{[]float64{42.088032, 8.876953}, "FR", "Corse"},
	{[]float64{-22.009467, 166.403046}, "NC", "Sud"}, // Noumea
	{[]float64{40.642094, 9.140625}, "IT", "Sardegna"},
	{[]float64{39.578967, 3.098145}, "ES", "Islas Baleares"}, // Palma
	{[]float64{18.13378, -66.63208}, "PR", "PRI-00 (Puerto Rico aggregation)"},
	{[]float64{16.087218, -61.66626}, "FR", "GLP-00 (Guadeloupe aggregation)"},
	{[]float64{46.418926, 43.769531}, "RU", "Rostov"},
	{[]float64{41.976689, -114.076538}, "US", "Nevada"}, // Nevada corner
	{[]float64{46.819651, -71.255951}, "CA", "Qubec"},   // Quebec city, source data destroyed accents
	{[]float64{-23.954352, -46.367455}, "BR", "So Paulo"},
	{[]float64{-23.84353, -45.341949}, "BR", "So Paulo"},
	{[]float64{41.059757, 45.012906}, "AZ", "Qazax"},
}

const (
	geoJSONIsland      = `{"type":"FeatureCollection","features":[{"type":"Feature","properties":{"stroke":"#555555","stroke-width":2,"stroke-opacity":1,"fill":"#555555","fill-opacity":0.5,"name":"Ile d'Orléans"},"geometry":{"type":"MultiPolygon","coordinates":[[[[-71.17218017578125,46.841407127005866],[-71.17218017578125,47.040182144806664],[-70.784912109375,47.040182144806664],[-70.784912109375,46.841407127005866],[-71.17218017578125,46.841407127005866]]]]}}]}`
	geoJSONoverlapping = `{"type":"FeatureCollection","features":[{"type":"Feature","properties":{"stroke":"#555555","stroke-width":2,"stroke-opacity":1,"fill":"#555555","fill-opacity":0.5,"name":"outter"},"geometry":{"type":"Polygon","coordinates":[[[2.253570556640625,48.80505453139158],[2.253570556640625,48.90128927649513],[2.429351806640625,48.90128927649513],[2.429351806640625,48.80505453139158],[2.253570556640625,48.80505453139158]]]}},{"type":"Feature","properties":{"stroke":"#555555","stroke-width":2,"stroke-opacity":1,"fill":"#555555","fill-opacity":0.5,"name":"inner"},"geometry":{"type":"Polygon","coordinates":[[[2.267303466796875,48.83353759505566],[2.267303466796875,48.87555444355432],[2.37030029296875,48.87555444355432],[2.37030029296875,48.83353759505566],[2.267303466796875,48.83353759505566]]]}},{"type":"Feature","properties":{"stroke":"#555555","stroke-width":2,"stroke-opacity":1,"fill":"#555555","fill-opacity":0.5,"name":"bigoutter"},"geometry":{"type":"Polygon","coordinates":[[[2.208251953125,48.78605682994539],[2.208251953125,48.9211457038064],[2.45819091796875,48.9211457038064],[2.45819091796875,48.78605682994539],[2.208251953125,48.78605682994539]]]}}]}`
)

// belle ile region
var cpoints = []geostore.CPoint{
	{Lat: 47.33148834860839, Lng: -3.114654101105884},
	{Lat: 47.355373440132155, Lng: -3.148793098023077},
	{Lat: 47.35814036718415, Lng: -3.151600714901065},
	{Lat: 47.37148672093542, Lng: -3.176503059268782},
	{Lat: 47.3875186220867, Lng: -3.221506313465625},
	{Lat: 47.389553126875285, Lng: -3.234120245852694},
	{Lat: 47.395331122633195, Lng: -3.242990689069075},
	{Lat: 47.39520905225595, Lng: -3.249623175669058},
	{Lat: 47.388861395337074, Lng: -3.253651496319492},
	{Lat: 47.38324616158104, Lng: -3.262074347852888},
	{Lat: 47.36880117443462, Lng: -3.256174282617053},
	{Lat: 47.352850652935274, Lng: -3.245350715163397},
	{Lat: 47.33637116100027, Lng: -3.237619595168212},
	{Lat: 47.32062409132874, Lng: -3.240956184280492},
	{Lat: 47.312160549070086, Lng: -3.22345943860222},
	{Lat: 47.30158112237086, Lng: -3.171701626528829},
	{Lat: 47.29661692942773, Lng: -3.093617317185476},
	{Lat: 47.301947333502596, Lng: -3.067005989334973},
	{Lat: 47.32062409132874, Lng: -3.062814907581924},
	{Lat: 47.32831452059861, Lng: -3.07274329256893},
	{Lat: 47.32733795847997, Lng: -3.102691209531713},
	{Lat: 47.33148834860839, Lng: -3.114654101105884},
}

func createTempDB(t testing.TB) (string, func()) {
	tmpfile, err := ioutil.TempFile("", "teststorage")
	require.NoError(t, err)
	return tmpfile.Name(), func() {
		err := os.Remove(tmpfile.Name())
		if err != nil {
			t.Error(err)
		}
	}
}

func TestStorage(t *testing.T) {
	tmpfile, clean := createTempDB(t)
	defer clean()

	opts := WithDebug(true)
	gs, err := NewGeoFenceBoltDB(tmpfile, opts)
	require.NoError(t, err)
	defer gs.Close()

	r := strings.NewReader(geoJSONIsland)

	err = regionagogo.ImportGeoJSONFile(gs, r, []string{"name"})
	require.NoError(t, err)

	region := gs.FenceByID(1)
	require.NotNil(t, region)

	region = gs.StubbingQuery(47.01492366313195, -70.842592064976714)
	require.NotNil(t, region)
}

func BenchmarkCities(tb *testing.B) {
	tmpfile, clean := createTempDB(tb)
	defer clean()

	fi, err := os.Open("../../testdata/world_states_10m.geojson")
	defer fi.Close()
	require.NoError(tb, err)

	r := bufio.NewReader(fi)

	gs, err := NewGeoFenceBoltDB(tmpfile)
	defer gs.Close()

	err = regionagogo.ImportGeoJSONFile(gs, r, []string{"iso_a2", "name"})
	require.NoError(tb, err)

	for i := 0; i < tb.N; i++ {
		for _, city := range cities {
			gs.StubbingQuery(city.c[0], city.c[1])
		}
	}
}

func TestCCW(t *testing.T) {
	tmpfile, clean := createTempDB(t)
	defer clean()

	fi, err := os.Open("../../testdata/paysdelaloire.geojson")
	require.NoError(t, err)

	r := bufio.NewReader(fi)

	gs, err := NewGeoFenceBoltDB(tmpfile)
	defer gs.Close()

	err = regionagogo.ImportGeoJSONFile(gs, r, []string{"iso_a2", "name"})
	require.NoError(t, err)

	err = fi.Close()
	require.NoError(t, err)
}

func TestCities(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	tmpfile, clean := createTempDB(t)
	defer clean()

	fi, err := os.Open("../../testdata/world_states_10m.geojson")
	defer fi.Close()
	require.NoError(t, err)

	r := bufio.NewReader(fi)

	gs, err := NewGeoFenceBoltDB(tmpfile)
	defer gs.Close()

	err = regionagogo.ImportGeoJSONFile(gs, r, []string{"iso_a2", "name"})
	require.NoError(t, err)

	for _, city := range cities {
		t.Log("testing for", city)
		region := gs.StubbingQuery(city.c[0], city.c[1])
		require.NotNil(t, region)
		require.Equal(t, city.code, region.Data["iso_a2"])
		require.Equal(t, city.name, region.Data["name"])
	}
}

func TestOverlappingRegion(t *testing.T) {
	tmpfile, clean := createTempDB(t)
	defer clean()

	opts := WithDebug(true)
	gs, err := NewGeoFenceBoltDB(tmpfile, opts)
	require.NoError(t, err)
	defer gs.Close()

	r := strings.NewReader(geoJSONoverlapping)

	err = regionagogo.ImportGeoJSONFile(gs, r, []string{"name"})
	require.NoError(t, err)

	// this point is inside both Polygons should return the smaller
	lat := 48.85206549830757
	lng := 2.3064422607421875
	p := s2.PointFromLatLng(s2.LatLngFromDegrees(lat, lng))

	region1 := gs.FenceByID(1)
	require.NotNil(t, region1)
	require.True(t, region1.Loop.ContainsPoint(p))

	region1 = gs.FenceByID(2)
	require.NotNil(t, region1)
	require.True(t, region1.Loop.ContainsPoint(p))

	region := gs.StubbingQuery(48.85206549830757, 2.3064422607421875)
	require.NotNil(t, region)
	require.Equal(t, "inner", region.Data["name"])
}
