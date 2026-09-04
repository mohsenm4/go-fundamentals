package linklist

// i want to test the iterator implementation for the linked list
import (
	"testing"
)

func TestIterator_EmptyList(t *testing.T) {
	ll := &LinkedList{}
	it := ll.Iterator()

	if it.HasNext() {
		t.Errorf("Expected HasNext to be false for empty list, got true")
	}

	defer func() {
		if r := recover(); r == nil {
			t.Errorf("Expected panic when calling Next on empty iterator, but did not panic")
		}
	}()

	it.Next()
}

func TestIterator_Traversal(t *testing.T) {
	ll := &LinkedList{}
	values := []int{1, 2, 3, 4, 5}
	for _, v := range values {
		ll.Push(v)
	}

	it := ll.Iterator()
	index := 0

	for it.HasNext() {
		val := it.Next()
		if val != values[index] {
			t.Errorf("Expected value %d at index %d, got %d", values[index], index, val)
		}
		index++
	}

	if index != len(values) {
		t.Errorf("Expected to traverse %d elements, but traversed %d", len(values), index)
	}
}

func TestIterator_Independent(t *testing.T) {
	ll := &LinkedList{}
	values := []int{10, 20, 30}
	for _, v := range values {
		ll.Push(v)
	}

	it1 := ll.Iterator()
	it2 := ll.Iterator()

	if !it1.HasNext() || !it2.HasNext() {
		t.Errorf("Expected both iterators to have next elements")
	}

	val1 := it1.Next()
	val2 := it2.Next()

	if val1 != values[0] || val2 != values[0] {
		t.Errorf("Expected both iterators to return the first value %d, got %d and %d", values[0], val1, val2)
	}

	val1 = it1.Next()
	val2 = it2.Next()

	if val1 != values[1] || val2 != values[1] {
		t.Errorf("Expected both iterators to return the second value %d, got %d and %d", values[1], val1, val2)
	}
}
