package main

import (
	"fmt"
	"strings"
)

func main() {
	var foo = Counter.Node(CounterProps{initial: Some(5)})
	actual := foo.(*Node[any])
	fmt.Printf("v:%v\nt:%t\nT:%T\n", actual, actual, actual)
	fmt.Println(RenderToString(foo))
}

func RenderToString(node AnyNode) string {
	var builder strings.Builder
	root := newRoot(builder, node)
	root.fiber.updateRendered()

	renderComp := func(renderedNode AnyNode) {
		switch comp := renderedNode.component().(type) {
		case nil:
			return
		case textComponent:
			props := node.props().(TextProps)
			builder.WriteString(props.Text)
			break
		case HtmlTag:
			props := node.props().(HTMLProps)
			children := node.children()
			if comp.SelfClose && len(children) == 0 {
				builder.WriteString(comp.SelfCloseTag(props))
			} else {
				builder.WriteString(comp.OpenTag(props))
				// TODO: render each child to string
				builder.WriteString(comp.CloseTag())
			}
		}
	}

	renderComp(root.fiber.rendered)

	// root.fiber
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
