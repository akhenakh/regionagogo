package mobile

import (
	"encoding/json"

	"github.com/akhenakh/regionagogo"
	"github.com/akhenakh/regionagogo/db/boltdb"
)

type GeoDB struct {
	db regionagogo.GeoFenceDB
}

type Fence struct {
	Iso     string
	Name    string
	GeoJSON string
}

func NewGeoDB() *GeoDB {
	g := &GeoDB{}
	return g
}

func NewFence() *Fence {
	f := &Fence{}
	return f
}

func (gf *GeoDB) OpenDB(path string) error {
	opts := []boltdb.GeoFenceBoltDBOption{
		boltdb.WithCachedEntries(20),
		boltdb.WithDebug(false),
		boltdb.WithReadOnly(true),
	}
	db, err := boltdb.NewGeoFenceBoltDB(path, opts...)
	if err != nil {
		return err
	}
	gf.db = db
	return nil
}

func (gf *GeoDB) Close() error {
	return gf.db.Close()
}

func (gf *GeoDB) FenceByID(id int) *Fence {
	region := gf.db.FenceByID(uint64(id))
	if region == nil {
		return nil
	}

	js, _ := json.Marshal(region.ToGeoJSON())
	fm := &Fence{
		Iso:     region.Data["iso_a2"],
		Name:    region.Data["name"],
		GeoJSON: string(js),
	}

	return fm
}

func (gf *GeoDB) QueryHandler(lat, lng float64) *Fence {
	region := gf.db.StubbingQuery(lat, lng)
	if region == nil {
		return nil
	}

	js, _ := json.Marshal(region.ToGeoJSON())
	fm := &Fence{
		Iso:     region.Data["iso_a2"],
		Name:    region.Data["name"],
		GeoJSON: string(js),
	}

	return fm
}
