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
	geoJSONbadcover    = `{"type":"FeatureCollection","crs":{"type":"name","properties":{"name":"urn:ogc:def:crs:OGC:1.3:CRS84"}},"features":[{"type":"Feature","properties":{"fid":4740,"ID_0":79,"ISO":"FRA","NAME_0":"France","ID_1":4,"NAME_1":"Île-de-France","ID_2":13,"NAME_2":"Hauts-de-Seine","ID_3":52,"NAME_3":"Nanterre","ID_4":524,"NAME_4":"Rueil-Malmaison","ID_5":4740,"NAME_5":"Rueil-Malmaison","CCN_5":null,"CCA_5":null,"TYPE_5":"Chef-lieu canton","ENGTYPE_5":"Commune","Shape_Length":0.17305815277123773,"Shape_Area":0.0017781421356630909},"geometry":{"type":"MultiPolygon","coordinates":[[[[2.197860956192073,48.854763031005973],[2.182619333267212,48.851566314697209],[2.159867525100708,48.847721099853629],[2.150411605834961,48.858501434326172],[2.153764247894401,48.864406585693359],[2.150387287140006,48.870868682861328],[2.158314466476384,48.880607604980582],[2.16934871673584,48.895812988281307],[2.207759618759155,48.874229431152344],[2.211281538009644,48.86854171752924],[2.200677394867057,48.86307525634777],[2.203563690185547,48.860630035400447],[2.197860956192073,48.854763031005973]]]]}},{"type":"Feature","properties":{"fid":5756,"ID_0":79,"ISO":"FRA","NAME_0":"France","ID_1":4,"NAME_1":"Île-de-France","ID_2":19,"NAME_2":"Yvelines","ID_3":89,"NAME_3":"Saint-Germain-en-Laye","ID_4":715,"NAME_4":"La Celle-Saint-Cloud","ID_5":5756,"NAME_5":"La Celle-Saint-Cloud","CCN_5":null,"CCA_5":null,"TYPE_5":"Chef-lieu canton","ENGTYPE_5":"Commune","Shape_Length":0.13498382692062408,"Shape_Area":0.0007196745059465904},"geometry":{"type":"MultiPolygon","coordinates":[[[[2.159867525100708,48.847721099853629],[2.150032281875667,48.846660614013672],[2.145872592926139,48.836902618408203],[2.120545625686759,48.836025238037166],[2.110636234283504,48.841087341308651],[2.11092042922985,48.849983215332031],[2.119793891906852,48.848270416259879],[2.122449874877987,48.850780487060661],[2.139863729476986,48.855907440185661],[2.144469022750854,48.861812591552848],[2.153764247894401,48.864406585693359],[2.150411605834961,48.858501434326172],[2.159867525100708,48.847721099853629]]]]}},{"type":"Feature","properties":{"fid":4722,"ID_0":79,"ISO":"FRA","NAME_0":"France","ID_1":4,"NAME_1":"Île-de-France","ID_2":13,"NAME_2":"Hauts-de-Seine","ID_3":51,"NAME_3":"Boulogne-Billancourt","ID_4":507,"NAME_4":"Chaville","ID_5":4722,"NAME_5":"Vaucresson","CCN_5":null,"CCA_5":null,"TYPE_5":"Commune simple","ENGTYPE_5":"Commune","Shape_Length":0.11045494594174084,"Shape_Area":0.00037342733185105235},"geometry":{"type":"MultiPolygon","coordinates":[[[[2.148475885391292,48.828491210937557],[2.145872592926139,48.836902618408203],[2.150032281875667,48.846660614013672],[2.159867525100708,48.847721099853629],[2.182619333267212,48.851566314697209],[2.179811716079769,48.845520019531364],[2.167349100112972,48.84089279174799],[2.166622877121029,48.837741851806697],[2.156419277191276,48.837657928466797],[2.151465654373169,48.821407318115348],[2.148475885391292,48.828491210937557]]]]}}]}`
	geoJSONbogusLoop   = `{"type":"FeatureCollection","crs":{"type":"name","properties":{"name":"urn:ogc:def:crs:OGC:1.3:CRS84"}},"features":[{"type":"Feature","properties":{"name":"Stuyvesant Town"},"geometry":{"type":"Polygon","coordinates":[[[-73.974378042082535,40.735081112182399],[-73.974377681392212,40.73508110966015],[-73.973959050297182,40.733421165538616],[-73.973943219249392,40.733403088863227],[-73.973907026951892,40.733349716608224],[-73.973866318655993,40.733310976086777],[-73.973844838548871,40.733265361113745],[-73.973865208024137,40.733246428039621],[-73.973857300365054,40.733216303638727],[-73.973841459626641,40.73321974036562],[-73.973849373812769,40.733232655106264],[-73.973831268466924,40.733245565109108],[-73.973809768465628,40.73324899937294],[-73.973783757834028,40.733227480705473],[-73.973802988474361,40.733211987539789],[-73.973783768671538,40.733199934826949],[-73.973764526494705,40.733212843387854],[-73.973727187077699,40.733219714350881],[-73.973689849825874,40.733218851106024],[-73.973676275760681,40.73320593388798],[-73.973692128598273,40.733174095747451],[-73.973737397623239,40.73314311937235],[-73.973818871718478,40.733099247654565],[-73.973872061740181,40.733094954806603],[-73.973868107762087,40.733065687195591],[-73.973658684424478,40.732244701586581],[-73.973514648369843,40.731680048528403],[-73.973430693696045,40.7313509277162],[-73.973413217308604,40.731282416428712],[-73.971963403012609,40.730000377217564],[-73.971955742642749,40.729988480733589],[-73.971479949072261,40.729249577698475],[-73.971427184435413,40.728485826387747],[-73.971555535149548,40.727703108598106],[-73.971569316470976,40.727693767198396],[-73.97157278760146,40.727645919767681],[-73.971589651499855,40.727640033693248],[-73.971618727069966,40.727650558185786],[-73.971651600755891,40.727643212534602],[-73.971685148665927,40.72740588663661],[-73.971536865927433,40.727392965100776],[-73.971520767896052,40.727386166239697],[-73.971512821812382,40.727374308581091],[-73.97149260058238,40.72737312335132],[-73.971490390928523,40.727350098110101],[-73.971629460631036,40.726760615951108],[-73.982552624994725,40.731374662598704],[-73.982022000000114,40.73201199999987],[-73.978527450968542,40.736854630838693],[-73.978526659827338,40.736854292908205],[-73.974907000000186,40.735312572323409],[-73.974648000000158,40.735081572323658],[-73.974377681392212,40.735079681983507],[-73.974378042082535,40.735081112182399]]]}}]}`
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

	err = regionagogo.ImportGeoJSONFile(gs, r, []string{"name"}, nil, nil)
	require.NoError(t, err)

	region := gs.FenceByID(1)
	require.NotNil(t, region)

	fences, err := gs.StubbingQuery(47.01492366313195, -70.842592064976714)
	require.NoError(t, err)
	require.Len(t, fences, 1)
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

	err = regionagogo.ImportGeoJSONFile(gs, r, []string{"iso_a2", "name"}, nil, nil)
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

	err = regionagogo.ImportGeoJSONFile(gs, r, []string{"iso_a2", "name"}, nil, nil)
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

	err = regionagogo.ImportGeoJSONFile(gs, r, []string{"iso_a2", "name"}, nil, nil)
	require.NoError(t, err)

	for _, city := range cities {
		t.Log("testing for", city)
		fences, err := gs.StubbingQuery(city.c[0], city.c[1])
		require.NoError(t, err)
		require.Len(t, fences, 1)
		require.Equal(t, city.code, fences[0].Data["iso_a2"])
		require.Equal(t, city.name, fences[0].Data["name"])
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

	err = regionagogo.ImportGeoJSONFile(gs, r, []string{"name"}, nil, nil)
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

	fences, err := gs.StubbingQuery(48.85206549830757, 2.3064422607421875)
	require.NoError(t, err)
	require.Len(t, fences, 1)
	require.Equal(t, "inner", fences[0].Data["name"])
}

func TestBadCover(t *testing.T) {
	tmpfile, clean := createTempDB(t)
	defer clean()

	opts := WithDebug(true)
	gs, err := NewGeoFenceBoltDB(tmpfile, opts)
	require.NoError(t, err)
	defer gs.Close()

	r := strings.NewReader(geoJSONbadcover)

	err = regionagogo.ImportGeoJSONFile(gs, r, []string{"NAME_5"}, nil, nil)
	require.NoError(t, err)

	// this point is inside geoJSONbadcover but was failing the test
	lat := 48.85071003048404
	lng := 2.1537494659423833
	p := s2.PointFromLatLng(s2.LatLngFromDegrees(lat, lng))

	region2 := gs.FenceByID(2)
	require.NotNil(t, region2)
	require.True(t, region2.Loop.ContainsPoint(p))

	fences, err := gs.StubbingQuery(lat, lng)
	require.NoError(t, err)
	require.Len(t, fences, 1)
	require.NotNil(t, fences[0].Data["NAME_5"], region2.Data["NAME_5"])
}

func TestBogusLoop(t *testing.T) {
	tmpfile, clean := createTempDB(t)
	defer clean()

	opts := WithDebug(true)
	gs, err := NewGeoFenceBoltDB(tmpfile, opts)
	require.NoError(t, err)
	defer gs.Close()

	r := strings.NewReader(geoJSONbogusLoop)

	err = regionagogo.ImportGeoJSONFile(gs, r, []string{"name"}, nil, nil)
	require.NoError(t, err)

	region1 := gs.FenceByID(0)
	require.Nil(t, region1)
}
