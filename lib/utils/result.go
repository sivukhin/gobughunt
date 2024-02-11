package utils

type Result[T any] struct {
	Value T
	Err   error
}

func NewResult[T any](value T, err error) Result[T] { return Result[T]{Value: value, Err: err} }
