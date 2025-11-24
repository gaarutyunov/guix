//go:build js && wasm
// +build js,wasm

package runtime

import (
	"fmt"
	"syscall/js"
)

// App represents a Guix application
type App struct {
	root      js.Value
	rootVNode *VNode
	component Component
	mounted   bool
}

// NewApp creates a new Guix application
func NewApp(component Component) *App {
	return &App{
		component: component,
	}
}

// Mount mounts the application to a DOM element
func (a *App) Mount(selector string) error {
	doc := js.Global().Get("document")
	root := doc.Call("querySelector", selector)

	if root.IsUndefined() || root.IsNull() {
		return fmt.Errorf("element not found: %s", selector)
	}

	a.root = root
	return a.render()
}

// MountElement mounts the application to a specific DOM element
func (a *App) MountElement(element js.Value) error {
	if element.IsUndefined() || element.IsNull() {
		return fmt.Errorf("invalid element")
	}

	a.root = element
	return a.render()
}

// render performs the initial render or updates
func (a *App) render() error {
	newVNode := a.component.Render()

	if !a.mounted {
		// Initial render
		if err := Mount(newVNode, a.root); err != nil {
			return err
		}
		a.rootVNode = newVNode
		a.mounted = true
	} else {
		// Update render
		patches := Diff(a.rootVNode, newVNode)
		if err := ApplyPatches(patches); err != nil {
			return err
		}
		a.rootVNode = newVNode
	}

	return nil
}

// Update triggers a re-render of the application
func (a *App) Update() {
	ScheduleUpdate(func() {
		if err := a.render(); err != nil {
			fmt.Printf("Render error: %v\n", err)
		}
	})
}

// Unmount unmounts the application and cleans up resources
func (a *App) Unmount() {
	if a.rootVNode != nil {
		Unmount(a.rootVNode)
		a.rootVNode = nil
	}
	a.mounted = false
}

// Render is a convenience function to create and mount an app
func Render(selector string, component Component) (*App, error) {
	app := NewApp(component)
	if err := app.Mount(selector); err != nil {
		return nil, err
	}
	return app, nil
}

// BaseComponent provides a basic component implementation
type BaseComponent struct {
	app    *App
	vnode  *VNode
	parent js.Value
}

// Mount mounts the component
func (c *BaseComponent) Mount(parent js.Value) {
	c.parent = parent
	vnode := c.Render()
	Mount(vnode, parent)
	c.vnode = vnode
}

// Unmount unmounts the component
func (c *BaseComponent) Unmount() {
	if c.vnode != nil {
		Unmount(c.vnode)
		c.vnode = nil
	}
}

// Update triggers a component update
func (c *BaseComponent) Update() {
	if c.app != nil {
		c.app.Update()
	}
}

// Render should be overridden by specific components
func (c *BaseComponent) Render() *VNode {
	return Div(Text("BaseComponent"))
}

// WaitForDOMReady waits for the DOM to be ready before executing callback
func WaitForDOMReady(callback func()) {
	doc := js.Global().Get("document")

	if doc.Get("readyState").String() == "complete" ||
		doc.Get("readyState").String() == "interactive" {
		callback()
		return
	}

	ready := make(chan struct{})
	var cb js.Func
	cb = js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		cb.Release()
		close(ready)
		return nil
	})

	doc.Call("addEventListener", "DOMContentLoaded", cb)
	<-ready
	callback()
}
