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

// Temp data valid only for a single render pass.
type fiberTemp struct {
	// Set by the renderer
	rendererTemp any

	// Identifies this render generation
	generation int

	// The nearest ancestor that is actually mounted in some kinda DOM If we have
	// a mounted node, it should be inserted into mountedParent's node by the end
	// of the render.
	mountedParent *fiber
	// Index in the nearest parent DOM node, which may not b `fiber.parent.mounted`
	startIndex int
	// Most fibers are 1 element wide. Fragments are N elements wide, so this
	// should be `N - 1` for fragments.
	indexOffsetWidth int
	// If true, fiber is known to be alive after this render
	alive bool
	// If true, fiber is known to need removal this render
	dead bool
	// If true, fiber is already inserted as intended. TODO: use for assertions
	inserted bool
}

// A fiber hosts an instance of a component instance across multiple renders. It
// stores component hook state and other information needed to re-render the
// component.
type fiber struct {
	// Coordination data.
	root         *root
	parent       *fiber
	children     map[string]*fiber
	deadChildren map[string]*fiber

	// TODO: separate attributes into "retained" between renders and "temporary"
	// for current render only data.
	temp fiberTemp

	// Rendering data.
	// Retained across re-renders.
	node     AnyNode // Component user resquested we render
	rendered AnyNode // Subtree of that component
	mounted  any     // associated renderer object
	dirty    bool    // If true, this fiber should re-render during next render

	// Hooks
	nextHook int
	hooks    []hook
}

type IFiber[Mounted any, Temp any] fiber

func (f *IFiber[Mounted, Temp]) SetTemp(v Temp) {
	f.temp.rendererTemp = v
}

func (f *IFiber[Mounted, Temp]) Temp() Temp {
	return f.temp.rendererTemp.(Temp)
}

func (f *IFiber[Mounted, Temp]) SetMounted(v Mounted) {
	f.mounted = v
}

func (f *IFiber[Mounted, Temp]) Mounted() Mounted {
	return f.mounted.(Mounted)
}

func (f *fiber) findChild(index int, childNode AnyNode) (childFiber *fiber, ok bool) {
	key := getKeyOrIndex(childNode, index)
	if childFiber, ok := f.children[key]; ok {
		if childFiber.node.component() != childNode.component() {
			// Component changed, eg div -> span
			return childFiber, false
		} else {
			return childFiber, true
		}
	}
	childFiber = &fiber{
		root:   f.root,
		parent: f,
	}
	if f.children == nil {
		f.children = make(map[string]*fiber)
	}
	f.children[key] = childFiber
	return childFiber, true
}

func (f *fiber) sweep() {
	if f.children == nil {
		return
	}

	for key, childFiber := range f.children {
		if !childFiber.temp.alive {
			debug.Printf("fiber.sweep(): remove unused child %T [%d]", childFiber.node.component(), childFiber.temp.startIndex)
			childFiber.unmount()
			delete(f.children, key)
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

type componentKindHandlers struct {
	Nil      func()
	Fragment func(FragmentProps)
	Text     func(TextProps)
	HTML     func(HtmlTag, HTMLProps)
	Other    func(AnyNode)
}

func (ops componentKindHandlers) Perform(node AnyNode) {
	comp := node.component()
	props := node.props()

	if ops.Nil != nil && comp == nil {
		ops.Nil()
		return
	}

	switch comp := comp.(type) {
	case fragmentComponent:
		if ops.Fragment != nil {
			ops.Fragment(props.(FragmentProps))
			return
		}
	case textComponent:
		if ops.Text != nil {
			ops.Text(props.(TextProps))
			return
		}
	case HtmlTag:
		if ops.HTML != nil {
			ops.HTML(comp, props.(HTMLProps))
			return
		}
	}

	ops.Other(node)
}

func isPrimitiveComponent(node AnyNode) bool {
	primitive := true
	componentKindHandlers{
		Nil:      func() {},
		Fragment: func(FragmentProps) {},
		Text:     func(TextProps) {},
		HTML:     func(HtmlTag, HTMLProps) {},
		Other: func(AnyNode) {
			primitive = false
		},
	}.Perform(node)
	return primitive
}

func getKeyOrIndex(node AnyNode, index int) string {
	key := node.key()
	if key != nil {
		return fmt.Sprintf("key:%s", *key)
	}
	return fmt.Sprintf("idx:%d", index)
}

type Renderer func(fiber *fiber, nextRendered AnyNode)

// TOOD:
// 0. validate?
//    Remove duplicate keys
func removeDuplicateKeys(children []AnyNode) {
	seen := make(map[string]bool)
	for i, child := range children {
		key := getKeyOrIndex(child, i)
		if seen[key] {
			fmt.Printf("Key used more than once: %s", key)
			if !child.clearKey() {
				panic(fmt.Errorf("Couldn't clear duplicate key: %s", key))
			}
		}
	}
}

// It seems like we need to phase our execution better:
// 1. Reconcile and mark fibers
//    - Pre-order depth first traversal:
//    - For each new node
//      - Find fiber for node
//        - Create missing fiber
//      - Mark fiber as alive
//      - Recurse on node
//        - Render each node -> rendered
//        - Recurse on rendered
//        - Assign DOM index in DOM parent (*Renderer specific*)
//      - On way back up:
//      - For each fiber
//        - If not alive, mark Dead
func ReconcileAndMark(parentFiber *fiber, rendered AnyNode) {
	fiber := parentFiber

	if !isPrimitiveComponent(parentFiber.node) {
		// We are rendering the contents of a custom component.
		// Those contents need their own fiber.
		childFiber, ok := parentFiber.findChild(0, rendered)
	}

	nextRendered := fiber.invokeRenderWithHooks()
	if isPrimitiveComponent(fiber.node) {
		renderer(fiber, nextRendered)
	} else {
		renderChildren(fiber, []AnyNode{nextRendered}, renderer)
	}
}

// 2. Sweep dead fibers
//    - Post-order depth first traversal: (on way back up)
//    - Sweep all dead fibers
//      - To sweep a fiber:
//        - Post order depth first traversal of the fiber
//          - If fiber has Use(Layout)Effect cleanups: run cleanup
//        - IFF fiber has no Dead ancestor, unmount it's top-level DOM nodes. (*Renderer specific*)
//          (If there's a parent Dead ancestor, then that ancestor's sweep should unmount a parent DOM node)
//          This will usually just be a single top-level DOM node,
//          but may be several if we're unmounting <Fragment><Thing /><Thing /><Thing /></Fragment>
// 3. Apply changes to DOM
//    - Pre-order depth first traversal: (on way down)
//      - Upsert DOM node for each fiber
//        - Create DOM node
//        - Mutate DOM node
//    - Post-order depth first traversal: (on way up)
//      - If node is a DOM node (or a top-level Fragment with no parent DOM node re-rendering)
//        - Insert/Re-order all DOM children
// 4. UseLayoutEffect (post-order DF)
//    - Run cleanup
//    - Run next effect
// 5. Paint (somehow)
// 6. UseEffect (post-order DF)
//    - Run cleanup
//    - Run next effect

func Sweep(ancestor *fiber)              {}
func FlushChanges(ancestor *fiber)       {}
func FlushLayoutEffects(ancestor *fiber) {}
func FlushPaint(ancestor *fiber)         {}
func FlushEffects(ancestor *fiber)       {}

func renderChildren(ancestor *fiber, children []AnyNode, renderer Renderer) {
	ancestorIndexOffset := 0
	if ancestor.mounted == nil {
		ancestorIndexOffset = ancestor.temp.startIndex
	}
	widthOffset := 0
	// TODO: this is basically our reconciler algo...
	for i, childNode := range children {
		childFiber, _ := ancestor.findChild(i, childNode)
		comp := childNode.component()
		prevNode := childFiber.node
		childFiber.temp.alive = true // Mark

		if memo, ok := comp.(MemoComponent); ok && memo.SameComponent(prevNode.component()) {
			childFiber.dirty = childFiber.dirty || memo.PropsEqual(prevNode.props(), childNode.props())
		} else {
			childFiber.dirty = true
		}

		if childFiber.dirty {
			childFiber.node = childNode
		}

		childFiber.temp.startIndex = i + widthOffset + ancestorIndexOffset
		render(childFiber, renderer)
		widthOffset += childFiber.temp.indexOffsetWidth
	}

	// Unmount fibers no longer retained after this render
	ancestor.sweep() // Sweep
	if ancestor.mounted == nil {
		ancestor.temp.indexOffsetWidth = widthOffset
		debug.Printf("renderChildren: %T with indexOffsetWidth %d", ancestor.node.component(), ancestorIndexOffset)
	}
}

func render(fiber *fiber, renderer Renderer) {
	if fiber.mounted != nil && !fiber.dirty {
		return
	}
	// This I'm always confused by.
	// If our custom component returned nil, what do?
	// Do we need to spawn another fiber depending on the child type?
	//
	// I think the algo is:
	// - if *fiber.node* is custom type, spawn child fiber for nextRendered w/ key, index 0
	// - if *fiber.node* is basic type, render nextRendered directly
	// TODO: change stuff around
	// TODO: support custom components
	// prevRendered := fiber.rendered

	nextRendered := fiber.invokeRenderWithHooks()
	if isPrimitiveComponent(fiber.node) {
		renderer(fiber, nextRendered)
	} else {
		renderChildren(fiber, []AnyNode{nextRendered}, renderer)
	}
}
