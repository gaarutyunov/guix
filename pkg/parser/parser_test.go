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

	if makeCall.ChanSize == nil {
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

	if makeCall.ChanSize != nil {
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

// TestParseCustomFunctionCall tests parsing of custom (non-package-qualified) function calls
// EXPECTED STATE: Parser should handle function calls to custom functions just like stdlib calls
func TestParseCustomFunctionCall(t *testing.T) {
	source := `
package main

func App() (Component) {
	state := makeControlState(true, 1.0)

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
		t.Fatalf("Failed to parse custom function call: %v", err)
	}

	comp := file.Components[0]
	if comp.Body == nil {
		t.Fatal("Expected component body")
	}

	if len(comp.Body.VarDecls) != 1 {
		t.Fatalf("Expected 1 variable declaration, got %d", len(comp.Body.VarDecls))
	}

	varDecl := comp.Body.VarDecls[0]
	if len(varDecl.Names) != 1 || varDecl.Names[0] != "state" {
		t.Errorf("Expected variable name 'state', got %v", varDecl.Names)
	}

	if len(varDecl.Values) != 1 {
		t.Fatal("Expected one variable value")
	}

	if varDecl.Values[0].Left.CallOrSel == nil {
		t.Fatal("Expected call or selector expression for custom function call")
	}

	callOrSel := varDecl.Values[0].Left.CallOrSel
	if callOrSel.Base != "makeControlState" {
		t.Errorf("Expected base 'makeControlState', got %s", callOrSel.Base)
	}

	// Verify it's a call (has args)
	if callOrSel.Args == nil {
		t.Fatal("Expected function call (with args)")
	}

	if len(callOrSel.Args) != 2 {
		t.Errorf("Expected 2 arguments, got %d", len(callOrSel.Args))
	}
}

// TestParseMultipleVarDeclsWithFunctionCalls tests multiple variable declarations including function calls
// EXPECTED STATE: Parser should handle multiple var declarations with different expression types
func TestParseMultipleVarDeclsWithFunctionCalls(t *testing.T) {
	source := `
package main

func App() (Component) {
	autoRotate := true
	speed := 1.0
	commands := make(chan ControlCommand, 10)
	controlState := makeControlStateChannel(autoRotate, speed)

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
		t.Fatalf("Failed to parse multiple var decls with function calls: %v", err)
	}

	comp := file.Components[0]
	if comp.Body == nil {
		t.Fatal("Expected component body")
	}

	if len(comp.Body.VarDecls) != 4 {
		t.Fatalf("Expected 4 variable declarations, got %d", len(comp.Body.VarDecls))
	}

	// Verify first var decl (autoRotate := true)
	varDecl0 := comp.Body.VarDecls[0]
	if len(varDecl0.Names) != 1 || varDecl0.Names[0] != "autoRotate" {
		t.Errorf("Expected variable name 'autoRotate', got %v", varDecl0.Names)
	}
	if varDecl0.Values[0].Left.Literal == nil || varDecl0.Values[0].Left.Literal.Bool == nil {
		t.Error("Expected boolean literal for autoRotate")
	}

	// Verify second var decl (speed := 1.0)
	varDecl1 := comp.Body.VarDecls[1]
	if len(varDecl1.Names) != 1 || varDecl1.Names[0] != "speed" {
		t.Errorf("Expected variable name 'speed', got %v", varDecl1.Names)
	}
	if varDecl1.Values[0].Left.Literal == nil || varDecl1.Values[0].Left.Literal.Number == nil {
		t.Error("Expected number literal for speed")
	}

	// Verify third var decl (commands := make(chan ControlCommand, 10))
	varDecl2 := comp.Body.VarDecls[2]
	if len(varDecl2.Names) != 1 || varDecl2.Names[0] != "commands" {
		t.Errorf("Expected variable name 'commands', got %v", varDecl2.Names)
	}
	if varDecl2.Values[0].Left.MakeCall == nil {
		t.Error("Expected make() call for commands")
	}

	// Verify fourth var decl (controlState := makeControlStateChannel(autoRotate, speed))
	varDecl3 := comp.Body.VarDecls[3]
	if len(varDecl3.Names) != 1 || varDecl3.Names[0] != "controlState" {
		t.Errorf("Expected variable name 'controlState', got %v", varDecl3.Names)
	}
	if varDecl3.Values[0].Left.CallOrSel == nil {
		t.Fatal("Expected custom function call for controlState")
	}
	if varDecl3.Values[0].Left.CallOrSel.Base != "makeControlStateChannel" {
		t.Errorf("Expected function 'makeControlStateChannel', got %s", varDecl3.Values[0].Left.CallOrSel.Base)
	}
}

// TestParseVarDeclsWithComplexElements tests var declarations followed by complex UI elements
// EXPECTED STATE: Parser should correctly transition from var decls to UI elements with props
func TestParseVarDeclsWithComplexElements(t *testing.T) {
	source := `
package main

func App() (Component) {
	count := 0

	Div(Class("container"), TabIndex(0)) {
		Button(Class("btn")) {
			"Click me"
		}
	}
}
`
	p, err := New()
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}

	file, err := p.Parse(strings.NewReader(source))
	if err != nil {
		t.Fatalf("Failed to parse var decls with complex elements: %v", err)
	}

	comp := file.Components[0]
	if comp.Body == nil {
		t.Fatal("Expected component body")
	}

	// Verify variable declaration
	if len(comp.Body.VarDecls) != 1 {
		t.Fatalf("Expected 1 variable declaration, got %d", len(comp.Body.VarDecls))
	}

	varDecl := comp.Body.VarDecls[0]
	if varDecl.Names[0] != "count" {
		t.Errorf("Expected variable name 'count', got %s", varDecl.Names[0])
	}

	// Verify UI elements
	if len(comp.Body.Children) != 1 {
		t.Fatalf("Expected 1 child element, got %d", len(comp.Body.Children))
	}

	divNode := comp.Body.Children[0]
	if divNode.Element == nil {
		t.Fatal("Expected element node")
	}

	if divNode.Element.Tag != "Div" {
		t.Errorf("Expected element tag 'Div', got %s", divNode.Element.Tag)
	}

	// Verify Div has 2 props (Class and TabIndex)
	if len(divNode.Element.Props) != 2 {
		t.Fatalf("Expected 2 props on Div, got %d", len(divNode.Element.Props))
	}

	if divNode.Element.Props[0].Name != "Class" {
		t.Errorf("Expected first prop 'Class', got %s", divNode.Element.Props[0].Name)
	}

	if divNode.Element.Props[1].Name != "TabIndex" {
		t.Errorf("Expected second prop 'TabIndex', got %s", divNode.Element.Props[1].Name)
	}

	// Verify Div has 1 child (Button)
	if len(divNode.Element.Children) != 1 {
		t.Fatalf("Expected 1 child in Div, got %d", len(divNode.Element.Children))
	}
}

// TestParseWebGPUCubeApp tests parsing the actual WebGPU cube app.gx code that was failing
// EXPECTED STATE: Parser should handle the full app.gx file with all var decls and complex UI
func TestParseWebGPUCubeApp(t *testing.T) {
	source := `
package main

func App() (Component) {
	rotationX := 0.0
	rotationY := 0.0
	autoRotate := true
	speed := 1.0
	commands := make(chan ControlCommand, 10)
	controlState := makeControlStateChannel(autoRotate, speed)

	Div(
		Class("webgpu-container"),
		TabIndex(0)
	) {
		Canvas(
			ID("webgpu-canvas"),
			Width(600),
			Height(400)
		) {
			GPUScene(NewCubeScene(0, 0))
		}
		Controls(WithCommands(commands), WithState(controlState))
	}
}
`
	p, err := New()
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}

	file, err := p.Parse(strings.NewReader(source))
	if err != nil {
		t.Fatalf("Failed to parse WebGPU cube app: %v", err)
	}

	comp := file.Components[0]
	if comp.Body == nil {
		t.Fatal("Expected component body")
	}

	// Should have 6 variable declarations
	if len(comp.Body.VarDecls) != 6 {
		t.Fatalf("Expected 6 variable declarations, got %d", len(comp.Body.VarDecls))
	}

	// Verify last var decl is the custom function call
	lastVarDecl := comp.Body.VarDecls[5]
	if lastVarDecl.Names[0] != "controlState" {
		t.Errorf("Expected variable 'controlState', got %s", lastVarDecl.Names[0])
	}
	if lastVarDecl.Values[0].Left.CallOrSel == nil {
		t.Fatal("Expected custom function call")
	}
	if lastVarDecl.Values[0].Left.CallOrSel.Base != "makeControlStateChannel" {
		t.Errorf("Expected function 'makeControlStateChannel', got %s", lastVarDecl.Values[0].Left.CallOrSel.Base)
	}

	// Should have 1 top-level UI element (Div)
	if len(comp.Body.Children) != 1 {
		t.Fatalf("Expected 1 child element, got %d", len(comp.Body.Children))
	}

	divNode := comp.Body.Children[0]
	if divNode.Element == nil || divNode.Element.Tag != "Div" {
		t.Fatal("Expected Div element")
	}

	// Div should have 2 props (Class and TabIndex)
	if len(divNode.Element.Props) != 2 {
		t.Fatalf("Expected 2 props on Div, got %d", len(divNode.Element.Props))
	}

	// Div should have 2 children (Canvas and Controls)
	if len(divNode.Element.Children) != 2 {
		t.Fatalf("Expected 2 children in Div, got %d", len(divNode.Element.Children))
	}

	// Verify Canvas element
	canvasNode := divNode.Element.Children[0]
	if canvasNode.Element == nil || canvasNode.Element.Tag != "Canvas" {
		t.Fatal("Expected Canvas element")
	}
	if len(canvasNode.Element.Props) != 3 {
		t.Fatalf("Expected 3 props on Canvas, got %d", len(canvasNode.Element.Props))
	}

	// Verify Controls element
	controlsNode := divNode.Element.Children[1]
	if controlsNode.Element == nil || controlsNode.Element.Tag != "Controls" {
		t.Fatal("Expected Controls element")
	}
	if len(controlsNode.Element.Props) != 2 {
		t.Fatalf("Expected 2 props on Controls, got %d", len(controlsNode.Element.Props))
	}
}

// TestParseGoStatement tests parsing of goroutine statements
// EXPECTED STATE: Parser should recognize go func() { ... }() statements
func TestParseGoStatement(t *testing.T) {
	source := `
package main

func App() (Component) {
	state := make(chan int, 10)

	go func() {
		state <- 42
	}()

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
		t.Fatalf("Failed to parse go statement: %v", err)
	}

	comp := file.Components[0]
	if comp.Body == nil {
		t.Fatal("Expected component body")
	}

	// Should have 1 variable declaration (state)
	if len(comp.Body.VarDecls) != 1 {
		t.Fatalf("Expected 1 variable declaration, got %d", len(comp.Body.VarDecls))
	}

	// Should have 1 statement (the go statement)
	if len(comp.Body.Statements) != 1 {
		t.Fatalf("Expected 1 statement, got %d", len(comp.Body.Statements))
	}

	// Verify it's a go statement
	stmt := comp.Body.Statements[0]
	if stmt.GoStmt == nil {
		t.Fatal("Expected go statement")
	}

	// Verify the go statement has a function literal
	if stmt.GoStmt.Func == nil {
		t.Fatal("Expected function literal in go statement")
	}

	// Verify the function has a body
	if stmt.GoStmt.Func.Body == nil {
		t.Fatal("Expected function body in go statement")
	}

	// Verify the body has statements
	if len(stmt.GoStmt.Func.Body.Statements) == 0 {
		t.Fatal("Expected statements in function body")
	}
}
