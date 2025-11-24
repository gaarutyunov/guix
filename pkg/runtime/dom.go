// +build js,wasm

package runtime

import (
	"fmt"
	"syscall/js"
)

// Mount creates a real DOM node from a VNode and appends it to parent
func Mount(vnode *VNode, parent js.Value) error {
	domNode, err := createDOMNode(vnode)
	if err != nil {
		return err
	}

	vnode.DOMNode = domNode
	parent.Call("appendChild", domNode)
	return nil
}

// createDOMNode creates a real DOM node from a VNode
func createDOMNode(vnode *VNode) (js.Value, error) {
	doc := js.Global().Get("document")

	switch vnode.Type {
	case TextNode:
		return doc.Call("createTextNode", vnode.Text), nil

	case ElementNode:
		elem := doc.Call("createElement", vnode.Tag)

		// Set attributes
		for key, value := range vnode.Attributes {
			elem.Call("setAttribute", key, value)
		}

		// Set properties
		for key, value := range vnode.Properties {
			elem.Set(key, value)
		}

		// Attach event handlers
		for name, handler := range vnode.Events {
			attachEventHandler(elem, name, handler, vnode)
		}

		// Mount children
		for _, child := range vnode.Children {
			childNode, err := createDOMNode(child)
			if err != nil {
				return js.Undefined(), err
			}
			child.DOMNode = childNode
			elem.Call("appendChild", childNode)
		}

		return elem, nil

	case FragmentNode:
		frag := doc.Call("createDocumentFragment")
		for _, child := range vnode.Children {
			childNode, err := createDOMNode(child)
			if err != nil {
				return js.Undefined(), err
			}
			child.DOMNode = childNode
			frag.Call("appendChild", childNode)
		}
		return frag, nil

	default:
		return js.Undefined(), fmt.Errorf("unsupported vnode type: %d", vnode.Type)
	}
}

// attachEventHandler attaches a Go event handler to a DOM element
func attachEventHandler(elem js.Value, eventName string, handler EventHandler, vnode *VNode) {
	jsFunc := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		if len(args) == 0 {
			return nil
		}

		jsEvent := args[0]

		// Create Go Event wrapper
		event := Event{
			Native: jsEvent,
			Type:   jsEvent.Get("type").String(),
		}

		// Extract target information
		target := jsEvent.Get("target")
		event.Target = EventTarget{
			Native: target,
		}

		// Get value if it exists
		if target.Get("value").Type() == js.TypeString {
			event.Target.Value = target.Get("value").String()
		}

		// Get checked if it exists
		if target.Get("checked").Type() == js.TypeBoolean {
			event.Target.Checked = target.Get("checked").Bool()
		}

		// Call the handler in a goroutine to avoid blocking the event loop
		go handler.Handler(event)

		return nil
	})

	// Store jsFunc for cleanup
	handler.jsFunc = jsFunc
	vnode.Events[eventName] = handler

	elem.Call("addEventListener", eventName, jsFunc)
}

// Unmount removes a VNode from the DOM and cleans up resources
func Unmount(vnode *VNode) {
	if vnode == nil || vnode.DOMNode.IsUndefined() {
		return
	}

	// Clean up event handlers
	for _, handler := range vnode.Events {
		if !handler.jsFunc.IsUndefined() {
			handler.jsFunc.Release()
		}
	}

	// Recursively unmount children
	for _, child := range vnode.Children {
		Unmount(child)
	}

	// Remove from DOM
	parent := vnode.DOMNode.Get("parentNode")
	if !parent.IsUndefined() && !parent.IsNull() {
		parent.Call("removeChild", vnode.DOMNode)
	}

	vnode.DOMNode = js.Undefined()
}

// UpdateElement updates a DOM element based on attribute/property changes
func UpdateElement(vnode *VNode, oldAttrs, newAttrs map[string]string, oldProps, newProps map[string]interface{}) {
	if vnode.DOMNode.IsUndefined() {
		return
	}

	elem := vnode.DOMNode

	// Remove old attributes
	for key := range oldAttrs {
		if _, exists := newAttrs[key]; !exists {
			elem.Call("removeAttribute", key)
		}
	}

	// Set new/updated attributes
	for key, value := range newAttrs {
		if oldVal, exists := oldAttrs[key]; !exists || oldVal != value {
			elem.Call("setAttribute", key, value)
		}
	}

	// Update properties
	for key, value := range newProps {
		elem.Set(key, value)
	}
}

// SetTextContent updates the text content of a text node
func SetTextContent(vnode *VNode, text string) {
	if vnode.DOMNode.IsUndefined() {
		return
	}
	vnode.DOMNode.Set("textContent", text)
	vnode.Text = text
}

// ReplaceNode replaces an old VNode with a new one in the DOM
func ReplaceNode(oldVNode, newVNode *VNode) error {
	if oldVNode.DOMNode.IsUndefined() {
		return fmt.Errorf("old vnode has no DOM node")
	}

	parent := oldVNode.DOMNode.Get("parentNode")
	if parent.IsUndefined() || parent.IsNull() {
		return fmt.Errorf("old vnode has no parent")
	}

	// Create new DOM node
	newDOMNode, err := createDOMNode(newVNode)
	if err != nil {
		return err
	}

	newVNode.DOMNode = newDOMNode

	// Replace in DOM
	parent.Call("replaceChild", newDOMNode, oldVNode.DOMNode)

	// Clean up old node
	Unmount(oldVNode)

	return nil
}

// InsertBefore inserts a VNode before a reference node
func InsertBefore(parent js.Value, vnode *VNode, referenceNode js.Value) error {
	domNode, err := createDOMNode(vnode)
	if err != nil {
		return err
	}

	vnode.DOMNode = domNode
	parent.Call("insertBefore", domNode, referenceNode)
	return nil
}

// MoveNode moves a VNode to a new position
func MoveNode(vnode *VNode, parent js.Value, beforeNode js.Value) {
	if vnode.DOMNode.IsUndefined() {
		return
	}

	if beforeNode.IsUndefined() || beforeNode.IsNull() {
		parent.Call("appendChild", vnode.DOMNode)
	} else {
		parent.Call("insertBefore", vnode.DOMNode, beforeNode)
	}
}
