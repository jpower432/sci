// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/goccy/go-yaml"
)

func TestConvertCUEToOpenAPI(t *testing.T) {
	out := filepath.Join(t.TempDir(), "openapi.yaml")
	if err := convertCUEToOpenAPI("../../..", out, ConvertOpts{}); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}
	var spec struct {
		Components struct {
			Schemas map[string]SchemaInfo `yaml:"schemas"`
		} `yaml:"components"`
	}
	if err := yaml.Unmarshal(data, &spec); err != nil {
		t.Fatal(err)
	}
	schemas := spec.Components.Schemas

	// Disjunctions of string literals emit enum values.
	if got := schemas["MethodType"].Enum; len(got) != 4 {
		t.Errorf("MethodType enum = %v, want 4 values", got)
	}
	if got := schemas["Lifecycle"].Enum; len(got) != 4 { // includes the *"Active" default
		t.Errorf("Lifecycle enum = %v, want 4 values", got)
	}

	// Hidden definitions and hidden fields don't leak into the output.
	for name, s := range schemas {
		if strings.HasPrefix(name, "_") {
			t.Errorf("hidden definition %s leaked into schemas", name)
		}
		for prop := range s.Properties {
			if strings.HasPrefix(prop, "_") {
				t.Errorf("hidden field %s.%s leaked into properties", name, prop)
			}
		}
	}

	// Refs to hidden definitions resolve to their visible base type.
	if !strings.Contains(string(data), "schemas/Mapping") || strings.Contains(string(data), "schemas/_") {
		t.Errorf("refs to hidden definitions not rewritten to base types")
	}
}
