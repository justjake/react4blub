package main

var globalState struct {
	currentFiber *fiber
}

type fiber struct {
	root   *root
	parent *fiber

	node     AnyNode // Component user resquested we render
	rendered AnyNode // Subtree of that component
	children []fiber // Fibers for each child node that is a component?
	mounted  *any
	dirty    bool

	nextHook int
	hooks    []hook
}

func (f *fiber) updateRendered() AnyNode {
	globalState.currentFiber = f
	f.nextHook = 0

	f.rendered = f.node.invokeRender()

	globalState.currentFiber = nil
	return f.rendered
}

type root struct {
	fiber *fiber
	host  any
}

func newRoot(host any, node AnyNode) *root {
	root := root{
		host: host,
	}
	rootFiber := fiber{
		root: &root,
		node: node,
	}
	root.fiber = &rootFiber
	return &root
}
