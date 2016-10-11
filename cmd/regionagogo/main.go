package main

import (
	"encoding/json"
	"flag"
	"log"
	"net/http"
	"strconv"

	"github.com/akhenakh/regionagogo"
	"github.com/akhenakh/regionagogo/db/boltdb"
)

type server struct {
	regionagogo.GeoFenceDB
}

// queryHandler takes a lat & lng query params and return a JSON
// with the country of the coordinate
func (s *server) queryHandler(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	slat := query.Get("lat")
	lat, err := strconv.ParseFloat(slat, 64)
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	slng := query.Get("lng")
	lng, err := strconv.ParseFloat(slng, 64)
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}

	fences, err := s.StubbingQuery(lat, lng)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	w.Header().Set("Content-Type", "application/json")

	if len(fences) < 1 {
		js, _ := json.Marshal(map[string]string{"name": "unknown"})
		w.Write(js)
		return
	}

	js, _ := json.Marshal(fences[0].Data)
	w.Write(js)
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	dbpath := flag.String("dbpath", "", "Database path")
	debug := flag.Bool("debug", false, "Enable debug")
	cachedEntries := flag.Uint("cachedEntries", 0, "Region Cache size, 0 for disabled")

	flag.Parse()
	opts := []boltdb.GeoFenceBoltDBOption{
		boltdb.WithCachedEntries(*cachedEntries),
		boltdb.WithDebug(*debug),
	}
	gs, err := boltdb.NewGeoFenceBoltDB(*dbpath, opts...)
	if err != nil {
		log.Fatal(err)
	}

	s := &server{GeoFenceDB: gs}
	http.HandleFunc("/query", s.queryHandler)
	log.Println(http.ListenAndServe(":8082", nil))
}
