package acctest_test

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestAcc_AllTestsHaveGenerateConfig verifies that every acceptance test
// that uses resource.ParallelTest or resource.Test has at least one TestStep
// with GenerateConfig enabled.
//
// This ensures import testing coverage across the test suite.
func TestAcc_AllTestsHaveGenerateConfig(t *testing.T) {
	// Find all _test.go files in the project
	// Start from the internal directory to cover all acceptance tests
	root := "../../internal"
	var testFiles []string

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(path, "_test.go") {
			basename := filepath.Base(path)
			if basename == "generate_config_test.go" {
				return nil // Skip this test file
			}
			testFiles = append(testFiles, path)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("failed to walk directory: %s", err)
	}

	// Track failures to report all at once
	var failures []string

	for _, path := range testFiles {
		content, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("failed to read %s: %s", path, err)
		}

		// Quick check: only check acceptance tests (files using acctest.NewTestTools)
		if !strings.Contains(string(content), "acctest.NewTestTools") {
			continue
		}

		// Quick check: if file doesn't contain resource.ParallelTest or resource.Test, skip
		if !strings.Contains(string(content), "resource.ParallelTest") &&
			!strings.Contains(string(content), "resource.Test") {
			continue
		}

		// Skip tests that don't need GenerateConfig:
		// - Data sources cannot be imported
		// - Action tests trigger operations, not importable resources
		// - Function tests test provider functions, not resources
		basename := filepath.Base(path)
		if strings.Contains(basename, "_data_source_test.go") ||
			strings.Contains(basename, "_source_test.go") ||
			strings.Contains(basename, "_action_test.go") {
			continue
		}

		// Skip function tests (in functions/ directory)
		if strings.Contains(path, "/functions/") {
			continue
		}

		// Parse the file
		fset := token.NewFileSet()
		file, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
		if err != nil {
			t.Fatalf("failed to parse %s: %s", path, err)
		}

		// Find all test functions
		for _, decl := range file.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok {
				continue
			}

			// Check if it's a test function
			if !isTestFunction(fn) {
				continue
			}

			testName := fn.Name.Name

			// Check if this test uses resource.ParallelTest or resource.Test
			hasTestCall := false
			hasGenerateConfig := false

			ast.Inspect(fn.Body, func(n ast.Node) bool {
				// Look for call expressions
				call, ok := n.(*ast.CallExpr)
				if !ok {
					return true
				}

				// Check if it's resource.ParallelTest or resource.Test
				sel, ok := call.Fun.(*ast.SelectorExpr)
				if !ok {
					return true
				}

				ident, ok := sel.X.(*ast.Ident)
				if !ok {
					return true
				}

				if ident.Name != "resource" {
					return true
				}

				if sel.Sel.Name == "ParallelTest" || sel.Sel.Name == "Test" {
					hasTestCall = true

					// Now look for TestStep structs in the arguments
					for _, arg := range call.Args {
						checkTestStepForGenerateConfig(arg, &hasGenerateConfig)
					}
				}

				return true
			})

			// Report if test uses resource.Test but no GenerateConfig
			if hasTestCall && !hasGenerateConfig {
				// Make path relative to project root
				relPath := strings.TrimPrefix(path, "../../")
				failures = append(failures, fmt.Sprintf("%s:%d - function %s", relPath, fset.Position(fn.Pos()).Line, testName))
			}
		}
	}

	if len(failures) > 0 {
		t.Errorf("Found %d tests without GenerateConfig enabled:\n%s", len(failures), strings.Join(failures, "\n"))
	}
}

// isTestFunction checks if a function is a test function (starts with "Test")
func isTestFunction(fn *ast.FuncDecl) bool {
	if fn.Recv != nil {
		return false // Skip methods
	}
	if fn.Type.Params == nil || fn.Type.Params.NumFields() == 0 {
		return false
	}
	// Check if function name starts with "Test"
	return strings.HasPrefix(fn.Name.Name, "Test")
}

// checkTestStepForGenerateConfig recursively checks if a TestStep has GenerateConfig enabled
func checkTestStepForGenerateConfig(n ast.Node, hasGenerateConfig *bool) {
	ast.Inspect(n, func(node ast.Node) bool {
		if node == nil || *hasGenerateConfig {
			return false
		}

		// Look for GenerateConfig: true
		kv, ok := node.(*ast.KeyValueExpr)
		if !ok {
			return true
		}

		key, ok := kv.Key.(*ast.Ident)
		if !ok {
			return true
		}

		if key.Name == "GenerateConfig" {
			// Check if value is true
			if ident, ok := kv.Value.(*ast.Ident); ok && ident.Name == "true" {
				*hasGenerateConfig = true
				return false
			}
		}

		return true
	})
}

// TestCheckTestStepForGenerateConfig tests the checkTestStepForGenerateConfig helper
func TestCheckTestStepForGenerateConfig(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		expected bool
	}{
		{
			name: "TestStep with GenerateConfig true",
			code: `package main

import "github.com/hashicorp/terraform-plugin-testing/helper/resource"

func TestExample(t *testing.T) {
	resource.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config:       "test",
				GenerateConfig: true,
			},
		},
	})
}`,
			expected: true,
		},
		{
			name: "TestStep without GenerateConfig",
			code: `package main

import "github.com/hashicorp/terraform-plugin-testing/helper/resource"

func TestExample(t *testing.T) {
	resource.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config: "test",
				Check:  nil,
			},
		},
	})
}`,
			expected: false,
		},
		{
			name: "TestStep with GenerateConfig false",
			code: `package main

import "github.com/hashicorp/terraform-plugin-testing/helper/resource"

func TestExample(t *testing.T) {
	resource.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config:       "test",
				GenerateConfig: false,
			},
		},
	})
}`,
			expected: false,
		},
		{
			name: "Multiple TestSteps with one having GenerateConfig",
			code: `package main

import "github.com/hashicorp/terraform-plugin-testing/helper/resource"

func TestExample(t *testing.T) {
	resource.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config: "test1",
			},
			{
				Config:       "test2",
				GenerateConfig: true,
			},
		},
	})
}`,
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fset := token.NewFileSet()
			file, err := parser.ParseFile(fset, "test.go", tt.code, parser.ParseComments)
			if err != nil {
				t.Fatalf("failed to parse code: %s", err)
			}

			hasGenerateConfig := false
			for _, decl := range file.Decls {
				fn, ok := decl.(*ast.FuncDecl)
				if !ok {
					continue
				}
				ast.Inspect(fn.Body, func(n ast.Node) bool {
					call, ok := n.(*ast.CallExpr)
					if !ok {
						return true
					}
					for _, arg := range call.Args {
						checkTestStepForGenerateConfig(arg, &hasGenerateConfig)
					}
					return true
				})
			}

			if hasGenerateConfig != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, hasGenerateConfig)
			}
		})
	}
}
