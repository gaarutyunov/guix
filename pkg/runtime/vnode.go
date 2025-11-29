// Package runtime provides the Guix virtual DOM runtime for WebAssembly
package runtime

import (
	"syscall/js"
)

// VNodeType represents the type of a virtual node
type VNodeType uint8

const (
	// ElementNode represents a DOM element
	ElementNode VNodeType = iota
	// TextNode represents a text node
	TextNode
	// ComponentNode represents a component
	ComponentNode
	// FragmentNode represents a fragment (multiple children without wrapper)
	FragmentNode
)

// VNode represents a virtual DOM node
type VNode struct {
	Type       VNodeType
	Tag        string                 // For ElementNode (div, button, etc.)
	Text       string                 // For TextNode
	Key        interface{}            // For keyed reconciliation
	Attributes map[string]string      // HTML attributes
	Properties map[string]interface{} // DOM properties
	Events     map[string]EventHandler
	Children   []*VNode
	DOMNode    js.Value // Reference to actual DOM node after mount
	Component  Component
}

// EventHandler wraps a Go function for DOM event handling
type EventHandler struct {
	Name    string
	Handler func(Event)
	jsFunc  js.Func // Stored for cleanup
}

// Event wraps JavaScript event objects
type Event struct {
	Native js.Value
	Target EventTarget
	Type   string
}

// EventTarget represents an event target
type EventTarget struct {
	Value   string
	Checked bool
	Native  js.Value
}

// Component interface for all Guix components
type Component interface {
	Render() *VNode
	Mount(parent js.Value)
	Unmount()
	Update()
}

// Builder functions for creating VNodes

// El creates an element VNode with optional children and attributes
func El(tag string, options ...interface{}) *VNode {
	node := &VNode{
		Type:       ElementNode,
		Tag:        tag,
		Attributes: make(map[string]string),
		Properties: make(map[string]interface{}),
		Events:     make(map[string]EventHandler),
	}

	for _, opt := range options {
		switch o := opt.(type) {
		case *VNode:
			node.Children = append(node.Children, o)
		case Attr:
			node.Attributes[o.Key] = o.Value
		case Prop:
			node.Properties[o.Key] = o.Value
		case EventHandler:
			node.Events[o.Name] = o
		case Class:
			node.Attributes["class"] = string(o)
		case Style:
			node.Attributes["style"] = string(o)
		case Key:
			node.Key = o.Value
		}
	}

	return node
}

// Text creates a text VNode
func Text(content string) *VNode {
	return &VNode{
		Type: TextNode,
		Text: content,
	}
}

// Fragment creates a fragment node
func Fragment(children ...*VNode) *VNode {
	return &VNode{
		Type:     FragmentNode,
		Children: children,
	}
}

// Attribute types for builder pattern

// Attr represents an HTML attribute
type Attr struct {
	Key   string
	Value string
}

// Prop represents a DOM property
type Prop struct {
	Key   string
	Value interface{}
}

// Class represents a class attribute
type Class string

// Style represents a style attribute
type Style string

// Key represents a keyed node for reconciliation
type Key struct {
	Value interface{}
}

// Common element builders

// Div creates a div element
func Div(options ...interface{}) *VNode {
	return El("div", options...)
}

// Span creates a span element
func Span(options ...interface{}) *VNode {
	return El("span", options...)
}

// Button creates a button element
func Button(options ...interface{}) *VNode {
	return El("button", options...)
}

// Input creates an input element
func Input(options ...interface{}) *VNode {
	return El("input", options...)
}

// H1 creates an h1 element
func H1(options ...interface{}) *VNode {
	return El("h1", options...)
}

// H2 creates an h2 element
func H2(options ...interface{}) *VNode {
	return El("h2", options...)
}

// P creates a p element
func P(options ...interface{}) *VNode {
	return El("p", options...)
}

// A creates an a element
func A(options ...interface{}) *VNode {
	return El("a", options...)
}

// Img creates an img element
func Img(options ...interface{}) *VNode {
	return El("img", options...)
}

// Canvas creates a canvas element (for WebGPU)
func Canvas(options ...interface{}) *VNode {
	return El("canvas", options...)
}

// Common attribute helpers

// ID sets the id attribute
func ID(value string) Attr {
	return Attr{Key: "id", Value: value}
}

// ClassAttr sets the class attribute
func ClassAttr(value string) Attr {
	return Attr{Key: "class", Value: value}
}

// Href sets the href attribute
func Href(value string) Attr {
	return Attr{Key: "href", Value: value}
}

// Src sets the src attribute
func Src(value string) Attr {
	return Attr{Key: "src", Value: value}
}

// Type sets the type attribute
func Type(value string) Attr {
	return Attr{Key: "type", Value: value}
}

// Placeholder sets the placeholder attribute
func Placeholder(value string) Attr {
	return Attr{Key: "placeholder", Value: value}
}

// Value sets the value property
func Value(value string) Prop {
	return Prop{Key: "value", Value: value}
}

// Disabled sets the disabled property
func Disabled(value bool) Prop {
	return Prop{Key: "disabled", Value: value}
}

// GPUScene creates a special VNode wrapper for WebGPU Scene components
// This allows Scene components to be used as children of Canvas elements
func GPUScene(scene Scene) *VNode {
	return &VNode{
		Type:       ElementNode,
		Tag:        "webgpu-scene",
		Properties: map[string]interface{}{"scene": scene},
	}
}

// Event handler helpers

// OnClick creates a click event handler
func OnClick(handler func(Event)) EventHandler {
	return EventHandler{
		Name:    "click",
		Handler: handler,
	}
}

// OnInput creates an input event handler
func OnInput(handler func(Event)) EventHandler {
	return EventHandler{
		Name:    "input",
		Handler: handler,
	}
}

// OnChange creates a change event handler
func OnChange(handler func(Event)) EventHandler {
	return EventHandler{
		Name:    "change",
		Handler: handler,
	}
}

// OnSubmit creates a submit event handler
func OnSubmit(handler func(Event)) EventHandler {
	return EventHandler{
		Name:    "submit",
		Handler: handler,
	}
}

// OnKeyDown creates a keydown event handler
func OnKeyDown(handler func(Event)) EventHandler {
	return EventHandler{
		Name:    "keydown",
		Handler: handler,
	}
}

// OnKeyUp creates a keyup event handler
func OnKeyUp(handler func(Event)) EventHandler {
	return EventHandler{
		Name:    "keyup",
		Handler: handler,
	}
}

// OnMouseOver creates a mouseover event handler
func OnMouseOver(handler func(Event)) EventHandler {
	return EventHandler{
		Name:    "mouseover",
		Handler: handler,
	}
}

// OnMouseOut creates a mouseout event handler
func OnMouseOut(handler func(Event)) EventHandler {
	return EventHandler{
		Name:    "mouseout",
		Handler: handler,
	}
}

// WithKey sets a key for reconciliation
func WithKey(key interface{}) Key {
	return Key{Value: key}
}
