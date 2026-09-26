// SPDX-License-Identifier: Apache-2.0

package schema_test

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

var (
	goAttrPattern = regexp.MustCompile(`@go\(([^)]*)\)`)
	defPattern    = regexp.MustCompile(`(?m)^#(\w+):`)
)

// goBuiltins are the Go types a @go(type=...) may name without a CUE definition
// behind them.
var goBuiltins = map[string]bool{
	"string": true, "int": true, "int64": true, "float64": true,
	"bool": true, "any": true, "byte": true,
}

// TestGoTypeAttributesNameRealDefinitions checks that every @go(type=...) names a
// definition that exists in the package. A @go attribute is an opaque string to
// CUE, so renaming a definition leaves any attribute mentioning it silently
// pointing at nothing, and the break only surfaces when go-gemara regenerates its
// types from this schema. Nothing else in the gate set reads these attributes.
func TestGoTypeAttributesNameRealDefinitions(t *testing.T) {
	root, err := filepath.Abs("..")
	if err != nil {
		t.Fatalf("resolve schema directory: %v", err)
	}

	files, err := filepath.Glob(filepath.Join(root, "*.cue"))
	if err != nil {
		t.Fatalf("list CUE files: %v", err)
	}
	if len(files) == 0 {
		t.Fatal("no CUE files found")
	}

	defined := map[string]bool{}
	sources := map[string]string{}
	for _, f := range files {
		data, err := os.ReadFile(f)
		if err != nil {
			t.Fatalf("read %s: %v", f, err)
		}
		sources[f] = string(data)
		for _, m := range defPattern.FindAllStringSubmatch(string(data), -1) {
			defined[m[1]] = true
		}
	}

	for f, src := range sources {
		for _, attr := range goAttrPattern.FindAllStringSubmatch(src, -1) {
			for _, part := range strings.Split(attr[1], ",") {
				value, found := strings.CutPrefix(strings.TrimSpace(part), "type=")
				if !found {
					continue
				}

				name := value
				name = strings.TrimPrefix(name, "[]")
				name = strings.TrimPrefix(name, "*")
				if name == "" || goBuiltins[name] {
					continue
				}

				if !defined[name] {
					t.Errorf("%s: @go(%s) names type %q, which is not a definition in this package",
						filepath.Base(f), attr[1], name)
				}
			}
		}
	}
}
