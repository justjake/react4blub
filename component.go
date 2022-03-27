package main

import "fmt"

type Component[Props any] interface {
	Render(Props Props) AnyNode
	Node(Props Props, children ...AnyNode) AnyNode
}

type AnyNode interface {
	key() *string
	invokeRender() AnyNode
	component() any
}

func FunctionComponent[Props any](fn ComponentFunc[Props]) ComponentFunc[Props] {
	return ComponentFunc[Props](fn)
}

type ComponentFunc[Props any] func(props Props) AnyNode

func (c ComponentFunc[Props]) Render(props Props) AnyNode {
	return c(props)
}

func (c ComponentFunc[Props]) Node(props Props, children ...AnyNode) AnyNode {
	return H[Props](c, props, children...)
}

type Node[Props any] struct {
	Component Component[Props]
	Props     Props
	Key       *string
	Children  []AnyNode
	// Ref       Ref
}

func (n *Node[Props]) key() *string {
	return n.Key
}

func (n *Node[Props]) invokeRender() AnyNode {
	return n.Component.Render(n.Props)
}

func (n *Node[Props]) component() any {
	return n.Component
}

func H[Props any](comp Component[Props], props Props, children ...AnyNode) AnyNode {
	// if settable, ok := any(props).(ChildrenSetter); ok {
	// 	settable.SetChildren(children)
	// }

	node := Node[Props]{
		Component: comp,
		Props:     props,
		Key:       GetKey(props),
		Children:  children,
	}
	return &node
}

type HTMLProps struct {
	className *string
	style     *string
	id        *string
	onClick   *func()
}

type HtmlTagComponent struct {
	TagName       string
	AllowChildren bool
}

func (_ HtmlTagComponent) Render(props HTMLProps) AnyNode {
	return nil
}

func (tag HtmlTagComponent) Node(props HTMLProps, children ...AnyNode) AnyNode {
	return &Node[HTMLProps]{
		Component: tag,
		Props:     props,
		Key:       GetKey(props),
		Children:  children,
	}
}

var Div = HtmlTagComponent{
	TagName:       "div",
	AllowChildren: true,
}

type textComponent struct{}
type TextProps struct {
	Text string
}

var Text = textComponent{}

func (t *textComponent) Render(props TextProps) AnyNode {
	return nil
}

func (t *textComponent) Node(props TextProps, children ...AnyNode) AnyNode {
	return &Node[TextProps]{
		Component: t,
		Props:     props,
		Key:       GetKey(props),
		Children:  nil,
	}
}

func (t *textComponent) F(format string, a ...any) AnyNode {
	return t.Node(
		TextProps{fmt.Sprintf(format, a...)},
	)
}

type fragmentComponent struct{}
type FragmentProps struct{}

var Fragment = fragmentComponent{}

func (f *fragmentComponent) Render(props FragmentProps) AnyNode {
	return nil
}

func (f *fragmentComponent) Node(props FragmentProps, children ...AnyNode) AnyNode {
	return &Node[FragmentProps]{
		Component: f,
		Props:     props,
		Key:       GetKey(props),
		Children:  children,
	}
}

func (f *fragmentComponent) Of(children ...AnyNode) AnyNode {
	return f.Node(FragmentProps{}, children...)
}
