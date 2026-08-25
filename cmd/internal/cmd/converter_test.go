// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"cuelang.org/go/cue/parser"
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

	// Disjunctions of string literals emit enum values. Expected values come
	// from #MethodType (policy.cue) and #Lifecycle (collections.cue); update
	// here when those CUE enums change.
	if got := schemas["MethodType"].Enum; !slices.Equal(got, []string{"Behavioral", "Intent", "Remediation", "Gate"}) {
		t.Errorf("MethodType enum = %v", got)
	}
	// "Active" is the *default in CUE; it appears here as a plain enum value.
	if got := schemas["Lifecycle"].Enum; !slices.Equal(got, []string{"Active", "Draft", "Deprecated", "Retired"}) {
		t.Errorf("Lifecycle enum = %v", got)
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

	// Refs to hidden definitions resolve to their visible base type
	// (#_MappingStrict -> Mapping, mappingdocument.cue).
	mappings, ok := schemas["MappingDocument"].Properties["mappings"].(map[string]interface{})
	if !ok {
		t.Fatalf("MappingDocument.mappings missing or not a map: %v", schemas["MappingDocument"].Properties["mappings"])
	}
	items, _ := mappings["items"].(map[string]interface{})
	if ref, _ := items["$ref"].(string); ref != "#/components/schemas/Mapping" {
		t.Errorf("MappingDocument.mappings items $ref = %q, want #/components/schemas/Mapping", ref)
	}
	if strings.Contains(string(data), "schemas/_") {
		t.Errorf("ref to a hidden definition leaked into output")
	}
}

func TestCollectEnumStrings(t *testing.T) {
	cases := []struct {
		expr string
		want []string // nil: expect ok=false
	}{
		{`"a" | "b" | "c"`, []string{"a", "b", "c"}},
		{`*"a" | "b"`, []string{"a", "b"}},
		{`"a" | *"b" | "c"`, []string{"a", "b", "c"}},
		{`"a" | #Other`, nil},
		{`int | string`, nil},
		{`"a" & "b"`, nil},
	}
	for _, tc := range cases {
		expr, err := parser.ParseExpr("test", tc.expr)
		if err != nil {
			t.Fatalf("%s: %v", tc.expr, err)
		}
		got, ok := collectEnumStrings(expr)
		if tc.want == nil {
			if ok {
				t.Errorf("%s: expected ok=false, got %v", tc.expr, got)
			}
		} else if !ok || !slices.Equal(got, tc.want) {
			t.Errorf("%s: got %v (ok=%v), want %v", tc.expr, got, ok, tc.want)
		}
	}
}
