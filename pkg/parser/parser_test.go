package parser

import (
	"strings"
	"testing"
)

func TestParseSimpleComponent(t *testing.T) {
	source := `
package main

func Button(label string) (Component) {
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

func Counter(count int) (Component) {
	Div {
		` + "`Counter {count}`" + `
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

func Counter(counterChannel <-chan int) (Component) {
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

func App() (Component) {
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
	if len(varDecl.Names) != 1 || varDecl.Names[0] != "ch" {
		t.Errorf("Expected variable name 'ch', got %v", varDecl.Names)
	}

	if len(varDecl.Values) != 1 {
		t.Fatal("Expected one variable value")
	}

	if varDecl.Values[0].Left.MakeCall == nil {
		t.Fatal("Expected make() call")
	}

	makeCall := varDecl.Values[0].Left.MakeCall
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

func App() (Component) {
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

	if len(varDecl.Values) != 1 || varDecl.Values[0].Left.MakeCall == nil {
		t.Fatal("Expected make() call")
	}

	makeCall := varDecl.Values[0].Left.MakeCall
	if makeCall.ChanType.Name != "string" {
		t.Errorf("Expected chan type 'string', got %s", makeCall.ChanType.Name)
	}

	if makeCall.Size != nil {
		t.Error("Expected no size argument in make() call")
	}
}

func TestParseSelector(t *testing.T) {
	source := `
package main

func Handler(e Event) (Component) {
	value := e.Target.Value

	Div {
		` + "`{value}`" + `
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
	if len(comp.Body.VarDecls) != 1 {
		t.Fatalf("Expected 1 variable declaration, got %d", len(comp.Body.VarDecls))
	}

	varDecl := comp.Body.VarDecls[0]
	if len(varDecl.Names) != 1 || varDecl.Names[0] != "value" {
		t.Errorf("Expected variable name 'value', got %v", varDecl.Names)
	}

	if len(varDecl.Values) != 1 {
		t.Fatal("Expected variable value")
	}

	if varDecl.Values[0].Left.CallOrSel == nil {
		t.Fatal("Expected call or selector expression")
	}

	callOrSel := varDecl.Values[0].Left.CallOrSel
	if callOrSel.Base != "e" {
		t.Errorf("Expected base 'e', got %s", callOrSel.Base)
	}

	expectedFields := []string{"Target", "Value"}
	if len(callOrSel.Fields) != len(expectedFields) {
		t.Fatalf("Expected %d fields, got %d", len(expectedFields), len(callOrSel.Fields))
	}

	for i, expected := range expectedFields {
		if callOrSel.Fields[i] != expected {
			t.Errorf("Expected field[%d] = %s, got %s", i, expected, callOrSel.Fields[i])
		}
	}

	// Verify it's a selector (no args)
	if callOrSel.Args != nil {
		t.Error("Expected selector (no args), but got call with args")
	}
}

func TestParseSelectorSingleField(t *testing.T) {
	source := `
package main

func Widget(obj Object) (Component) {
	name := obj.Name

	Div {
		` + "`{name}`" + `
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

	if len(varDecl.Values) != 1 || varDecl.Values[0].Left.CallOrSel == nil {
		t.Fatal("Expected call or selector expression")
	}

	callOrSel := varDecl.Values[0].Left.CallOrSel
	if callOrSel.Base != "obj" {
		t.Errorf("Expected base 'obj', got %s", callOrSel.Base)
	}

	if len(callOrSel.Fields) != 1 {
		t.Fatalf("Expected 1 field, got %d", len(callOrSel.Fields))
	}

	if callOrSel.Fields[0] != "Name" {
		t.Errorf("Expected field 'Name', got %s", callOrSel.Fields[0])
	}

	// Verify it's a selector (no args)
	if callOrSel.Args != nil {
		t.Error("Expected selector (no args), but got call with args")
	}
}

func TestParseMethodCall(t *testing.T) {
	source := `
package main

func Parser() (Component) {
	result := strconv.Atoi("123")

	Div {
		` + "`{result}`" + `
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

	if len(varDecl.Values) != 1 {
		t.Fatal("Expected one value")
	}

	if varDecl.Values[0].Left.CallOrSel == nil {
		t.Fatal("Expected call or selector expression")
	}

	callOrSel := varDecl.Values[0].Left.CallOrSel
	if callOrSel.Base != "strconv" {
		t.Errorf("Expected base 'strconv', got %s", callOrSel.Base)
	}

	if len(callOrSel.Fields) != 1 || callOrSel.Fields[0] != "Atoi" {
		t.Errorf("Expected fields ['Atoi'], got %v", callOrSel.Fields)
	}

	// Verify it's a call (has args)
	if callOrSel.Args == nil {
		t.Fatal("Expected call (with args), but got selector")
	}

	if len(callOrSel.Args) != 1 {
		t.Errorf("Expected 1 argument, got %d", len(callOrSel.Args))
	}
}

func TestParseMultipleAssignment(t *testing.T) {
	source := `
package main

func Handler() (Component) {
	n, err := strconv.Atoi("42")

	Div {
		` + "`{n}`" + `
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

	// Check multiple names
	if len(varDecl.Names) != 2 {
		t.Fatalf("Expected 2 names, got %d", len(varDecl.Names))
	}

	if varDecl.Names[0] != "n" {
		t.Errorf("Expected first name 'n', got %s", varDecl.Names[0])
	}

	if varDecl.Names[1] != "err" {
		t.Errorf("Expected second name 'err', got %s", varDecl.Names[1])
	}

	// Check single value (function call with multiple returns)
	if len(varDecl.Values) != 1 {
		t.Fatalf("Expected 1 value, got %d", len(varDecl.Values))
	}

	if varDecl.Values[0].Left.CallOrSel == nil {
		t.Fatal("Expected call or selector expression")
	}

	// Verify it's a call (has args)
	if varDecl.Values[0].Left.CallOrSel.Args == nil {
		t.Fatal("Expected function call (with args)")
	}
}
