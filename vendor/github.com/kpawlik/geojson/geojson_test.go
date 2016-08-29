package geojson

import (
	"strings"
	"testing"
)

type test struct {
	*testing.T
}

func (t test) grabDescription(descr ...string) string {
	var res string
	if len(descr) > 0 {
		res = strings.Join(descr, ", ")
	}
	return res
}

func (t *test) AssertMarshal(o interface{}, res string, args ...string) {
	if s, err := Marshal(o); err != nil {
		t.Fatalf("%s: %s", err)
	} else {
		if s != res {
			descr := t.grabDescription(args...)
			t.Errorf("ErrorAssertMarshal(%s): %s != %s", descr, res, s)
		}
	}
}

func (t *test) AssertEq(o1, o2 interface{}, args ...string) {
	if o1 != o2 {
		descr := t.grabDescription(args...)
		t.Errorf("ErrorAssertEq(%s): %v != %v", descr, o1, o2)
	}
}

func (t *test) AssertNeq(o1, o2 interface{}, args ...string) {
	if o1 == o2 {
		descr := t.grabDescription(args...)
		t.Errorf("ErrorAssertNeq: %s", descr)
	}
}
func (t *test) AssertCoordinates(c1, c2 Coordinate, args ...string) {
	if c1[0] != c2[0] || c1[1] != c2[1] {
		descr := t.grabDescription(args...)
		t.Errorf("Coordinates not equal %v != %v. %s", c1, c2, descr)
	}
}
