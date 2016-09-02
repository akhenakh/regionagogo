package regionagogo

import (
	"bufio"
	"io/ioutil"
	"os"
	"strings"
	"testing"

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

var geoJSONIsland = `{"type":"FeatureCollection","features":[{"type":"Feature","properties":{"stroke":"#555555","stroke-width":2,"stroke-opacity":1,"fill":"#555555","fill-opacity":0.5,"name":"Ile d'Orléans"},"geometry":{"type":"MultiPolygon","coordinates":[[[[-71.17218017578125,46.841407127005866],[-71.17218017578125,47.040182144806664],[-70.784912109375,47.040182144806664],[-70.784912109375,46.841407127005866],[-71.17218017578125,46.841407127005866]]]]}}]}`

// belle ile region
var cpoints = []CPoint{
	CPoint{47.33148834860839, -3.114654101105884},
	CPoint{47.355373440132155, -3.148793098023077},
	CPoint{47.35814036718415, -3.151600714901065},
	CPoint{47.37148672093542, -3.176503059268782},
	CPoint{47.3875186220867, -3.221506313465625},
	CPoint{47.389553126875285, -3.234120245852694},
	CPoint{47.395331122633195, -3.242990689069075},
	CPoint{47.39520905225595, -3.249623175669058},
	CPoint{47.388861395337074, -3.253651496319492},
	CPoint{47.38324616158104, -3.262074347852888},
	CPoint{47.36880117443462, -3.256174282617053},
	CPoint{47.352850652935274, -3.245350715163397},
	CPoint{47.33637116100027, -3.237619595168212},
	CPoint{47.32062409132874, -3.240956184280492},
	CPoint{47.312160549070086, -3.22345943860222},
	CPoint{47.30158112237086, -3.171701626528829},
	CPoint{47.29661692942773, -3.093617317185476},
	CPoint{47.301947333502596, -3.067005989334973},
	CPoint{47.32062409132874, -3.062814907581924},
	CPoint{47.32831452059861, -3.07274329256893},
	CPoint{47.32733795847997, -3.102691209531713},
	CPoint{47.33148834860839, -3.114654101105884},
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

	gs, err := NewGeoSearch(tmpfile)
	require.NoError(t, err)
	gs.Debug = true
	defer gs.Close()

	r := strings.NewReader(geoJSONIsland)

	err = gs.ImportGeoJSONFile(r, []string{"name"})
	require.NoError(t, err)

	region := gs.RegionByID(1)
	require.NotNil(t, region)

	err = gs.ImportGeoData()
	require.NoError(t, err)

	region = gs.StubbingQuery(47.01492366313195, -70.842592064976714)
	require.NotNil(t, region)
}

func BenchmarkCities(tb *testing.B) {
	tmpfile, clean := createTempDB(tb)
	defer clean()

	fi, err := os.Open("testdata/world_states_10m.geojson")
	defer fi.Close()
	require.NoError(tb, err)

	r := bufio.NewReader(fi)

	gs, err := NewGeoSearch(tmpfile)
	defer gs.Close()

	err = gs.ImportGeoJSONFile(r, []string{"iso_a2", "name"})
	require.NoError(tb, err)

	err = gs.ImportGeoData()
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

	fi, err := os.Open("testdata/paysdelaloire.geojson")
	require.NoError(t, err)

	r := bufio.NewReader(fi)

	gs, err := NewGeoSearch(tmpfile)
	defer gs.Close()

	err = gs.ImportGeoJSONFile(r, []string{"iso_a2", "name"})
	require.NoError(t, err)

	err = fi.Close()
	require.NoError(t, err)

	err = gs.ImportGeoData()
	require.NoError(t, err)
}

func TestCities(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	tmpfile, clean := createTempDB(t)
	defer clean()

	fi, err := os.Open("testdata/world_states_10m.geojson")
	defer fi.Close()
	require.NoError(t, err)

	r := bufio.NewReader(fi)

	gs, err := NewGeoSearch(tmpfile)
	defer gs.Close()

	err = gs.ImportGeoJSONFile(r, []string{"iso_a2", "name"})
	require.NoError(t, err)

	err = gs.ImportGeoData()
	require.NoError(t, err)

	for _, city := range cities {
		t.Log("testing for", city)
		region := gs.StubbingQuery(city.c[0], city.c[1])
		require.NotNil(t, region)
		require.Equal(t, city.code, region.Data["iso_a2"])
		require.Equal(t, city.name, region.Data["name"])
	}
}
