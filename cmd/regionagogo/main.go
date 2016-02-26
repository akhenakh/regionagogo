package main

import (
	"encoding/json"
	"log"
	"strconv"

	"net/http"

	"github.com/akhenakh/regionagogo"
)

type Server struct {
	*regionagogo.GeoSearch
}

// countryHandler takes a lat & lng query params and return a JSON
// with the country of the coordinate
func (s *Server) countryHandler(w http.ResponseWriter, r *http.Request) {
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

	region := s.Query(lat, lng)
	w.Header().Set("Content-Type", "application/json")

	if region == nil {
		js, _ := json.Marshal(regionagogo.Region{Code: "unknown", Name: "unknown"})
		w.Write(js)
		return
	}

	js, _ := json.Marshal(region)
	w.Write(js)
}

func main() {

	gs := regionagogo.NewGeoSearch()

	data, err := Asset("bindata/geodata")
	if err != nil {
		log.Fatal(err)
	}

	err = gs.ImportGeoData(data)
	if err != nil {
		log.Fatal(err)
	}
	s := &Server{GeoSearch: gs}

	http.HandleFunc("/country", s.countryHandler)
	http.ListenAndServe(":8082", nil)
}
