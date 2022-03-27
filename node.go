package main

type AnyNode interface {
	key() *string
	props() any
	invokeRender() AnyNode
	children() []AnyNode
	component() any
	eachChild(func(node AnyNode))
}

type Node[Props IProps] struct {
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

func (n *Node[Props]) props() any {
	return n.Props
}

func (n *Node[Props]) children() []AnyNode {
	return n.Children
}

func (n *Node[Props]) eachChild(fn func(node AnyNode)) {
	for _, child := range n.Children {
		fn(child)
	}
}

// Create a Node
func H[Props IProps](comp Component[Props], props Props, children ...AnyNode) *Node[Props] {
	if settable, ok := any(props).(ChildrenSetter); ok {
		settable.SetChildren(children)
	}

	node := Node[Props]{
		Component: comp,
		Props:     props,
		Key:       GetKey(props),
		Children:  children,
	}

	return &node
}
