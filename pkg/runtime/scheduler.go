// +build js,wasm

package runtime

import (
	"sync"
	"syscall/js"
)

// Scheduler manages batched DOM updates using requestAnimationFrame
type Scheduler struct {
	pending   []func()
	scheduled bool
	mu        sync.Mutex
	raf       js.Value
}

// Global scheduler instance
var globalScheduler = &Scheduler{
	raf: js.Global().Get("requestAnimationFrame"),
}

// GetScheduler returns the global scheduler instance
func GetScheduler() *Scheduler {
	return globalScheduler
}

// Schedule adds a function to be executed in the next animation frame
func (s *Scheduler) Schedule(fn func()) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.pending = append(s.pending, fn)

	if !s.scheduled {
		s.scheduled = true
		s.raf.Invoke(js.FuncOf(s.flush))
	}
}

// flush executes all pending operations
func (s *Scheduler) flush(this js.Value, args []js.Value) interface{} {
	s.mu.Lock()
	work := s.pending
	s.pending = nil
	s.scheduled = false
	s.mu.Unlock()

	// Execute all pending work
	for _, fn := range work {
		fn()
	}

	return nil
}

// ScheduleUpdate is a convenience function to schedule an update
func ScheduleUpdate(fn func()) {
	globalScheduler.Schedule(fn)
}

// Immediate executes a function immediately without scheduling
func Immediate(fn func()) {
	fn()
}
