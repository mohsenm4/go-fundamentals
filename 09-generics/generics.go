package generics

func Map[T, U any](xs []T, f func(T) U) []U {
	result := make([]U, len(xs))

	for i, x := range xs {
		result[i] = f(x)
	}

	return result
}

func Filter[T any](xs []T, keep func(T) bool) []T {
	result := make([]T, 0)

	for _, x := range xs {
		if keep(x) {
			result = append(result, x)
		}
	}

	return result
}

func Reduce[T, U any](xs []T, init U, f func(U, T) U) U {
	result := init

	for _, x := range xs {
		result = f(result, x)
	}

	return result
}

type Set[T comparable] struct {
	items map[T]struct{}
}

func (s *Set[T]) Add(x T) {
	if s.items == nil {
		s.items = make(map[T]struct{})
	}

	s.items[x] = struct{}{}
}

func (s *Set[T]) Has(x T) bool {
	_, ok := s.items[x]
	return ok
}

func (s *Set[T]) Len() int {
	return len(s.items)
}
