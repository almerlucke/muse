package utils

type Factory[T any] interface {
	New() T
}
