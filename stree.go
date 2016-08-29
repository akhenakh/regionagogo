package regionagogo

import (
	"github.com/Workiva/go-datastructures/augmentedtree"
	"github.com/golang/geo/s2"
)

// S2Interval is a CellID interval conforms augmentedtree Interval interface
type S2Interval struct {
	s2.CellID
	LoopIDs []uint64
}

// LowAtDimension returns an integer representing the lower bound
// at the requested dimension.
func (s *S2Interval) LowAtDimension(d uint64) int64 { return int64(s.CellID.RangeMin()) }

// HighAtDimension returns an integer representing the higher bound
// at the requested dimension.
func (s *S2Interval) HighAtDimension(d uint64) int64 { return int64(s.CellID.RangeMax()) }

// OverlapsAtDimension should return a bool indicating if the provided
// interval overlaps this interval at the dimension requested.
func (s *S2Interval) OverlapsAtDimension(iv augmentedtree.Interval, d uint64) bool {
	return s.HighAtDimension(d) > iv.LowAtDimension(d) &&
		s.LowAtDimension(d) < iv.HighAtDimension(d)
}

// ID should be a unique ID representing this interval.  This
// is used to identify which interval to delete from the tree if
// there are duplicates.
func (s *S2Interval) ID() uint64 { return uint64(s.CellID) }
