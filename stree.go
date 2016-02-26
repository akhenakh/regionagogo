package regionagogo

import (
	"github.com/Workiva/go-datastructures/augmentedtree"
	"github.com/golang/geo/s2"
)

type S2Interval struct {
	s2.CellID
	LoopIDs []int
}

func (s *S2Interval) LowAtDimension(d uint64) int64  { return int64(s.CellID.RangeMin()) }
func (s *S2Interval) HighAtDimension(d uint64) int64 { return int64(s.CellID.RangeMax()) }

func (s *S2Interval) OverlapsAtDimension(iv augmentedtree.Interval, d uint64) bool {
	return s.HighAtDimension(d) > iv.LowAtDimension(d) &&
		s.LowAtDimension(d) < iv.HighAtDimension(d)
}

func (s *S2Interval) ID() uint64 { return uint64(s.CellID) }
