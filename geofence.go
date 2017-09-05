package regionagogo

import (
	"github.com/akhenakh/regionagogo/geostore"
	"github.com/golang/geo/s2"
)

var (
	defaultCoverer = &s2.RegionCoverer{MinLevel: 1, MaxLevel: 24, MaxCells: 32}
)

// GeoFenceDB is the main interface to store and query your geo database
type GeoFenceDB interface {
	// returns a Fence by it's storage id
	FenceByID(loopID uint64) *Fence

	// returns the fence for the corresponding lat, lng coordinates
	StubbingQuery(lat, lng float64, opts ...QueryOptionsFunc) (Fences, error)

	// RectQuery perform rectangular query ur upper right bl bottom left
	RectQuery(urlat, urlng, bllat, bllng float64, opts ...QueryOptionsFunc) (Fences, error)

	// RadiusQuery is performing a radius query
	RadiusQuery(lat, lng, radius float64, opts ...QueryOptionsFunc) (Fences, error)

	// Store a Fence into the DB
	StoreFence(rs *geostore.FenceStorage, cover []uint64) error

	// Close the DB
	Close() error
}

type QueryOptionsFunc func(*QueryOptions)

// queryOptions used to pass options to DB queries
type QueryOptions struct {
	// Returns all fences when multiple fences match
	MultipleFences bool
}

// WithMultipleFences enable multi fences in responses
func WithMultipleFences(mf bool) QueryOptionsFunc {
	return func(o *QueryOptions) {
		o.MultipleFences = mf
	}
}
