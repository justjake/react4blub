package react

type IProps interface {
	Keyed
}

// TODO: Should we make children arrays type-safe?
type IPropsWithChildren interface {
	IProps
	HasChildren
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

type WithRef[T any] struct {
	Ref Ref[*T]
}

func (wr *WithRef[T]) GetRef() Ref[*T] {
	return wr.Ref
}

type HasRef[T any] interface {
	GetRef() Ref[T]
}
