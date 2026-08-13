package linklist

type Node struct {
	Value int
	Next  *Node
}

type LinkedList struct {
	Head   *Node
	length int
}

func (ll *LinkedList) Push(value int) {
	newNode := &Node{Value: value}
	if ll.Head == nil {
		ll.Head = newNode
	} else {
		current := ll.Head
		for current.Next != nil {
			current = current.Next
		}
		current.Next = newNode
	}
	ll.length++
}

func (ll *LinkedList) Print() {
	current := ll.Head
	for current != nil {
		println(current.Value)
		current = current.Next
	}
}

func (ll *LinkedList) Pop() (int, bool) {
	if ll.Head == nil {
		return 0, false
	}
	value := ll.Head.Value
	ll.Head = ll.Head.Next
	ll.length--

	return value, true
}

func (ll *LinkedList) Length() int {
	return ll.length
}
