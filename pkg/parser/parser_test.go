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

func TestParseMakeCall(t *testing.T) {
	source := `
package main

func App() {
	ch := make(chan int, 10)

	Div {
		"Test"
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
	if comp.Body == nil {
		t.Fatal("Expected component body")
	}

	if len(comp.Body.VarDecls) != 1 {
		t.Fatalf("Expected 1 variable declaration, got %d", len(comp.Body.VarDecls))
	}

	varDecl := comp.Body.VarDecls[0]
	if varDecl.Name != "ch" {
		t.Errorf("Expected variable name 'ch', got %s", varDecl.Name)
	}

	if varDecl.Value == nil {
		t.Fatal("Expected variable value")
	}

	if varDecl.Value.MakeCall == nil {
		t.Fatal("Expected make() call")
	}

	makeCall := varDecl.Value.MakeCall
	if makeCall.ChanType == nil {
		t.Fatal("Expected channel type in make() call")
	}

	if makeCall.ChanType.Name != "int" {
		t.Errorf("Expected chan type 'int', got %s", makeCall.ChanType.Name)
	}

	if makeCall.Size == nil {
		t.Fatal("Expected size argument in make() call")
	}
}

func TestParseMakeCallWithoutSize(t *testing.T) {
	source := `
package main

func App() {
	ch := make(chan string)

	Div {
		"Test"
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

	comp := file.Components[0]
	varDecl := comp.Body.VarDecls[0]

	if varDecl.Value.MakeCall == nil {
		t.Fatal("Expected make() call")
	}

	makeCall := varDecl.Value.MakeCall
	if makeCall.ChanType.Name != "string" {
		t.Errorf("Expected chan type 'string', got %s", makeCall.ChanType.Name)
	}

	if makeCall.Size != nil {
		t.Error("Expected no size argument in make() call")
	}
}
