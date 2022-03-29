package main

import (
	"fmt"
	"strings"
)

type Component[Props IProps] interface {
	Render(Props Props) AnyNode
	Node(Props Props, children ...AnyNode) AnyNode
}

// ComponentFunc - user defined components
type ComponentFunc[Props IProps] func(props Props) AnyNode

func (c ComponentFunc[Props]) Render(props Props) AnyNode {
	return c(props)
}

func (c ComponentFunc[Props]) Node(props Props, children ...AnyNode) AnyNode {
	return H[Props](c, props, children...)
}

// FunctionComponent infers the ComponentFunc from a function that takes a
// IProps and returns an AnyNode.
func FunctionComponent[Props IProps](fn ComponentFunc[Props]) ComponentFunc[Props] {
	return ComponentFunc[Props](fn)
}

// HtmlTag - literal HTML tag components
type HtmlTag struct {
	TagName   string
	SelfClose bool
}

type HTMLProps struct {
	WithChildren // TODO: HTMLPropsWithoutChildren?
	WithKey
	className *string
	style     *string
	id        *string
	onClick   *func()
}

func (tag HtmlTag) Render(props HTMLProps) AnyNode {
	return tag.Node(props)
}

func (tag HtmlTag) Node(props HTMLProps, children ...AnyNode) AnyNode {
	return H[HTMLProps](tag, props, children...)
}

func (tag HtmlTag) StartTag(props HTMLProps) string {
	var builder strings.Builder
	builder.WriteRune('<')
	builder.WriteString(tag.TagName)
	if props.id != nil {
		builder.WriteString(fmt.Sprintf(` id="%s"`, *props.id))
	}
	if props.className != nil {
		builder.WriteString(fmt.Sprintf(` class="%s"`, *props.className))
	}
	if props.style != nil {
		builder.WriteString(fmt.Sprintf(` style="%s"`, *props.style))
	}
	return builder.String()
}

func (tag HtmlTag) OpenTag(props HTMLProps) string {
	return tag.StartTag(props) + ">"
}

func (tag HtmlTag) CloseTag() string {
	return fmt.Sprintf("</%s>", tag.TagName)
}

func (tag HtmlTag) SelfCloseTag(props HTMLProps) string {
	return tag.StartTag(props) + "/>"
}

// HTML <div> tag
var Div = HtmlTag{
	TagName: "div",
}

// Text component renders its contents as Text
var Text = textComponent{}

type textComponent struct{}

type TextProps struct {
	// Inner text of this Text node
	Text string
	WithKey
}

func (t textComponent) Render(props TextProps) AnyNode {
	return t.Node(props)
}

func (t textComponent) Node(props TextProps, children ...AnyNode) AnyNode {
	return &Node[TextProps]{
		Component: t,
		Props:     props,
		Key:       GetKey(props),
		Children:  nil,
	}
}

func (t textComponent) F(format string, a ...any) AnyNode {
	return t.Node(
		TextProps{Text: fmt.Sprintf(format, a...)},
	)
}

// Fragment only renders its children.
var Fragment = fragmentComponent{}

type fragmentComponent struct{}
type FragmentProps struct {
	WithChildren
	WithKey
}

func (f fragmentComponent) Render(props FragmentProps) AnyNode {
	return f.Node(props)
}

func (f fragmentComponent) Node(props FragmentProps, children ...AnyNode) AnyNode {
	return H[FragmentProps](f, props, children...)
}

func (f fragmentComponent) Of(children ...AnyNode) AnyNode {
	return f.Node(FragmentProps{}, children...)
}

type ComparableProps interface {
	IProps
	comparable
}

type ComparableComponent[Props ComparableProps] interface {
	Component[Props]
	comparable
}

// ComparableMemoComponent will only re-render if its props change.
// By default components always re-render when a parent component re-renders.
// TODO: Maybe ComparableMemoComponent should be default?
type ComparableMemoComponent[Props ComparableProps, Comp ComparableComponent[Props]] struct {
	comp *Comp
}

type MemoComponent interface {
	SameComponent(otherComponent any) bool
	PropsEqual(prev any, next any) bool
}

func (m *ComparableMemoComponent[Props, Comp]) Render(props Props) AnyNode {
	return (*m.comp).Render(props)
}

func (m *ComparableMemoComponent[Props, Comp]) Node(props Props, children ...AnyNode) AnyNode {
	return H[Props](m, props, children...)
}

func Memo[Props ComparableProps, Comp ComparableComponent[Props]](base Comp) *ComparableMemoComponent[Props, Comp] {
	return &ComparableMemoComponent[Props, Comp]{&base}
}

func (m *ComparableMemoComponent[Props, Comp]) SameComponent(other any) bool {
	if comparable, ok := other.(*ComparableMemoComponent[Props, Comp]); ok {
		return m == comparable
	}
	return false
}

func (*ComparableMemoComponent[Props, Comp]) PropsEqual(prev any, next any) bool {
	comparablePrev, prevOk := prev.(Props)
	if !prevOk {
		return false
	}
	comparableNext, nextOk := next.(Props)
	if !nextOk {
		return false
	}
	return comparablePrev == comparableNext
}
