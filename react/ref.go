package react

type Ref[T any] interface {
	Set(val T)
}

type RefStruct[T any] struct {
	Current T
}
