package parser

import (
	"strings"
	"testing"
)

func TestParseSimpleComponent(t *testing.T) {
	source := `
package main

func Button(label: string) {
	Div {
	}
}
`
	p, err := New()
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}

	file, err := p.Parse(strings.NewReader(source))
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	if file.Package != "main" {
		t.Errorf("Expected package main, got %s", file.Package)
	}

	if len(file.Components) != 1 {
		t.Fatalf("Expected 1 component, got %d", len(file.Components))
	}

	comp := file.Components[0]
	if comp.Name != "Button" {
		t.Errorf("Expected component name Button, got %s", comp.Name)
	}

	if len(comp.Params) != 1 {
		t.Fatalf("Expected 1 parameter, got %d", len(comp.Params))
	}

	param := comp.Params[0]
	if param.Name != "label" {
		t.Errorf("Expected parameter name label, got %s", param.Name)
	}
}

func TestParseTemplate(t *testing.T) {
	source := `
package main

func Counter(count: int) {
	Div {
		` + "`Counter: {count}`" + `
	}
}
`
	p, err := New()
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}

	file, err := p.Parse(strings.NewReader(source))
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	if len(file.Components) != 1 {
		t.Fatalf("Expected 1 component, got %d", len(file.Components))
	}

	comp := file.Components[0]
	if comp.Body == nil || len(comp.Body.Children) == 0 {
		t.Fatal("Expected component to have children")
	}
}

func TestParseChannelParameter(t *testing.T) {
	source := `
package main

func Counter(counterChannel: <-chan int) {
	Div {
	}
}
`
	p, err := New()
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}

	file, err := p.Parse(strings.NewReader(source))
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	if len(file.Components) != 1 {
		t.Fatalf("Expected 1 component, got %d", len(file.Components))
	}

	comp := file.Components[0]
	if len(comp.Params) != 1 {
		t.Fatalf("Expected 1 parameter, got %d", len(comp.Params))
	}

	param := comp.Params[0]
	if param.Type == nil {
		t.Fatal("Expected parameter type")
	}
	if !param.Type.IsChannel {
		t.Error("Expected channel parameter")
	}
}
