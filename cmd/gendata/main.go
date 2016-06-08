package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/akhenakh/regionagogo"
)

// FieldFlag reusable parse Value to create import command
type FieldFlag struct {
	Fields []string
}

func (ff *FieldFlag) String() string {
	return fmt.Sprint(ff.Fields)
}

func (ff *FieldFlag) Set(value string) error {
	if len(ff.Fields) > 0 {
		return fmt.Errorf("The field flag is already set")
	}

	ff.Fields = strings.Split(value, ",")
	return nil
}

func main() {
	var ff FieldFlag

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
