//go:build js && wasm
// +build js,wasm

package runtime

import (
	"fmt"
	"syscall/js"
)

// console provides access to browser console for debugging
var console = js.Global().Get("console")

// log writes a debug message to the browser console
func log(args ...interface{}) {
	// Convert all args to strings to avoid js.ValueOf errors
	jsArgs := make([]interface{}, len(args))
	for i, arg := range args {
		jsArgs[i] = fmt.Sprint(arg)
	}
	console.Call("log", jsArgs...)
}

// logError writes an error message to the browser console
func logError(args ...interface{}) {
	// Convert all args to strings to avoid js.ValueOf errors
	jsArgs := make([]interface{}, len(args))
	for i, arg := range args {
		jsArgs[i] = fmt.Sprint(arg)
	}
	console.Call("error", jsArgs...)
}

// Mount creates a real DOM node from a VNode and appends it to parent
func Mount(vnode *VNode, parent js.Value) error {
	log("DOM: Mount called for vnode type:", vnode.Type, "tag:", vnode.Tag)
	domNode, err := createDOMNode(vnode)
	if err != nil {
		logError("DOM: Mount failed to create DOM node:", err)
		return err
	}

	vnode.DOMNode = domNode
	parent.Call("appendChild", domNode)
	log("DOM: Successfully mounted", vnode.Tag)
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
			log("DOM: Attaching event handler:", name, "to element:", vnode.Tag)
			attachEventHandler(elem, name, handler, vnode)
		}

		// Mount children (skip webgpu-scene wrappers)
		var sceneComponent Scene
		for _, child := range vnode.Children {
			// Check if this is a webgpu-scene wrapper
			if child.Type == ElementNode && child.Tag == "webgpu-scene" {
				// Extract Scene from the wrapper
				if sceneValue, hasScene := child.Properties["scene"]; hasScene {
					if scene, ok := sceneValue.(Scene); ok {
						sceneComponent = scene
						log("DOM: Found Scene component in canvas children")
					}
				}
				continue // Don't mount the wrapper as a DOM node
			}

			childNode, err := createDOMNode(child)
			if err != nil {
				return js.Undefined(), err
			}
			child.DOMNode = childNode
			elem.Call("appendChild", childNode)
		}

		// Special handling for canvas elements with WebGPU scene
		if vnode.Tag == "canvas" && sceneComponent != nil {
			log("DOM: Initializing WebGPU canvas with scene")
			go initializeWebGPUCanvas(elem, sceneComponent, vnode)
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
			log("DOM: Event handler called with no args")
			return nil
		}

		jsEvent := args[0]
		log("DOM: Event fired:", eventName, "on element:", elem.Get("tagName"))

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
			log("DOM: Event target value:", event.Target.Value)
		}

		// Get checked if it exists
		if target.Get("checked").Type() == js.TypeBoolean {
			event.Target.Checked = target.Get("checked").Bool()
		}

		// Extract keyboard event fields if they exist
		if jsEvent.Get("key").Type() == js.TypeString {
			event.Key = jsEvent.Get("key").String()
		}
		if jsEvent.Get("code").Type() == js.TypeString {
			event.Code = jsEvent.Get("code").String()
		}
		if jsEvent.Get("ctrlKey").Type() == js.TypeBoolean {
			event.CtrlKey = jsEvent.Get("ctrlKey").Bool()
		}
		if jsEvent.Get("shiftKey").Type() == js.TypeBoolean {
			event.ShiftKey = jsEvent.Get("shiftKey").Bool()
		}
		if jsEvent.Get("altKey").Type() == js.TypeBoolean {
			event.AltKey = jsEvent.Get("altKey").Bool()
		}
		if jsEvent.Get("metaKey").Type() == js.TypeBoolean {
			event.MetaKey = jsEvent.Get("metaKey").Bool()
		}

		log("DOM: Calling event handler in goroutine")
		// Call the handler in a goroutine to avoid blocking the event loop
		go func() {
			defer func() {
				if r := recover(); r != nil {
					logError("DOM: Event handler panicked:", r)
				}
			}()
			handler.Handler(event)
			log("DOM: Event handler completed")
		}()

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
func UpdateElement(vnode *VNode, oldAttrs, newAttrs map[string]string, oldProps, newProps map[string]interface{}, oldEvents, newEvents map[string]EventHandler) {
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

	// Update event handlers
	// Remove old handlers
	for name, oldHandler := range oldEvents {
		if _, exists := newEvents[name]; !exists {
			// Event no longer exists, clean up
			if oldHandler.jsFunc.Value.Truthy() {
				oldHandler.jsFunc.Release()
			}
		}
	}

	// Add/update new handlers
	for name, newHandler := range newEvents {
		if oldHandler, exists := oldEvents[name]; exists {
			// Handler exists, clean up old one first
			if oldHandler.jsFunc.Value.Truthy() {
				oldHandler.jsFunc.Release()
			}
		}

		// Attach new handler
		jsFunc := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			if len(args) > 0 {
				event := args[0]
				newHandler.Handler(Event{
					Native: event,
					Target: EventTarget{
						Value:   event.Get("target").Get("value").String(),
						Checked: event.Get("target").Get("checked").Bool(),
						Native:  event.Get("target"),
					},
					Type:     event.Get("type").String(),
					Key:      event.Get("key").String(),
					Code:     event.Get("code").String(),
					CtrlKey:  event.Get("ctrlKey").Bool(),
					ShiftKey: event.Get("shiftKey").Bool(),
					AltKey:   event.Get("altKey").Bool(),
					MetaKey:  event.Get("metaKey").Bool(),
				})
			}
			return nil
		})
		elem.Call("addEventListener", name, jsFunc)
		newHandler.jsFunc = jsFunc
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

// initializeWebGPUCanvas initializes WebGPU for a canvas element with a scene
func initializeWebGPUCanvas(canvasElem js.Value, scene Scene, vnode *VNode) {
	log("WebGPU: Initializing canvas")

	// Check WebGPU support
	if !IsWebGPUSupported() {
		logError("WebGPU is not supported in this browser")
		showCanvasError(canvasElem, "WebGPU is not supported. Please use Chrome 113+ or Edge 113+")
		return
	}

	// Initialize WebGPU
	log("WebGPU: Initializing WebGPU context")
	gpuCtx, err := InitWebGPU()
	if err != nil {
		logError("WebGPU: Failed to initialize:", err)
		showCanvasError(canvasElem, fmt.Sprintf("Failed to initialize WebGPU: %v", err))
		return
	}

	log("WebGPU: WebGPU initialized successfully")

	// Get canvas dimensions from attributes or use defaults
	width := 800
	height := 600
	if w, ok := vnode.Properties["width"]; ok {
		if wInt, ok := w.(int); ok {
			width = wInt
		}
	}
	if h, ok := vnode.Properties["height"]; ok {
		if hInt, ok := h.(int); ok {
			height = hInt
		}
	}

	// Create GPU canvas
	config := GPUCanvasConfig{
		Width:            width,
		Height:           height,
		DevicePixelRatio: 1.0,
		AlphaMode:        "premultiplied",
		FrameLoop:        "always",
	}

	log("WebGPU: Creating GPU canvas with config:", fmt.Sprintf("width=%d, height=%d", width, height))
	canvas, err := createGPUCanvasFromElement(canvasElem, config, gpuCtx)
	if err != nil {
		logError("WebGPU: Failed to create GPU canvas:", err)
		showCanvasError(canvasElem, fmt.Sprintf("Failed to create GPU canvas: %v", err))
		return
	}

	log("WebGPU: GPU canvas created successfully")

	// Render the scene
	sceneNode := scene.RenderScene()
	if sceneNode == nil {
		logError("WebGPU: Scene RenderScene() returned nil")
		showCanvasError(canvasElem, "Scene rendering failed")
		return
	}

	log("WebGPU: Creating scene renderer")
	renderer, err := NewSceneRenderer(canvas, sceneNode)
	if err != nil {
		logError("WebGPU: Failed to create renderer:", err)
		showCanvasError(canvasElem, fmt.Sprintf("Failed to create renderer: %v", err))
		return
	}

	log("WebGPU: Scene renderer created successfully")

	// Check for render update callback
	var renderUpdate func(float64, interface{})
	if updateValue, hasUpdate := vnode.Properties["gpuRenderUpdate"]; hasUpdate {
		if callback, ok := updateValue.(func(float64, interface{})); ok {
			renderUpdate = callback
			log("WebGPU: Found render update callback")
		}
	}

	// Set render function
	canvas.SetRenderFunc(func(c *GPUCanvas, delta float64) {
		// Call update callback if provided
		if renderUpdate != nil {
			renderUpdate(delta, renderer)
		}
		renderer.Render()
	})

	// Start render loop
	canvas.Start()

	log("WebGPU: Render loop started successfully")
}

// showCanvasError displays an error message on the canvas
func showCanvasError(canvasElem js.Value, message string) {
	parent := canvasElem.Get("parentElement")
	if parent.IsUndefined() || parent.IsNull() {
		return
	}

	doc := js.Global().Get("document")
	errorDiv := doc.Call("createElement", "div")
	errorDiv.Call("setAttribute", "style",
		"padding: 20px; background: #ff4444; color: white; border-radius: 8px; margin: 10px;")
	errorDiv.Set("innerHTML", fmt.Sprintf("<h3>WebGPU Error</h3><p>%s</p>", message))

	parent.Call("insertBefore", errorDiv, canvasElem)
}
