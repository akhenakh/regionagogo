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
	var forceFields fieldFlag
	flag.Var(&forceFields, "forceFields", "List of fields to enforce as a property, eg level=country")

	// rename a field on the fly
	var renameFields fieldFlag
	flag.Var(&renameFields, "renameFields", "List of fields to be renamed on the fly as a property, eg NAME_EN=name\n\tnote NAME_EN needs to be in importFields even if it will be renamed")

	filename := flag.String("filename", "", "A geojson file")
	dbpath := flag.String("dbpath", "", "Database path")
	debug := flag.Bool("debug", false, "Enable debug")
	featureImport := flag.Bool("featureImport", false, "the GeoJSON is a feature not a featureCollection")

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

	if len(forceFields.Fields) > 0 {
		// check the syntax
		for _, field := range forceFields.Fields {
			if len(strings.Split(field, "=")) != 2 {
				fmt.Println("Invalid forceField", field)
				flag.PrintDefaults()
				os.Exit(2)
			}
			split := strings.Split(field, "=")
			forceFieldsMap[split[0]] = split[1]
		}
	}

	renameFieldsMap := make(map[string]string)

	if len(renameFields.Fields) > 0 {
		// check the syntax
		for _, field := range renameFields.Fields {
			if len(strings.Split(field, "=")) != 2 {
				fmt.Println("Invalid forceField", field)
				flag.PrintDefaults()
				os.Exit(2)
			}
			split := strings.Split(field, "=")
			renameFieldsMap[split[0]] = split[1]
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

	i := regionagogo.NewGeoJSONImport(gs, r, importFields.Fields, forceFieldsMap, renameFieldsMap)
	i.FeatureImport = *featureImport
	if err := i.Start(); err != nil {
		log.Fatal(err)
	}

}
