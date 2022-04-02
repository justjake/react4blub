package react

type AnyNode interface {
	Keyed
	HasChildren
	ClearKey() bool
	GetProps() any
	InvokeRender() AnyNode
	GetComponent() any
}

type Node[Props IProps] struct {
	Component Component[Props]
	Props     Props
	Key       *string
	Children  []AnyNode
	// Ref       Ref
}

func (n *Node[Props]) GetKey() *string {
	return n.Key
}

func (n *Node[Props]) ClearKey() bool {
	if n.Key == nil {
		return false
	}
	n.Key = nil
	return true
}

func (n *Node[Props]) InvokeRender() AnyNode {
	return n.Component.Render(n.Props)
}

func (n *Node[Props]) GetComponent() any {
	return n.Component
}

func (n *Node[Props]) GetProps() any {
	return n.Props
}

func (n *Node[Props]) GetChildren() []AnyNode {
	return n.Children
}

// Create a Node
func JSX[Props IProps](comp Component[Props], props Props, children ...AnyNode) *Node[Props] {
	// var s ChildrenSetter
	// var boxProps BoxProps

	// s = &boxProps

	// _ = any(props).(ChildrenSetter)
	if settable, ok := any(&props).(ChildrenSetter); ok {
		settable.SetChildren(children)
	} else if len(children) > 0 {
		Logger.Printf("JSX: can't set children on %T: %#v <-x- %v", props, props, children)
	}

	node := Node[Props]{
		Component: comp,
		Props:     props,
		Key:       GetKey(props),
		Children:  children,
	}

	// Logger.Printf("JSX: create node <%T key=%v ...%#v>%#v</>", comp, node.GetKey(), node.props(), node.children())

	return &node
}
