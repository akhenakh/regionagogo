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
	var ff fieldFlag

	filename := flag.String("filename", "", "A geojson file")
	dbpath := flag.String("dbpath", "", "Database path")
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

	err = regionagogo.ImportGeoJSONFile(gs, r, ff.Fields)
	if err != nil {
		log.Fatal(err)
	}

}
