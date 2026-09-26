package generics

import (
	"reflect"
	"strconv"
	"testing"
)

func TestMap(t *testing.T) {
	got := Map([]int{1, 2, 3}, strconv.Itoa)
	want := []string{"1", "2", "3"}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestFilter(t *testing.T) {
	got := Filter([]int{1, 2, 3, 4}, func(x int) bool {
		return x%2 == 0
	})
	want := []int{2, 4}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestReduce(t *testing.T) {
	got := Reduce([]int{1, 2, 3, 4}, 0, func(sum, x int) int {
		return sum + x
	})
	want := 10

	if got != want {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestSet(t *testing.T) {
	var s Set[int]

	s.Add(1)
	s.Add(1)

	if !s.Has(1) {
		t.Error("expected Set to contain 1")
	}

	if s.Len() != 1 {
		t.Errorf("got length %d, want 1", s.Len())
	}
}
