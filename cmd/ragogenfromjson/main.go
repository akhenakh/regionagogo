package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/akhenakh/regionagogo"
	"github.com/akhenakh/regionagogo/db/boltdb"
)

// fieldFlag reusable parse Value to create import command
type fieldFlag struct {
	Fields []string
}

func (ff *fieldFlag) String() string {
	return fmt.Sprint(ff.Fields)
}

func (ff *fieldFlag) Set(value string) error {
	if len(ff.Fields) > 0 {
		return fmt.Errorf("The field flag is already set")
	}

	ff.Fields = strings.Split(value, ",")
	return nil
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	// fields name we want to import as metadata
	var importFields fieldFlag
	flag.Var(&importFields, "importFields", "List of fields to fetch inside GeoJSON properties")

	// add a property to each entries imported
	var forceField fieldFlag
	flag.Var(&forceField, "forceFields", "List of fields to enforce as a property, eg level=country")

	filename := flag.String("filename", "", "A geojson file")
	dbpath := flag.String("dbpath", "", "Database path")
	debug := flag.Bool("debug", false, "Enable debug")

	flag.Parse()

	if len(*filename) == 0 {
		flag.PrintDefaults()
		os.Exit(2)
	}
	if len(importFields.Fields) < 1 {
		flag.PrintDefaults()
		os.Exit(2)
	}

	forceFieldsMap := make(map[string]string)

	if len(forceField.Fields) > 0 {
		// check the syntax
		for _, field := range forceField.Fields {
			if len(strings.Split(field, "=")) != 2 {
				fmt.Println("Invalid forceField", field)
				flag.PrintDefaults()
				os.Exit(2)
			}
			split := strings.Split(field, "=")
			forceFieldsMap[split[0]] = split[1]
		}
	}

	opts := boltdb.WithDebug(*debug)

	gs, err := boltdb.NewGeoFenceBoltDB(*dbpath, opts)
	if err != nil {
		log.Fatal(err)
	}

	fi, err := os.Open(*filename)
	defer fi.Close()
	if err != nil {
		log.Fatal(err)
	}
	r := bufio.NewReader(fi)

	err = regionagogo.ImportGeoJSONFile(gs, r, importFields.Fields, forceFieldsMap)
	if err != nil {
		log.Fatal(err)
	}

}
