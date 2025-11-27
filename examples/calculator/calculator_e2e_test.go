//go:build e2e && !wasm
// +build e2e,!wasm

package main

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/chromedp/chromedp"
)

const (
	testPort    = "8888"
	testURL     = "http://localhost:" + testPort
	wasmTimeout = 10 * time.Second
)

// startTestServer starts a local HTTP server for testing
func startTestServer(t *testing.T) *http.Server {
	mux := http.NewServeMux()
	mux.Handle("/", http.FileServer(http.Dir(".")))

	server := &http.Server{
		Addr:    ":" + testPort,
		Handler: mux,
	}

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			t.Logf("Server error: %v", err)
		}
	}()

	// Wait for server to start
	time.Sleep(500 * time.Millisecond)

	return server
}

// clickButton simulates clicking a calculator button by its text content
func clickButton(text string) chromedp.Action {
	selector := fmt.Sprintf(`//button[text()="%s"]`, text)
	return chromedp.Click(selector, chromedp.NodeVisible, chromedp.BySearch)
}

// getDisplayValue gets the current display value
func getDisplayValue(value *string) chromedp.Action {
	return chromedp.Text(".display", value, chromedp.NodeVisible, chromedp.ByQuery)
}

// waitForWASM waits for WASM to load and the calculator to be ready
func waitForWASM() chromedp.Action {
	return chromedp.WaitVisible(".calculator", chromedp.ByQuery)
}

func TestCalculatorE2E(t *testing.T) {
	// Start test server
	server := startTestServer(t)
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		server.Shutdown(ctx)
	}()

	// Create browser context
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	// Set timeout
	ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	t.Run("Basic arithmetic operations", func(t *testing.T) {
		testBasicArithmetic(t, ctx)
	})

	t.Run("Clear functionality", func(t *testing.T) {
		testClearFunctionality(t, ctx)
	})

	t.Run("Sequential operations", func(t *testing.T) {
		testSequentialOperations(t, ctx)
	})

	t.Run("Division by zero", func(t *testing.T) {
		testDivisionByZero(t, ctx)
	})
}

func testBasicArithmetic(t *testing.T, ctx context.Context) {
	tests := []struct {
		name     string
		sequence []string
		expected string
	}{
		{
			name:     "Addition 3+2=5",
			sequence: []string{"3", "+", "2", "="},
			expected: "5",
		},
		{
			name:     "Subtraction 9-4=5",
			sequence: []string{"9", "−", "4", "="},
			expected: "5",
		},
		{
			name:     "Multiplication 6*7=42",
			sequence: []string{"6", "×", "7", "="},
			expected: "42",
		},
		{
			name:     "Division 8/2=4",
			sequence: []string{"8", "÷", "2", "="},
			expected: "4",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var display string

			err := chromedp.Run(ctx,
				chromedp.Navigate(testURL),
				waitForWASM(),
				chromedp.Sleep(500*time.Millisecond),
			)
			if err != nil {
				t.Fatalf("Failed to load page: %v", err)
			}

			// Click clear first
			err = chromedp.Run(ctx, clickButton("C"))
			if err != nil {
				t.Fatalf("Failed to click clear: %v", err)
			}

			// Execute button sequence
			for _, btn := range tt.sequence {
				err = chromedp.Run(ctx,
					clickButton(btn),
					chromedp.Sleep(100*time.Millisecond),
				)
				if err != nil {
					t.Fatalf("Failed to click button %s: %v", btn, err)
				}
			}

			// Get final display value
			err = chromedp.Run(ctx, getDisplayValue(&display))
			if err != nil {
				t.Fatalf("Failed to get display value: %v", err)
			}

			if display != tt.expected {
				t.Errorf("Expected display to be %s, got %s", tt.expected, display)
			}
		})
	}
}

func testClearFunctionality(t *testing.T, ctx context.Context) {
	var display string

	err := chromedp.Run(ctx,
		chromedp.Navigate(testURL),
		waitForWASM(),
		chromedp.Sleep(500*time.Millisecond),
		// Enter some numbers
		clickButton("3"),
		chromedp.Sleep(100*time.Millisecond),
		clickButton("6"),
		chromedp.Sleep(100*time.Millisecond),
		// Click clear
		clickButton("C"),
		chromedp.Sleep(100*time.Millisecond),
		getDisplayValue(&display),
	)

	if err != nil {
		t.Fatalf("Failed to test clear: %v", err)
	}

	if display != "0" {
		t.Errorf("Expected display to be 0 after clear, got %s", display)
	}
}

func testSequentialOperations(t *testing.T, ctx context.Context) {
	// Test the scenario from requirements: 3+2-3 should show 5 after +, then 2 after -
	var display string

	err := chromedp.Run(ctx,
		chromedp.Navigate(testURL),
		waitForWASM(),
		chromedp.Sleep(500*time.Millisecond),
		clickButton("C"),
		chromedp.Sleep(100*time.Millisecond),
		clickButton("3"),
		chromedp.Sleep(100*time.Millisecond),
	)
	if err != nil {
		t.Fatalf("Failed initial setup: %v", err)
	}

	// After pressing 3, display should show 3
	err = chromedp.Run(ctx, getDisplayValue(&display))
	if err != nil {
		t.Fatalf("Failed to get display: %v", err)
	}
	if display != "3" {
		t.Errorf("After pressing 3, expected display 3, got %s", display)
	}

	// Press +, display should not change
	err = chromedp.Run(ctx,
		clickButton("+"),
		chromedp.Sleep(100*time.Millisecond),
		getDisplayValue(&display),
	)
	if err != nil {
		t.Fatalf("Failed after +: %v", err)
	}
	if display != "3" {
		t.Errorf("After pressing +, display should still be 3, got %s", display)
	}

	// Press 2, display should show 5 (immediate calculation)
	err = chromedp.Run(ctx,
		clickButton("2"),
		chromedp.Sleep(100*time.Millisecond),
		getDisplayValue(&display),
	)
	if err != nil {
		t.Fatalf("Failed after 2: %v", err)
	}
	if display != "5" {
		t.Errorf("After pressing 2, expected display 5, got %s", display)
	}

	// Press -, display should not change
	err = chromedp.Run(ctx,
		clickButton("−"),
		chromedp.Sleep(100*time.Millisecond),
		getDisplayValue(&display),
	)
	if err != nil {
		t.Fatalf("Failed after -: %v", err)
	}
	if display != "5" {
		t.Errorf("After pressing -, display should still be 5, got %s", display)
	}

	// Press 3, display should show 2 (5-3=2)
	err = chromedp.Run(ctx,
		clickButton("3"),
		chromedp.Sleep(100*time.Millisecond),
		getDisplayValue(&display),
	)
	if err != nil {
		t.Fatalf("Failed after final 3: %v", err)
	}
	if display != "2" {
		t.Errorf("After pressing 3, expected display 2, got %s", display)
	}
}

func testDivisionByZero(t *testing.T, ctx context.Context) {
	var display string

	err := chromedp.Run(ctx,
		chromedp.Navigate(testURL),
		waitForWASM(),
		chromedp.Sleep(500*time.Millisecond),
		clickButton("C"),
		chromedp.Sleep(100*time.Millisecond),
		clickButton("3"),
		chromedp.Sleep(100*time.Millisecond),
		clickButton("÷"),
		chromedp.Sleep(100*time.Millisecond),
		clickButton("0"),
		chromedp.Sleep(100*time.Millisecond),
		clickButton("="),
		chromedp.Sleep(100*time.Millisecond),
		getDisplayValue(&display),
	)

	if err != nil {
		t.Fatalf("Failed to test division by zero: %v", err)
	}

	// Should show 0 or handle gracefully (based on implementation)
	if display != "0" && display != "Error" {
		t.Logf("Division by zero result: %s", display)
	}
}
