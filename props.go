package main

type IProps interface {
	Keyed
}

type Keyed interface {
	GetKey() *string
}

func GetKey(val any) *string {
	if keyed, ok := val.(Keyed); ok {
		return keyed.GetKey()
	}
	return nil
}

type WithKey struct {
	Key *string
}

func Key(key string) WithKey {
	return WithKey{&key}
}

func (k WithKey) GetKey() *string {
	return k.Key
}

func Children(children []AnyNode) WithChildren {
	return WithChildren{children}
}

type WithChildren struct {
	Children []AnyNode
}

type HasChildren interface {
	GetChildren() []AnyNode
}

func GetChildren(val any) []AnyNode {
	if has, ok := val.(HasChildren); ok {
		return has.GetChildren()
	}
	return nil
}

type ChildrenSetter interface {
	SetChildren(newChildren []AnyNode)
}

func (c *WithChildren) GetChildren() []AnyNode {
	return c.Children
}

func (c *WithChildren) SetChildren(newChildren []AnyNode) {
	c.Children = newChildren
}
