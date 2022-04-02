package web

import (
	"fmt"
	"strings"

	. "github.com/justjake/react4c/react"
)

// HtmlTag - literal HTML tag components
type HtmlTag struct {
	TagName   string
	SelfClose bool
}

type HTMLProps struct {
	WithChildren // TODO: HTMLPropsWithoutChildren?
	WithKey
	ClassName *string
	Style     *string
	Id        *string
	OnClick   *func()
}

func (tag HtmlTag) Render(props HTMLProps) AnyNode {
	return tag.Node(props)
}

func (tag HtmlTag) Node(props HTMLProps, children ...AnyNode) AnyNode {
	return JSX[HTMLProps](tag, props, children...)
}

func (tag HtmlTag) StartTag(props HTMLProps) string {
	var builder strings.Builder
	builder.WriteRune('<')
	builder.WriteString(tag.TagName)
	if props.Id != nil {
		builder.WriteString(fmt.Sprintf(` id="%s"`, *props.Id))
	}
	if props.ClassName != nil {
		builder.WriteString(fmt.Sprintf(` class="%s"`, *props.ClassName))
	}
	if props.Style != nil {
		builder.WriteString(fmt.Sprintf(` style="%s"`, *props.Style))
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
