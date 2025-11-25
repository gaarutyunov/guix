//go:build js && wasm
// +build js,wasm

package runtime

import (
	"sync"
)

// Signal represents a reactive value that notifies subscribers when it changes
type Signal[T comparable] struct {
	value       T
	subscribers []func(T)
	mu          sync.RWMutex
	app         *App
}

// CreateSignal creates a new reactive signal
func CreateSignal[T comparable](initial T) *Signal[T] {
	return &Signal[T]{
		value:       initial,
		subscribers: make([]func(T), 0),
	}
}

// CreateSignalWithApp creates a new reactive signal with app binding
func CreateSignalWithApp[T comparable](initial T, app *App) *Signal[T] {
	return &Signal[T]{
		value:       initial,
		subscribers: make([]func(T), 0),
		app:         app,
	}
}

// Get returns the current value of the signal
func (s *Signal[T]) Get() T {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.value
}

// Set updates the signal's value and notifies all subscribers
func (s *Signal[T]) Set(newValue T) {
	s.mu.Lock()
	oldValue := s.value
	s.value = newValue
	subscribers := make([]func(T), len(s.subscribers))
	copy(subscribers, s.subscribers)
	app := s.app
	s.mu.Unlock()

	// Only notify if value actually changed
	if oldValue != newValue {
		// Notify all subscribers
		for _, sub := range subscribers {
			sub(newValue)
		}

		// Trigger app update if bound
		if app != nil {
			app.Update()
		}
	}
}

// Update modifies the signal's value using a function
func (s *Signal[T]) Update(fn func(T) T) {
	s.mu.Lock()
	oldValue := s.value
	s.value = fn(s.value)
	newValue := s.value
	subscribers := make([]func(T), len(s.subscribers))
	copy(subscribers, s.subscribers)
	app := s.app
	s.mu.Unlock()

	// Only notify if value actually changed
	if oldValue != newValue {
		// Notify all subscribers
		for _, sub := range subscribers {
			sub(newValue)
		}

		// Trigger app update if bound
		if app != nil {
			app.Update()
		}
	}
}

// Subscribe adds a subscriber that will be called when the signal changes
func (s *Signal[T]) Subscribe(callback func(T)) func() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.subscribers = append(s.subscribers, callback)

	// Return unsubscribe function
	return func() {
		s.mu.Lock()
		defer s.mu.Unlock()
		for i, sub := range s.subscribers {
			// Compare function pointers (note: this is a simplified approach)
			if &sub == &callback {
				s.subscribers = append(s.subscribers[:i], s.subscribers[i+1:]...)
				break
			}
		}
	}
}

// BindApp binds the signal to an app for automatic updates
func (s *Signal[T]) BindApp(app *App) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.app = app
}

// Computed represents a derived reactive value
type Computed[T comparable] struct {
	compute func() T
	signal  *Signal[T]
	deps    []interface{} // Signals this computed depends on
}

// CreateComputed creates a computed value that automatically updates when dependencies change
func CreateComputed[T comparable](compute func() T) *Computed[T] {
	c := &Computed[T]{
		compute: compute,
		signal:  CreateSignal(compute()),
		deps:    make([]interface{}, 0),
	}
	return c
}

// Get returns the current computed value
func (c *Computed[T]) Get() T {
	return c.signal.Get()
}

// AddDep adds a dependency signal
func (c *Computed[T]) AddDep(dep interface{}) {
	c.deps = append(c.deps, dep)

	// Subscribe to the dependency
	switch d := dep.(type) {
	case *Signal[int]:
		d.Subscribe(func(_ int) {
			c.signal.Set(c.compute())
		})
	case *Signal[string]:
		d.Subscribe(func(_ string) {
			c.signal.Set(c.compute())
		})
	case *Signal[bool]:
		d.Subscribe(func(_ bool) {
			c.signal.Set(c.compute())
		})
	// Add more types as needed
	}
}

// BindApp binds the computed signal to an app
func (c *Computed[T]) BindApp(app *App) {
	c.signal.BindApp(app)
}

// Effect represents a side effect that runs when dependencies change
type Effect struct {
	fn   func()
	deps []interface{}
}

// CreateEffect creates an effect that runs when dependencies change
func CreateEffect(fn func(), deps ...interface{}) *Effect {
	e := &Effect{
		fn:   fn,
		deps: deps,
	}

	// Subscribe to all dependencies
	for _, dep := range deps {
		switch d := dep.(type) {
		case *Signal[int]:
			d.Subscribe(func(_ int) {
				e.fn()
			})
		case *Signal[string]:
			d.Subscribe(func(_ string) {
				e.fn()
			})
		case *Signal[bool]:
			d.Subscribe(func(_ bool) {
				e.fn()
			})
		// Add more types as needed
		}
	}

	// Run initially
	fn()

	return e
}

// ReactiveComponent provides reactive component capabilities
type ReactiveComponent struct {
	BaseComponent
	signals []*Signal[int] // Track all signals used by this component
	mu      sync.Mutex
}

// TrackSignal registers a signal with the component for automatic cleanup
func (rc *ReactiveComponent) TrackSignal(signal *Signal[int]) {
	rc.mu.Lock()
	defer rc.mu.Unlock()
	rc.signals = append(rc.signals, signal)
}

// CleanupSignals removes all signal subscriptions
func (rc *ReactiveComponent) CleanupSignals() {
	rc.mu.Lock()
	defer rc.mu.Unlock()
	rc.signals = nil
}
