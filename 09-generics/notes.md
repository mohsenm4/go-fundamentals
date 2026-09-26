# Generics Notes

## 1. فقط یک type واقعی داری

```go
func Sum[T int](xs []T) T {
	// ...
}