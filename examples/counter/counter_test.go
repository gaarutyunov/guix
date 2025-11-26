package main

import (
	"testing"
	"time"
)

// TestCounterChannelListener tests that the channel listener updates the counter value
func TestCounterChannelListener(t *testing.T) {
	// Create a counter with a channel
	counterChan := make(chan int, 10)
	counter := NewCounter(WithCounterChannel(counterChan))

	// Manually start the listener (simulating what BindApp does)
	updateCalled := false
	go func() {
		for val := range counter.CounterChannel {
			counter.currentCounterChannel = val
			updateCalled = true
		}
	}()

	// Give the goroutine time to start
	time.Sleep(10 * time.Millisecond)

	// Send a value to the channel
	counterChan <- 42

	// Give the goroutine time to process
	time.Sleep(10 * time.Millisecond)

	// Verify the counter value was updated
	if counter.currentCounterChannel != 42 {
		t.Errorf("Expected currentCounterChannel to be 42, got %d", counter.currentCounterChannel)
	}

	if !updateCalled {
		t.Error("Expected update to be triggered when channel receives value")
	}

	close(counterChan)
}

// TestCounterChannelMultipleValues tests that multiple values update the counter correctly
func TestCounterChannelMultipleValues(t *testing.T) {
	counterChan := make(chan int, 10)
	counter := NewCounter(WithCounterChannel(counterChan))

	updateCount := 0
	go func() {
		for val := range counter.CounterChannel {
			counter.currentCounterChannel = val
			updateCount++
		}
	}()

	time.Sleep(10 * time.Millisecond)

	// Send multiple values
	values := []int{10, 25, 100}
	for _, v := range values {
		counterChan <- v
		time.Sleep(10 * time.Millisecond)

		if counter.currentCounterChannel != v {
			t.Errorf("Expected currentCounterChannel to be %d, got %d", v, counter.currentCounterChannel)
		}
	}

	// Verify all values were processed
	if updateCount != len(values) {
		t.Errorf("Expected %d updates, got %d", len(values), updateCount)
	}

	close(counterChan)
}

// TestCounterWithoutChannel tests that counter works without a channel
func TestCounterWithoutChannel(t *testing.T) {
	counter := NewCounter()

	// Should have zero value
	if counter.currentCounterChannel != 0 {
		t.Errorf("Expected currentCounterChannel to be 0, got %d", counter.currentCounterChannel)
	}

	// Counter should be created successfully without channel
	if counter.CounterChannel != nil {
		t.Error("Expected CounterChannel to be nil when not provided")
	}
}

// TestCounterCreationProblem demonstrates the issue with inline Counter creation
func TestCounterCreationProblem(t *testing.T) {
	t.Log("This test demonstrates the problem with how Counter is created in App.Render()")

	// This is what happens in App.Render() - a new Counter is created each time
	createCounterInline := func() *Counter {
		counterChan := make(chan int, 10)
		counter := NewCounter(WithCounterChannel(counterChan))
		// BindApp is NEVER called, so the listener is never started
		return counter
	}

	counter1 := createCounterInline()

	// Send a value to the channel
	if counter1.CounterChannel != nil {
		counter1.CounterChannel <- 42
		time.Sleep(10 * time.Millisecond)

		// The value will NOT be updated because BindApp was never called
		if counter1.currentCounterChannel == 42 {
			t.Error("Counter updated without BindApp being called - this shouldn't happen")
		} else {
			t.Log("CONFIRMED: Counter does NOT update when BindApp is not called")
			t.Log("This is why the e2e tests are failing!")
		}
	}
}
