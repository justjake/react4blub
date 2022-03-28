package main

import (
	"fmt"
	"strings"
)

func main() {
	var foo = Counter.Node(CounterProps{initial: Some(5)})
	actual := foo.(*Node[CounterProps])
	fmt.Printf("v:%v\nt:%t\nT:%T\n", actual, actual, actual)
	fmt.Println(RenderToString(foo))
}

type renderOps struct {
	Nil      func()
	Fragment func(FragmentProps)
	Text     func(TextProps)
	HTML     func(HtmlTag, HTMLProps)
	Other    func(AnyNode)
}

func (ops renderOps) Perform(node AnyNode) {
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

func getKeyOrIndex(node AnyNode, index int) string {
	key := node.key()
	if key != nil {
		return fmt.Sprintf("key:%s", *key)
	}
	return fmt.Sprintf("idx:%d", index)
}

type Renderer func(fiber *fiber, rendered AnyNode)

func renderChildren(ancestor *fiber, children []AnyNode, renderer Renderer) {
	for i, childNode := range children {
		childFiber := ancestor.findChild(i, childNode)
		childFiber.node = childNode
		childFiber.dirty = true
		render(childFiber, renderer)
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
	renderer(fiber, nextRendered)
}

func RenderToString(node AnyNode) string {
	var builder strings.Builder
	root := newRoot(builder, node)
	var renderer Renderer
	renderer = func(fiber *fiber, nextNode AnyNode) {
		debug.Printf("RenderToString fiber %v, node %v", fiber, nextNode)
		renderOps{
			Nil: func() {},
			Fragment: func(props FragmentProps) {
				renderChildren(fiber, props.Children, renderer)
			},
			Text: func(props TextProps) {
				builder.WriteString(props.Text)
			},
			HTML: func(tag HtmlTag, props HTMLProps) {
				if tag.SelfClose && len(props.Children) == 0 {
					builder.WriteString(tag.SelfCloseTag(props))
				} else {
					builder.WriteString(tag.OpenTag(props))
					renderChildren(fiber, props.Children, renderer)
					builder.WriteString(tag.CloseTag())
				}
			},
			Other: func(node AnyNode) {
				renderChildren(fiber, []AnyNode{node}, renderer)
			},
		}.Perform(nextNode)
	}
	render(root.fiber, renderer)
	return builder.String()
}

type CounterProps struct {
	initial   *int
	increment *int
	WithKey
}

// Sketch
var Counter = FunctionComponent(func(props CounterProps) AnyNode {
	increment := Default(props.increment, 1)
	count, setCount := UseState(Default(props.initial, 0))
	handleClick := UseCallback(func() {
		nextCount := count + increment
		setCount(nextCount)
	}, struct{ increment, count int }{increment, count})

	return Div.Node(HTMLProps{},
		Text.F("Current count: %d ", count),
		Box.Node(BoxProps{
			HTMLProps: HTMLProps{
				onClick: Some(handleClick)},
			Padding: 3},
			Text.F("Increment by %d", increment)),
	)
})

type BoxProps struct {
	HTMLProps
	Padding int
}

var Box = FunctionComponent(func(props BoxProps) AnyNode {
	props.HTMLProps.className = Some("box")
	props.HTMLProps.style = Some(fmt.Sprintf("border: 1px solid black; padding: %d", props.Padding))
	return Div.Node(props.HTMLProps, props.Children...)
})
