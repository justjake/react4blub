package main

import (
	"fmt"
	"strings"
	"time"

	. "github.com/justjake/react4c/react"
	"github.com/justjake/react4c/testdom"
	. "github.com/justjake/react4c/web"
)

// Sketch
var Counter = FunctionComponent(func(props CounterProps) AnyNode {
	increment := Default(props.increment, 1)
	count, setCount := UseState(Default(props.initial, 0))
	handleClick := UseCallback(func() {
		nextCount := count + increment
		setCount(nextCount)
	}, struct{ increment, count int }{increment, count})

	return Div.Node(HTMLProps{},
		Text("My component:"),
		Text.F("Current count: %d ", count),
		Box.Node(BoxProps{
			HTMLProps: HTMLProps{
				OnClick: Some(handleClick)},
			Padding: 3},
			Text.F("Increment by %d", increment)),
	)
})

type ClockProps struct {
	WithKey
	Interval time.Duration
}

var Clock = FunctionComponent(func(props ClockProps) AnyNode {
	now, setNow := UseStateLazy(time.Now)
	UseEffect(func() func() {
		ticker := time.NewTicker(props.Interval)
		go func() {
			for range ticker.C {
				setNow(time.Now())
			}
		}()
		return func() { ticker.Stop() }
	}, props.Interval)
	return Text.F("%s", now)
})

type BoxProps struct {
	HTMLProps
	Padding int
}

var Box = FunctionComponent(func(props BoxProps) AnyNode {
	props.ClassName = Some("box")
	props.Style = Some(fmt.Sprintf("border: 1px solid black; padding: %d", props.Padding))
	return Div.Node(props.HTMLProps, props.Children...)
})

func main() {
	var foo = Fragment(
		Text("Counter:"),
		Counter.Node(CounterProps{initial: Some(5)}),
		Text("Clock:"),
		Clock.Node(ClockProps{Interval: time.Second}),
	)
	fmt.Println(RenderToString(foo))
}

func RenderToString(node AnyNode) string {
	var builder strings.Builder
	root := newRoot(builder, node)
	var renderer Renderer
	renderer = func(fiber *fiber, nextNode AnyNode) {
		componentKindHandlers{
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

func RenderToTestDom(node AnyNode, parentNode testdom.Node) testdom.Node {
	root := newRoot(parentNode, node)
	var renderer Renderer
	renderer = func(fiber *fiber, nextNode AnyNode) {
		updateMounted := componentKindHandlers{
			Nil: func() {},
			Fragment: func(props FragmentProps) {
				// TODO: how do we index this?
				renderChildren(fiber, props.Children, renderer)
			},
			Text: func(props TextProps) {
				if fiber.mounted != nil {
					prevProps := fiber.rendered.props().(TextProps)
					if prevProps.Text != props.Text {
						fiber.mounted.(testdom.Node).SetAttribute("innerText", props.Text)
					}
				} else {
					fiber.mounted = testdom.NewText(props.Text)
				}
			},
			HTML: func(tag HtmlTag, props HTMLProps) {
				prevProps := HTMLProps{}
				if fiber.mounted == nil {
					fiber.mounted = testdom.NewElement(tag.TagName)
				} else {
					// is this supposed to be node, or rendered???
					// seems like rendered
					prevProps = fiber.rendered.props().(HTMLProps)
				}

				element := fiber.mounted.(testdom.Node)
				applyHtmlPropDiff(prevProps.ClassName, props.ClassName, "class", element)
				applyHtmlPropDiff(prevProps.Style, props.Style, "style", element)
				applyHtmlPropDiff(prevProps.Id, props.Id, "id", element)
				applyHtmlPropDiff(prevProps.OnClick, props.OnClick, "onclick", element)
			},
			Other: func(node AnyNode) {
				renderChildren(fiber, []AnyNode{node}, renderer)
			},
		}
		updateMounted.Perform(nextNode)
	}
}

func applyHtmlPropDiff[T any](prev *T, next *T, name string, node testdom.Node) {
	if prev != nil && next == nil {
		node.DeleteAttribute(name)
	} else if next != nil {
		node.SetAttribute(name, *next)
	}
}

type CounterProps struct {
	initial   *int
	increment *int
	WithKey
}
