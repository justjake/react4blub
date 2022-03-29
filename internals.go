package main

import (
	"fmt"
	"log"
	"os"
)

var debug = log.New(os.Stderr, "react", log.LstdFlags|log.Lshortfile)

var globalState struct {
	currentFiber *fiber
}

// TODO: actually be a fiber
type fiber struct {
	root   *root
	parent *fiber

	// Index in the nearest parent DOM node, which may not b `fiber.parent.mounted`
	startIndex int
	// Most fibers are 1 element wide. Fragments are N elements wide, so this
	// should be `N - 1` for fragments.
	indexOffsetWidth int

	node     AnyNode // Component user resquested we render
	rendered AnyNode // Subtree of that component
	mounted  any
	dirty    bool
	retain   bool

	children map[string]*fiber

	nextHook int
	hooks    []hook
}

func (f *fiber) findChild(index int, childNode AnyNode) *fiber {
	key := getKeyOrIndex(childNode, index)
	if childFiber, ok := f.children[key]; ok {
		if childFiber.node.component() != childNode.component() {
			// Component changed, eg div -> span
			childFiber.unmount()
			delete(f.children, key)
		} else {
			return childFiber
		}
	}
	childFiber := &fiber{
		root:   f.root,
		parent: f,
	}
	if f.children == nil {
		f.children = make(map[string]*fiber)
	}
	f.children[key] = childFiber
	return childFiber
}

func (f *fiber) sweep() {
	if f.children == nil {
		return
	}

	for key, childFiber := range f.children {
		if !childFiber.retain {
			debug.Printf("fiber.sweep(): remove unused child %T [%d]", childFiber.node.component(), childFiber.startIndex)
			childFiber.unmount()
			delete(f.children, key)
		} else {
			childFiber.retain = false
		}
	}
}

// TODO: schedule for re-render
func (f *fiber) shouldRerender() {
	if !f.dirty {
		f.dirty = true
		// TODO: schedule
	}
}

// TODO
func (f *fiber) unmount() {
	for _, hook := range f.hooks {
		hook.unmount()
	}
	f.mounted = nil
}

func (f *fiber) invokeRenderWithHooks() AnyNode {
	globalState.currentFiber = f
	defer func() { globalState.currentFiber = nil }()
	f.nextHook = 0

	result := f.node.invokeRender()
	if f.nextHook < len(f.hooks) {
		panic(fmt.Errorf("Re-render invoked %d hooks out of %d hooks", f.nextHook-1, len(f.hooks)))
	}

	return result
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
