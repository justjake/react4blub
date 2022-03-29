package testdom

import (
	"fmt"
	"log"
	"strings"

	"golang.org/x/exp/slices"
)

type Node interface {
	Parent() Node
	Remove()
	RemoveChild(node Node)
	SetAttribute(attr string, val any)
	DeleteAttribute(attr string)
	AddChildBefore(node Node, before Node)
	AddChildAfter(node Node, after Node)
	setParent(parent Node)
	Summary() string
}

type Text struct {
	InnerText string
	parent    Node
}

func (el *Text) Summary() string {
	return "text"
}

func (el *Text) Parent() Node {
	return el.parent
}

func (el *Text) Remove() {
	if el.Parent() == nil {
		return
	}

	el.Parent().RemoveChild(el)
	el.parent = nil
}

func (el *Text) SetAttribute(attr string, val any) {
	if attr == "innerText" {
		el.InnerText = val.(string)
	}
}

func (el *Text) AddChildBefore(node Node, before Node) {
	panic("not implemented") // TODO: Implement
}

func (el *Text) AddChildAfter(node Node, after Node) {
	panic("not implemented") // TODO: Implement
}

func (el *Text) setParent(parent Node) {
	if el.Parent() != nil && el.Parent() != parent {
		el.Remove()
	}
	el.parent = parent
}

func NewText(innerText string) *Text {
	return &Text{InnerText: innerText}
}

type Element struct {
	parent     Node
	tagName    string
	Attributes map[string]any
	Children   []Node
}

func (el *Element) TagName() string {
	return el.tagName
}

func (el *Element) Summary() string {
	return fmt.Sprintf("<%s>", el.TagName())
}

func (el *Element) Parent() Node {
	return el.parent
}

func (el *Element) Remove() {
	if el.Parent() == nil {
		return
	}

	el.Parent().RemoveChild(el)
	el.parent = nil
}

func NewElement(tagName string) *Element {
	return &Element{nil, tagName, make(map[string]any), []Node{}}
}

func (parent *Element) Index(node Node) int {
	return slices.IndexFunc(parent.Children, func(it Node) bool { return it == node })
}

func (el *Element) SetAttribute(attr string, val any) {
	el.Attributes[attr] = val
}

func (el *Element) AddChildAt(node Node, index int) {
	node.setParent(el)
	el.Children = slices.Insert(el.Children, index, node)
}

func (el *Element) AddChildBefore(node Node, before Node) {
	idx := 0
	if before != nil {
		found := el.Index(before)
		if found != -1 {
			idx = found
		}
	}

	el.AddChildAt(node, idx)
}

func (el *Element) AddChildAfter(node Node, after Node) {
	idx := len(el.Children) - 1
	if after != nil {
		found := el.Index(after)
		if found != -1 {
			idx = found
		}
	}
	el.AddChildAt(node, idx+1)
}

func (el *Element) setParent(parent Node) {
	if el.Parent() != nil && el.Parent() != parent {
		el.Remove()
	}
	el.parent = parent
}

func (parent *Element) RemoveChild(node Node) {
	childIndex := parent.Index(node)
	if childIndex > -1 {
		parent.Children = slices.Delete(parent.Children, childIndex, childIndex)
	}
}
func (el *Text) RemoveChild(node Node) {
	panic("not implemented") // TODO: Implement
}

type NodeLogger struct {
	Node
	Logger *log.Logger
}

func (el *NodeLogger) Remove() {
	el.Logger.Printf("%s.Remove()", el)
	el.Node.Remove()
}

func (el *NodeLogger) RemoveChild(node Node) {
	el.Logger.Printf("%s.RemoveChild(%s)", el, node.Summary())
	el.Node.RemoveChild(node)
}

func (el *NodeLogger) SetAttribute(attr string, val string) {
	el.Logger.Printf("%s.SetAttribute(%q, %q)", el, attr, val)
	el.Node.SetAttribute(attr, val)
}

func (el *NodeLogger) AddChildBefore(node Node, before Node) {
	valSummary := "nil"
	if before != nil {
		valSummary = before.Summary()
	}
	el.Logger.Printf("%s.AddChildBefore(%s, %q)", el, node.Summary(), valSummary)
	el.Node.AddChildBefore(node, before)
}

func (el *NodeLogger) AddChildAfter(node Node, after Node) {
	valSummary := "nil"
	if after != nil {
		valSummary = after.Summary()
	}
	el.Logger.Printf("%s.AddChildAfter(%s, %q)", el, node.Summary(), valSummary)
	el.Node.AddChildBefore(node, after)
}

func (el *NodeLogger) String() string {
	var builder strings.Builder
	stack := []Node{}
	for parent := el.Node.Parent(); parent != nil; parent = parent.Parent() {
		stack = append(stack, parent)
	}
	for _, parent := range stack {
		builder.WriteString(fmt.Sprintf("%s > ", parent.Summary()))
	}
	builder.WriteString(el.Summary())
	return builder.String()
}

func (el *Text) DeleteAttribute(attr string) {
	panic("not implemented") // TODO: Implement
}

func (el *Element) DeleteAttribute(attr string) {
	delete(el.Attributes, attr)
}
