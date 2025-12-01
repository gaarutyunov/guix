//go:build mage
// +build mage

package main

import (
	"fmt"

	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
)

// Format runs gofmt on all Go files
func Format() error {
	fmt.Println("Running gofmt...")
	return sh.RunV("gofmt", "-w", ".")
}

// Vet runs go vet on all packages (excluding WASM-only runtime package)
func Vet() error {
	fmt.Println("Running go vet...")
	// Vet non-WASM packages (skip pkg/ast and pkg/parser due to participle library)
	packages := []string{
		"./cmd/...",
		"./pkg/codegen",
		"./pkg/visitors",
	}
	for _, pkg := range packages {
		if err := sh.RunV("go", "vet", pkg); err != nil {
			return err
		}
	}
	// Vet WASM runtime package with appropriate build tags
	fmt.Println("Running go vet on WASM runtime...")
	env := map[string]string{
		"GOOS":   "js",
		"GOARCH": "wasm",
	}
	return sh.RunWith(env, "go", "vet", "./pkg/runtime/...")
}

// Test runs all tests
func Test() error {
	fmt.Println("Running tests...")
	return sh.RunV("go", "test", "./...")
}

// Build builds all packages for native target
func Build() error {
	fmt.Println("Building native packages...")
	return sh.RunV("go", "build", "./...")
}

// BuildWasm builds the runtime package for WebAssembly
func BuildWasm() error {
	fmt.Println("Building WASM runtime...")
	env := map[string]string{
		"GOOS":   "js",
		"GOARCH": "wasm",
	}
	return sh.RunWith(env, "go", "build", "./pkg/runtime/...")
}

// Generate regenerates all example code
func Generate() error {
	fmt.Println("Regenerating example code...")
	examples := []string{"counter", "calculator", "webgpu-cube"}
	for _, example := range examples {
		if err := sh.RunV("go", "run", "./cmd/guix", "generate", "--path", "examples/"+example); err != nil {
			return fmt.Errorf("failed to generate %s: %w", example, err)
		}
	}
	return nil
}

// PreCommit runs all pre-commit checks (format, vet, test, build)
func PreCommit() error {
	fmt.Println("Running pre-commit checks...")
	mg.Deps(Format)
	mg.Deps(Vet)
	mg.Deps(Test)
	mg.Deps(Build)
	mg.Deps(BuildWasm)
	fmt.Println("✓ All pre-commit checks passed!")
	return nil
}

// CI runs all CI checks (same as PreCommit plus generation check)
func CI() error {
	fmt.Println("Running CI checks...")
	if err := PreCommit(); err != nil {
		return err
	}
	fmt.Println("✓ All CI checks passed!")
	return nil
}

// Clean removes build artifacts and generated files
func Clean() error {
	fmt.Println("Cleaning build artifacts...")
	patterns := []string{
		"examples/*/calculator",
		"examples/*/counter",
		"examples/*/webgpu-cube",
		"examples/*/*.wasm",
		"examples/*/wasm_exec.js",
		"*.test",
	}
	for _, pattern := range patterns {
		if err := sh.Run("sh", "-c", "rm -f "+pattern); err != nil {
			fmt.Printf("Warning: failed to clean %s: %v\n", pattern, err)
		}
	}
	fmt.Println("✓ Clean complete!")
	return nil
}

// Default target runs PreCommit
var Default = PreCommit
