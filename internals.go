package main

type fiber struct {
	children []fiber
	mounted  *any
	nextHook int
	hooks    []hook
}
