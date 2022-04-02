package react

import (
	"fmt"
)

type Component[Props IProps] interface {
	Render(Props Props) AnyNode
	// Create a Node with this type.
	// This is for convenience. TODO: remove...
	// Node(Props Props, children ...AnyNode) AnyNode
}

// ComponentFunc - user defined components
type ComponentFunc[Props IProps] func(props Props) AnyNode

func (c ComponentFunc[Props]) Render(props Props) AnyNode {
	return c(props)
}

func (c ComponentFunc[Props]) Node(props Props, children ...AnyNode) AnyNode {
	return JSX[Props](c, props, children...)
}

// FunctionComponent infers the ComponentFunc from a function that takes a
// IProps and returns an AnyNode.
func FunctionComponent[Props IProps](fn ComponentFunc[Props]) ComponentFunc[Props] {
	return ComponentFunc[Props](fn)
}

// Text component renders its contents as Text
var Text textComponent

type textComponent func(string) *Node[TextProps]

func init() {
	Text = textComponentImpl
	Fragment = fragmentComponentImpl
}

func textComponentImpl(s string) *Node[TextProps] {
	props := TextProps{Text: s}
	return JSX[TextProps](Text, props)
}

type TextProps struct {
	// Inner text of this Text node
	Text string
	WithKey
}

func (t textComponent) Render(props TextProps) AnyNode {
	return t(props.Text)
}

func (t textComponent) F(format string, a ...any) AnyNode {
	return t(fmt.Sprintf(format, a...))
}

// Fragment only renders its children.
var Fragment fragmentComponent

type fragmentComponent func(...AnyNode) *Node[FragmentProps]

func fragmentComponentImpl(children ...AnyNode) *Node[FragmentProps] {
	return JSX[FragmentProps](Fragment, FragmentProps{}, children...)
}

type FragmentProps struct {
	WithChildren
	WithKey
}

func (f fragmentComponent) Render(props FragmentProps) AnyNode {
	return f.Node(props)
}

func (f fragmentComponent) Node(props FragmentProps, children ...AnyNode) AnyNode {
	return JSX[FragmentProps](f, props, children...)
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
	return JSX[Props](m, props, children...)
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
