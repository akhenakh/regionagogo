package main

import (
	"flag"
	"log"

	"github.com/akhenakh/regionagogo"
)

func main() {
	filename := flag.String("filename", "", "A geojson file")
	debug := flag.Bool("debug", false, "Enable debug")
	flag.Parse()

	if len(*filename) == 0 {
		flag.PrintDefaults()
		return
	}

	err := regionagogo.ImportGeoJSONFile(*filename, *debug)
	if err != nil {
		log.Fatal(err)
	}
}
