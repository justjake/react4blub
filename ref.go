package main

type RefSetter[T any] interface {
	Set(val T)
}

type RefGetter[T any] interface {
	Get() T
}
