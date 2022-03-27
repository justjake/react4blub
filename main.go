package main

import "fmt"

func main() {
	var foo = Counter.Node(CounterProps{initial: Some(5)})
	fmt.Printf("v:%v\nt:%t\nT:%T\n", foo, foo, foo)
	fmt.Println(RenderToString(foo))
}

func RenderToString(node AnyNode) string {

}

type CounterProps struct {
	MaybeString
	initial   *int
	increment *int
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
	WithChildren
	HTMLProps
	Padding int
}

var Box = FunctionComponent(func(props BoxProps) AnyNode {
	props.HTMLProps.className = Some("box")
	props.HTMLProps.style = Some(fmt.Sprintf("border: 1px solid black; padding: %d", props.Padding))
	return Div.Node(props.HTMLProps, props.Children...)
})
