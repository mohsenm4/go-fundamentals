package linklist

type Iterator interface {
	Next() int
	HasNext() bool
}

type LinkedListIterator struct {
	current *Node
}

func (it *LinkedListIterator) Next() int {
	if it.current == nil {
		panic("Next called on exhausted iterator")
	}
	value := it.current.Value
	it.current = it.current.Next
	return value
}

func (it *LinkedListIterator) HasNext() bool {
	return it.current != nil
}

func (ll *LinkedList) Iterator() Iterator {
	return &LinkedListIterator{current: ll.Head}
}
