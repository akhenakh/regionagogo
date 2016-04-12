package main

import (
	"flag"
	"log"
	"os"

	"github.com/akhenakh/regionagogo"
)

func main() {
	var ff regionagogo.FieldFlag

	filename := flag.String("filename", "", "A geojson file")
	debug := flag.Bool("debug", false, "Enable debug")
	flag.Var(&ff, "fields", "List of fields to fetch inside GeoJSON properties")
	flag.Parse()

	if len(*filename) == 0 {
		flag.PrintDefaults()
		os.Exit(2)
	}
	if len(ff.Fields) < 1 {
		flag.PrintDefaults()
		os.Exit(2)
	}

	err := regionagogo.ImportGeoJSONFile(*filename, *debug, ff.Fields)
	if err != nil {
		log.Fatal(err)
	}
}
