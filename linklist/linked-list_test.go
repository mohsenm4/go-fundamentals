package linklist

import "testing"

func TestPop(t *testing.T) {
	tests := []struct {
		name   string
		values []int
		want   int
		wantOk bool
	}{
		{name: "empty list", values: []int{}, want: 0, wantOk: false},
		{name: "single pop", values: []int{1}, want: 1, wantOk: true},
		{name: "pop from three-element list", values: []int{1, 2, 3}, want: 1, wantOk: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ll := &LinkedList{}
			for _, v := range tt.values {
				ll.Push(v)
			}
			got, gotOk := ll.Pop()
			if got != tt.want || gotOk != tt.wantOk {
				t.Errorf("Pop() = (%d, %v), want (%d, %v)", got, gotOk, tt.want, tt.wantOk)
			}
		})
	}
}

func TestPushAndLength(t *testing.T) {
	tests := []struct {
		name   string
		values []int
		want   int
	}{
		{name: "empty list", values: []int{}, want: 0},
		{name: "single push", values: []int{1}, want: 1},
		{name: "three pushes", values: []int{1, 2, 3}, want: 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ll := &LinkedList{}
			for _, v := range tt.values {
				ll.Push(v)
			}
			if got := ll.Length(); got != tt.want {
				t.Errorf("Length() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestPushPopSequence(t *testing.T) {
	ll := &LinkedList{}
	ll.Push(1)
	ll.Push(2)
	ll.Push(3)

	if got, ok := ll.Pop(); got != 1 || !ok {
		t.Errorf("Pop() = (%d, %v), want (1, true)", got, ok)
	}
	if got, ok := ll.Pop(); got != 2 || !ok {
		t.Errorf("Pop() = (%d, %v), want (2, true)", got, ok)
	}
	if got, ok := ll.Pop(); got != 3 || !ok {
		t.Errorf("Pop() = (%d, %v), want (3, true)", got, ok)
	}
	if got, ok := ll.Pop(); got != 0 || ok {
		t.Errorf("Pop() = (%d, %v), want (0, false)", got, ok)
	}
}
