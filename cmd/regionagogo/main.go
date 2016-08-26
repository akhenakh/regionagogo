package main

import (
	"bufio"
	"encoding/json"
	"log"
	"os"
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

	data := s.Query(lat, lng)
	w.Header().Set("Content-Type", "application/json")

	if len(data) == 0 {
		js, _ := json.Marshal(map[string]string{"name": "unknown"})
		w.Write(js)
		return
	}

	js, _ := json.Marshal(data)
	w.Write(js)
}

func main() {

	gs := regionagogo.NewGeoSearch()

	fi, err := os.Open("geodata")
	defer fi.Close()
	if err != nil {
		log.Fatal(err)
	}

	r := bufio.NewReader(fi)

	err = gs.ImportGeoData(r)
	if err != nil {
		log.Fatal(err)
	}
	s := &Server{GeoSearch: gs}

	http.HandleFunc("/country", s.countryHandler)
	http.ListenAndServe(":8082", nil)
}
