package codegen

import (
	"strings"
	"testing"

	"github.com/gaarutyunov/guix/pkg/parser"
)

// normalizeWhitespace normalizes whitespace in generated code for comparison
func normalizeWhitespace(s string) string {
	lines := strings.Split(s, "\n")
	var normalized []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" {
			normalized = append(normalized, trimmed)
		}
	}
	return strings.Join(normalized, "\n")
}

func TestGenerateSimpleComponent(t *testing.T) {
	tests := []struct {
		name     string
		source   string
		contains []string // Strings that should be present in the generated code
	}{
		{
			name: "simple component with string parameter",
			source: `package main

func Button(label: string) {
	Div {
		Span {
			` + "`{label}`" + `
		}
	}
}`,
			contains: []string{
				"type ButtonProps struct",
				"Label string",
				"type ButtonOption func(*Button)",
				"func WithLabel(v string) ButtonOption",
				"func NewButton(opts ...ButtonOption) *Button",
				"func (c *Button) Render() *runtime.VNode",
				"runtime.Div(",
				"runtime.Span(",
				"runtime.Text(",
			},
		},
		{
			name: "component with channel parameter",
			source: `package main

func Counter(counterChannel: chan int) {
	Div(Class("counter-display")) {
		Span(Class("counter-value")) {
			` + "`Counter: {<-counterChannel}`" + `
		}
	}
}`,
			contains: []string{
				"type CounterProps struct",
				"CounterChannel chan int",
				"type Counter struct",
				"CounterChannel chan int",
				"currentCounterChannel int",
				"func (c *Counter) BindApp(app *runtime.App)",
				"func (c *Counter) startCounterChannelListener()",
				"go func() {",
				"for val := range c.CounterChannel {",
				"c.currentCounterChannel = val",
				"c.app.Update()",
				"func WithCounterChannel(v chan int) CounterOption",
			},
		},
		{
			name: "component with multiple channels",
			source: `package main

func Dashboard(dataChannel: chan string, statusChannel: chan int) {
	Div {
		Span {
			` + "`Data: {<-dataChannel}`" + `
		}
		Span {
			` + "`Status: {<-statusChannel}`" + `
		}
	}
}`,
			contains: []string{
				"type DashboardProps struct",
				"DataChannel",
				"chan string",
				"StatusChannel",
				"chan int",
				"type Dashboard struct",
				"currentDataChannel",
				"currentStatusChannel",
				"func (c *Dashboard) BindApp(app *runtime.App)",
				"if c.DataChannel != nil {",
				"c.startDataChannelListener()",
				"if c.StatusChannel != nil {",
				"c.startStatusChannelListener()",
				"func (c *Dashboard) startDataChannelListener()",
				"for val := range c.DataChannel {",
				"c.currentDataChannel = val",
				"func (c *Dashboard) startStatusChannelListener()",
				"for val := range c.StatusChannel {",
				"c.currentStatusChannel = val",
			},
		},
		{
			name: "component with mixed parameters",
			source: `package main

func UserCard(name: string, updateChannel: chan string, age: int) {
	Div(Class("user-card")) {
		H1 {
			` + "`{name}`" + `
		}
		P {
			` + "`Age: {age}`" + `
		}
		Span {
			` + "`Update: {<-updateChannel}`" + `
		}
	}
}`,
			contains: []string{
				"type UserCardProps struct",
				"Name",
				"UpdateChannel",
				"chan string",
				"Age",
				"type UserCard struct",
				"currentUpdateChannel",
				"func WithName(v string) UserCardOption",
				"func WithUpdateChannel(v chan string) UserCardOption",
				"func WithAge(v int) UserCardOption",
				"func (c *UserCard) BindApp(app *runtime.App)",
				"if c.UpdateChannel != nil {",
				"c.startUpdateChannelListener()",
				"for val := range c.UpdateChannel {",
				"c.currentUpdateChannel = val",
			},
		},
		{
			name: "component with complex HTML structure",
			source: `package main

func Card(title: string, content: string) {
	Div(Class("card")) {
		Div(Class("card-header")) {
			H2(Class("card-title")) {
				` + "`{title}`" + `
			}
		}
		Div(Class("card-body")) {
			P(Class("card-content")) {
				` + "`{content}`" + `
			}
		}
		Div(Class("card-footer")) {
			Button(Class("btn"), Type("button")) {
				"Click me"
			}
		}
	}
}`,
			contains: []string{
				"type CardProps struct",
				"Title",
				"Content",
				"func (c *Card) Render() *runtime.VNode",
				"runtime.Div(",
				"runtime.Class(\"card\")",
				"runtime.H2(",
				"runtime.Class(\"card-title\")",
				"runtime.P(",
				"runtime.Button(",
				"runtime.Type(\"button\")",
			},
		},
		{
			name: "component without parameters",
			source: `package main

func Header() {
	Div(Class("header")) {
		H1 {
			"My App"
		}
	}
}`,
			contains: []string{
				"type Header struct",
				"func NewHeader() *Header",
				"func (c *Header) Render() *runtime.VNode",
				"runtime.Div(",
				"runtime.Class(\"header\")",
				"runtime.H1(",
				"runtime.Text(\"My App\")",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Parse the source
			p, err := parser.New()
			if err != nil {
				t.Fatalf("Failed to create parser: %v", err)
			}

			file, err := p.Parse(strings.NewReader(tt.source))
			if err != nil {
				t.Fatalf("Failed to parse source: %v", err)
			}

			// Generate code
			gen := New("main")
			generated, err := gen.Generate(file)
			if err != nil {
				t.Fatalf("Failed to generate code: %v", err)
			}

			generatedStr := string(generated)

			// Check that all expected strings are present
			for _, expected := range tt.contains {
				if !strings.Contains(generatedStr, expected) {
					t.Errorf("Generated code does not contain expected string: %q\nGenerated code:\n%s", expected, generatedStr)
				}
			}
		})
	}
}

func TestGenerateChannelListener(t *testing.T) {
	source := `package main

func Watcher(eventChannel: chan string) {
	Div {
		` + "`Event: {<-eventChannel}`" + `
	}
}`

	p, err := parser.New()
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}

	file, err := p.Parse(strings.NewReader(source))
	if err != nil {
		t.Fatalf("Failed to parse source: %v", err)
	}

	gen := New("main")
	generated, err := gen.Generate(file)
	if err != nil {
		t.Fatalf("Failed to generate code: %v", err)
	}

	generatedStr := string(generated)

	// Verify that the channel listener is properly generated with val variable
	if !strings.Contains(generatedStr, "for val := range c.EventChannel {") {
		t.Error("Generated code should contain 'for val := range c.EventChannel {'")
	}

	// Verify that val is used in the body
	if !strings.Contains(generatedStr, "c.currentEventChannel = val") {
		t.Error("Generated code should contain 'c.currentEventChannel = val'")
	}

	// Verify that Update is called
	if !strings.Contains(generatedStr, "c.app.Update()") {
		t.Error("Generated code should contain 'c.app.Update()'")
	}
}

func TestGenerateMultipleComponents(t *testing.T) {
	source := `package main

func Header(title: string) {
	H1 {
		` + "`{title}`" + `
	}
}

func Footer(copyright: string) {
	Div(Class("footer")) {
		P {
			` + "`{copyright}`" + `
		}
	}
}`

	p, err := parser.New()
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}

	file, err := p.Parse(strings.NewReader(source))
	if err != nil {
		t.Fatalf("Failed to parse source: %v", err)
	}

	gen := New("main")
	generated, err := gen.Generate(file)
	if err != nil {
		t.Fatalf("Failed to generate code: %v", err)
	}

	generatedStr := string(generated)

	// Check both components are generated
	expectedStructs := []string{
		"type Header struct",
		"type Footer struct",
		"func NewHeader",
		"func NewFooter",
		"func (c *Header) Render()",
		"func (c *Footer) Render()",
	}

	for _, expected := range expectedStructs {
		if !strings.Contains(generatedStr, expected) {
			t.Errorf("Generated code does not contain expected string: %q", expected)
		}
	}
}

func TestGenerateComponentWithMultipleChannelsAndParameters(t *testing.T) {
	source := `package main

func ComplexWidget(
	title: string,
	dataStream: chan string,
	count: int,
	statusStream: chan int,
	enabled: bool
) {
	Div(Class("widget")) {
		H1 {
			` + "`{title}`" + `
		}
		Div {
			` + "`Data: {<-dataStream}`" + `
		}
		Div {
			` + "`Count: {count}`" + `
		}
		Div {
			` + "`Status: {<-statusStream}`" + `
		}
		Div {
			` + "`Enabled: {enabled}`" + `
		}
	}
}`

	p, err := parser.New()
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}

	file, err := p.Parse(strings.NewReader(source))
	if err != nil {
		t.Fatalf("Failed to parse source: %v", err)
	}

	gen := New("main")
	generated, err := gen.Generate(file)
	if err != nil {
		t.Fatalf("Failed to generate code: %v", err)
	}

	generatedStr := string(generated)

	// Verify all parameters are in the struct
	expectedFields := []string{
		"Title",
		"DataStream",
		"chan string",
		"Count",
		"StatusStream",
		"chan int",
		"Enabled",
		"currentDataStream",
		"currentStatusStream",
	}

	for _, expected := range expectedFields {
		if !strings.Contains(generatedStr, expected) {
			t.Errorf("Generated code does not contain expected field: %q", expected)
		}
	}

	// Verify BindApp method exists and starts both channel listeners
	bindAppChecks := []string{
		"func (c *ComplexWidget) BindApp(app *runtime.App)",
		"if c.DataStream != nil {",
		"c.startDataStreamListener()",
		"if c.StatusStream != nil {",
		"c.startStatusStreamListener()",
	}

	for _, expected := range bindAppChecks {
		if !strings.Contains(generatedStr, expected) {
			t.Errorf("Generated code does not contain expected BindApp code: %q", expected)
		}
	}

	// Verify both listener methods exist
	listenerChecks := []string{
		"func (c *ComplexWidget) startDataStreamListener()",
		"for val := range c.DataStream {",
		"c.currentDataStream = val",
		"func (c *ComplexWidget) startStatusStreamListener()",
		"for val := range c.StatusStream {",
		"c.currentStatusStream = val",
	}

	for _, expected := range listenerChecks {
		if !strings.Contains(generatedStr, expected) {
			t.Errorf("Generated code does not contain expected listener code: %q", expected)
		}
	}

	// Verify all option functions are generated
	optionFuncs := []string{
		"func WithTitle(v string) ComplexWidgetOption",
		"func WithDataStream(v chan string) ComplexWidgetOption",
		"func WithCount(v int) ComplexWidgetOption",
		"func WithStatusStream(v chan int) ComplexWidgetOption",
		"func WithEnabled(v bool) ComplexWidgetOption",
	}

	for _, expected := range optionFuncs {
		if !strings.Contains(generatedStr, expected) {
			t.Errorf("Generated code does not contain expected option function: %q", expected)
		}
	}
}

func TestGenerateImports(t *testing.T) {
	source := `package main

func Simple() {
	Div {
		"Hello"
	}
}`

	p, err := parser.New()
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}

	file, err := p.Parse(strings.NewReader(source))
	if err != nil {
		t.Fatalf("Failed to parse source: %v", err)
	}

	gen := New("main")
	generated, err := gen.Generate(file)
	if err != nil {
		t.Fatalf("Failed to generate code: %v", err)
	}

	generatedStr := string(generated)

	// Verify required imports are present
	requiredImports := []string{
		`"syscall/js"`,
		`"fmt"`,
		`"github.com/gaarutyunov/guix/pkg/runtime"`,
	}

	for _, imp := range requiredImports {
		if !strings.Contains(generatedStr, imp) {
			t.Errorf("Generated code does not contain required import: %s", imp)
		}
	}

	// Verify code generation comment
	if !strings.Contains(generatedStr, "// Code generated by guix. DO NOT EDIT.") {
		t.Error("Generated code should contain code generation comment")
	}
}

func TestGenerateInterfaceMethods(t *testing.T) {
	source := `package main

func Widget(value: string) {
	Div {
		` + "`{value}`" + `
	}
}`

	p, err := parser.New()
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}

	file, err := p.Parse(strings.NewReader(source))
	if err != nil {
		t.Fatalf("Failed to parse source: %v", err)
	}

	gen := New("main")
	generated, err := gen.Generate(file)
	if err != nil {
		t.Fatalf("Failed to generate code: %v", err)
	}

	generatedStr := string(generated)

	// Verify all interface methods are generated
	interfaceMethods := []string{
		"func (c *Widget) Render() *runtime.VNode",
		"func (c *Widget) Mount(parent js.Value)",
		"func (c *Widget) Unmount()",
		"func (c *Widget) Update()",
	}

	for _, method := range interfaceMethods {
		if !strings.Contains(generatedStr, method) {
			t.Errorf("Generated code does not contain expected interface method: %q", method)
		}
	}
}

func TestGenerateMakeCall(t *testing.T) {
	source := `package main

func App() {
	counter := make(chan int, 10)

	Div {
		` + "`{counter}`" + `
	}
}`

	p, err := parser.New()
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}

	file, err := p.Parse(strings.NewReader(source))
	if err != nil {
		t.Fatalf("Failed to parse source: %v", err)
	}

	gen := New("main")
	generated, err := gen.Generate(file)
	if err != nil {
		t.Fatalf("Failed to generate code: %v", err)
	}

	generatedStr := string(generated)

	// Verify make() call is generated correctly
	expectedCode := []string{
		"make(chan int, 10)",
		"counter := make(chan int, 10)",
	}

	for _, expected := range expectedCode {
		if !strings.Contains(generatedStr, expected) {
			t.Errorf("Generated code does not contain expected code: %q\nGenerated:\n%s", expected, generatedStr)
		}
	}
}

func TestGenerateMakeCallWithoutSize(t *testing.T) {
	source := `package main

func Watcher() {
	events := make(chan string)

	Div {
		"Watching"
	}
}`

	p, err := parser.New()
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}

	file, err := p.Parse(strings.NewReader(source))
	if err != nil {
		t.Fatalf("Failed to parse source: %v", err)
	}

	gen := New("main")
	generated, err := gen.Generate(file)
	if err != nil {
		t.Fatalf("Failed to generate code: %v", err)
	}

	generatedStr := string(generated)

	// Verify make() call without size is generated correctly
	if !strings.Contains(generatedStr, "make(chan string)") {
		t.Errorf("Generated code does not contain 'make(chan string)'\nGenerated:\n%s", generatedStr)
	}

	// Should not have a second argument
	if strings.Contains(generatedStr, "make(chan string,") {
		t.Error("Generated code should not have size argument in make() call")
	}
}

func TestGenerateCompleteAppWithMake(t *testing.T) {
	source := `package main

func App() {
	dataChannel := make(chan string, 5)
	statusChannel := make(chan int)

	Div(Class("app")) {
		H1 {
			"App with Channels"
		}
		Span {
			` + "`Data: {<-dataChannel}`" + `
		}
		Span {
			` + "`Status: {<-statusChannel}`" + `
		}
	}
}`

	p, err := parser.New()
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}

	file, err := p.Parse(strings.NewReader(source))
	if err != nil {
		t.Fatalf("Failed to parse source: %v", err)
	}

	gen := New("main")
	generated, err := gen.Generate(file)
	if err != nil {
		t.Fatalf("Failed to generate code: %v", err)
	}

	generatedStr := string(generated)

	// Verify both make() calls are present
	expectedCode := []string{
		"dataChannel := make(chan string, 5)",
		"statusChannel := make(chan int)",
		"type App struct",
		"func NewApp() *App",
		"func (c *App) Render() *runtime.VNode",
	}

	for _, expected := range expectedCode {
		if !strings.Contains(generatedStr, expected) {
			t.Errorf("Generated code does not contain expected code: %q", expected)
		}
	}
}

func TestGenerateSelector(t *testing.T) {
	source := `package main

func Handler(e: Event) {
	value := e.Target.Value

	Div {
		` + "`{value}`" + `
	}
}`

	p, err := parser.New()
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}

	file, err := p.Parse(strings.NewReader(source))
	if err != nil {
		t.Fatalf("Failed to parse source: %v", err)
	}

	gen := New("main")
	generated, err := gen.Generate(file)
	if err != nil {
		t.Fatalf("Failed to generate code: %v", err)
	}

	generatedStr := string(generated)

	// Verify selector expression is generated correctly
	if !strings.Contains(generatedStr, "e.Target.Value") {
		t.Errorf("Generated code does not contain 'e.Target.Value'\nGenerated:\n%s", generatedStr)
	}

	// Verify it's in the right context (variable declaration)
	if !strings.Contains(generatedStr, "value := e.Target.Value") {
		t.Errorf("Generated code does not contain 'value := e.Target.Value'")
	}
}

func TestGenerateSelectorSingleField(t *testing.T) {
	source := `package main

func Widget(obj: Object) {
	name := obj.Name

	Div {
		` + "`{name}`" + `
	}
}`

	p, err := parser.New()
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}

	file, err := p.Parse(strings.NewReader(source))
	if err != nil {
		t.Fatalf("Failed to parse source: %v", err)
	}

	gen := New("main")
	generated, err := gen.Generate(file)
	if err != nil {
		t.Fatalf("Failed to generate code: %v", err)
	}

	generatedStr := string(generated)

	// Verify single-field selector is generated
	if !strings.Contains(generatedStr, "obj.Name") {
		t.Errorf("Generated code does not contain 'obj.Name'\nGenerated:\n%s", generatedStr)
	}
}

func TestGenerateSelectorInFunctionCall(t *testing.T) {
	source := `package main

func Logger(req: Request) {
	msg := req.User.Name

	Div {
		` + "`{msg}`" + `
	}
}`

	p, err := parser.New()
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}

	file, err := p.Parse(strings.NewReader(source))
	if err != nil {
		t.Fatalf("Failed to parse source: %v", err)
	}

	gen := New("main")
	generated, err := gen.Generate(file)
	if err != nil {
		t.Fatalf("Failed to generate code: %v", err)
	}

	generatedStr := string(generated)

	// Verify chained selector is generated
	if !strings.Contains(generatedStr, "req.User.Name") {
		t.Errorf("Generated code does not contain 'req.User.Name'\nGenerated:\n%s", generatedStr)
	}
}
