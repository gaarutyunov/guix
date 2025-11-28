// Package parser provides the Guix language parser using participle
package parser

import (
	"fmt"
	"io"

	"github.com/alecthomas/participle/v2"
	"github.com/alecthomas/participle/v2/lexer"
	"github.com/gaarutyunov/guix/pkg/ast"
)

// guixLexer defines the stateful lexer for Guix with template support
var guixLexer = lexer.MustStateful(lexer.Rules{
	"Root": {
		{"Comment", `//[^\n]*`, nil},
		{"Whitespace", `\s+`, nil},
		{"Keyword", `\b(package|import|type|struct|if|else|for|in|return|func|chan|true|false|make)\b`, nil},
		{"Op", `(<-|:=|\+=|-=|\*=|/=|==|!=|<=|>=|&&|\|\||[+\-*/<>&|!.=])`, nil},
		{"Ident", `[a-zA-Z_][a-zA-Z0-9_]*`, nil},
		{"Number", `\d+\.?\d*`, nil},
		{"String", `"(?:\\.|[^"\\])*"`, nil},
		{"Backtick", "`", lexer.Push("Template")},
		{"Punct", `[{}()\[\],;:]`, nil},
	},
	"Template": {
		{"BacktickEnd", "`", lexer.Pop()},
		{"ExprStart", `\{`, lexer.Push("TemplateExpr")},
		{"TemplateText", `[^{` + "`" + `]+`, nil},
	},
	"TemplateExpr": {
		{"ExprEnd", `\}`, lexer.Pop()},
		lexer.Include("Root"),
	},
})

// Parser is the Guix language parser
type Parser struct {
	parser *participle.Parser[ast.File]
}

// New creates a new Guix parser
func New() (*Parser, error) {
	p, err := participle.Build[ast.File](
		participle.Lexer(guixLexer),
		participle.Elide("Comment", "Whitespace"),
		participle.UseLookahead(10), // Increased for better disambiguation
	)
	if err != nil {
		return nil, fmt.Errorf("failed to build parser: %w", err)
	}

	return &Parser{parser: p}, nil
}

// Parse parses a Guix source file
func (p *Parser) Parse(r io.Reader) (*ast.File, error) {
	file, err := p.parser.Parse("", r)
	if err != nil {
		return nil, fmt.Errorf("parse error: %w", err)
	}
	return file, nil
}

// ParseString parses a Guix source string
func (p *Parser) ParseString(source string) (*ast.File, error) {
	file, err := p.parser.ParseString("", source)
	if err != nil {
		return nil, fmt.Errorf("parse error: %w", err)
	}
	return file, nil
}

// ParseBytes parses Guix source bytes
func (p *Parser) ParseBytes(filename string, source []byte) (*ast.File, error) {
	file, err := p.parser.ParseBytes(filename, source)
	if err != nil {
		return nil, fmt.Errorf("parse error in %s: %w", filename, err)
	}
	return file, nil
}

// Validate performs semantic validation on the parsed AST
func Validate(file *ast.File) error {
	// Basic validation - ensure components have bodies
	for _, comp := range file.Components {
		if comp.Body == nil {
			return fmt.Errorf("component %s at %s has no body", comp.Name, comp.Pos)
		}
	}
	return nil
}
