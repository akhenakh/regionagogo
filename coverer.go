package regionagogo

import "github.com/golang/geo/s2"

// LoopRegion Making s2.Loop implement s2.Region
type LoopRegion struct {
	*s2.Loop
}

// LoopRegionFromPoints creates a LoopRegion from a list of s2.Point
func LoopRegionFromPoints(points []s2.Point) *LoopRegion {
	loop := s2.LoopFromPoints(points)
	return &LoopRegion{loop}
}

// CapBound returns the cap that contains this loop
func (l *LoopRegion) CapBound() s2.Cap {
	return l.Loop.CapBound()
}

// ContainsCell checks whether the cell is completely enclosed by this loop.
// Does not count for loop interior and uses raycasting.
func (l *LoopRegion) ContainsCell(c s2.Cell) bool {
	for i := 0; i < 4; i++ {
		v := c.Vertex(i)
		if !l.ContainsPoint(v) {
			return false
		}
	}
	return true
}

// IntersectsCell checks if any edge of the cell intersects the loop or if the cell is contained.
// Does not count for loop interior and uses raycasting.
func (l *LoopRegion) IntersectsCell(c s2.Cell) bool {
	// if any of the cell's vertices is contained by the loop
	// they intersect
	for i := 0; i < 4; i++ {
		v := c.Vertex(i)
		if l.ContainsPoint(v) {
			return true
		}
	}
	// missing case from the above implementation
	// where the loop is fully contained by the cell
	for _, v := range l.Vertices() {
		if c.ContainsPoint(v) {
			return true
		}
	}

	return false
}
