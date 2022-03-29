package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/justjake/react4c/testdom"
)

func main() {
	var foo = Fragment.Of(
		Counter.Node(CounterProps{initial: Some(5)}),
		Clock.Node(ClockProps{Interval: time.Second}),
	)
	fmt.Println(RenderToString(foo))
}

type componentKindHandlers struct {
	Nil      func()
	Fragment func(FragmentProps)
	Text     func(TextProps)
	HTML     func(HtmlTag, HTMLProps)
	Other    func(AnyNode)
}

func (ops componentKindHandlers) Perform(node AnyNode) {
	comp := node.component()
	props := node.props()

	if ops.Nil != nil && comp == nil {
		ops.Nil()
		return
	}

	switch comp := comp.(type) {
	case fragmentComponent:
		if ops.Fragment != nil {
			ops.Fragment(props.(FragmentProps))
			return
		}
	case textComponent:
		if ops.Text != nil {
			ops.Text(props.(TextProps))
			return
		}
	case HtmlTag:
		if ops.HTML != nil {
			ops.HTML(comp, props.(HTMLProps))
			return
		}
	}

	ops.Other(node)
}

func isPrimitiveComponent(node AnyNode) bool {
	primitive := true
	componentKindHandlers{
		Nil:      func() {},
		Fragment: func(FragmentProps) {},
		Text:     func(TextProps) {},
		HTML:     func(HtmlTag, HTMLProps) {},
		Other: func(AnyNode) {
			primitive = false
		},
	}.Perform(node)
	return primitive
}

func getKeyOrIndex(node AnyNode, index int) string {
	key := node.key()
	if key != nil {
		return fmt.Sprintf("key:%s", *key)
	}
	return fmt.Sprintf("idx:%d", index)
}

type Renderer func(fiber *fiber, nextRendered AnyNode)

func renderChildren(ancestor *fiber, children []AnyNode, renderer Renderer) {
	ancestorIndexOffset := 0
	if ancestor.mounted == nil {
		ancestorIndexOffset = ancestor.startIndex
	}
	widthOffset := 0
	// TODO: this is basically our reconciler algo...
	for i, childNode := range children {
		childFiber := ancestor.findChild(i, childNode)
		comp := childNode.component()
		prevNode := childFiber.node
		childFiber.retain = true // Mark

		if memo, ok := comp.(MemoComponent); ok && memo.SameComponent(prevNode.component()) {
			childFiber.dirty = childFiber.dirty || memo.PropsEqual(prevNode.props(), childNode.props())
		} else {
			childFiber.dirty = true
		}

		if childFiber.dirty {
			childFiber.node = childNode
		}

		childFiber.startIndex = i + widthOffset + ancestorIndexOffset
		render(childFiber, renderer)
		widthOffset += childFiber.indexOffsetWidth
	}

	// Unmount fibers no longer retained after this render
	ancestor.sweep() // Sweep
	if ancestor.mounted == nil {
		ancestor.indexOffsetWidth = widthOffset
		debug.Printf("renderChildren: %T with indexOffsetWidth %d", ancestor.node.component(), ancestorIndexOffset)
	}
	// TODO:::::: when should we move nodes around?
	// It seems like we need to phase our execution better:
	// 1. Reconcile and mark fibers
	//    - Pre-order depth first traversal:
	//    - For each new node
	//      - Find fiber for node
	//        - Create missing fiber
	//      - Mark fiber as alive
	//      - Recurse on node
	//        - Render each node -> rendered
	//        - Recurse on rendered
	//        - Assign DOM index in DOM parent
	//      - On way back up:
	//      - For each fiber
	//        - If not alive, mark Dead
	// 2. Sweep dead fibers
	//    - Post-order depth first traversal: (on way back up)
	//    - Sweep all dead fibers
	//      - To sweep a fiber:
	//        - Post order depth first traversal of the fiber
	//          - If fiber has Use(Layout)Effect cleanups: run cleanup
	//        - IFF fiber has no Dead ancestor, unmount it's top-level DOM nodes.
	//          (If there's a parent Dead ancestor, then that ancestor's sweep should unmount a parent DOM node)
	//          This will usually just be a single top-level DOM node,
	//          but may be several if we're unmounting <Fragment><Thing /><Thing /><Thing /></Fragment>
	//    - TODO: should we run useLayoutEffect cleanup here?
	// 3. Apply changes to DOM
	//    - Pre-order depth first traversal: (on way down)
	//      - Upsert DOM node for each fiber
	//        - Create DOM node
	//        - Mutate DOM node
	//    - Post-order depth first traversal: (on way up)
	//      - If node is a DOM node (or a top-level Fragment with no parent DOM node re-rendering)
	//        - Insert/Re-order all DOM children
	// 4. UseLayoutEffect (post-order DF)
	//    - Run cleanup
	//    - Run next effect
	// 5. Paint (somehow)
	// 6. UseEffect (post-order DF)
	//    - Run cleanup
	//    - Run next effect
}

func render(fiber *fiber, renderer Renderer) {
	if fiber.mounted != nil && !fiber.dirty {
		return
	}
	// This I'm always confused by.
	// If our custom component returned nil, what do?
	// Do we need to spawn another fiber depending on the child type?
	//
	// I think the algo is:
	// - if *fiber.node* is custom type, spawn child fiber for nextRendered w/ key, index 0
	// - if *fiber.node* is basic type, render nextRendered directly
	// TODO: change stuff around
	// TODO: support custom components
	// prevRendered := fiber.rendered

	nextRendered := fiber.invokeRenderWithHooks()
	if isPrimitiveComponent(fiber.node) {
		renderer(fiber, nextRendered)
	} else {
		renderChildren(fiber, []AnyNode{nextRendered}, renderer)
	}
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
				applyHtmlPropDiff(prevProps.className, props.className, "class", element)
				applyHtmlPropDiff(prevProps.style, props.style, "style", element)
				applyHtmlPropDiff(prevProps.id, props.id, "id", element)
				applyHtmlPropDiff(prevProps.onClick, props.onClick, "onclick", element)
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
	props.HTMLProps.className = Some("box")
	props.HTMLProps.style = Some(fmt.Sprintf("border: 1px solid black; padding: %d", props.Padding))
	return Div.Node(props.HTMLProps, props.Children...)
})
