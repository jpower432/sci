// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"testing"

	"github.com/goccy/go-yaml"
)

// schemaRoot is the repo root (relative to this test's working directory,
// cmd/internal/cmd) where the top-level CUE schema files live.
const schemaRoot = "../../.."

// TestBreakingCheckIntegration exercises the full breaking-check pipeline end to
// end: it generates an OpenAPI projection of the current CUE schema, confirms an
// identical candidate reports no breaking change (exit 0), then structurally
// mutates the candidate to drop a required property and confirms the mutation is
// flagged as breaking (exit 1). It is skipped under -short because it invokes CUE
// generation.
func TestBreakingCheckIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test that invokes CUE generation in -short mode")
	}

	tmp := t.TempDir()
	baseYaml := filepath.Join(tmp, "base.yaml")
	if err := convertCUEToOpenAPI(schemaRoot, baseYaml, ConvertOpts{}); err != nil {
		t.Fatalf("convertCUEToOpenAPI(%q): %v", schemaRoot, err)
	}

	// Identical base and candidate: no breaking change.
	code, err := runBreakingCheck(breakingCheckOpts{basePath: baseYaml, candidatePath: baseYaml})
	if err != nil {
		t.Fatalf("runBreakingCheck (identical): %v", err)
	}
	if code != 0 {
		t.Fatalf("expected exit code 0 for identical candidate, got %d", code)
	}

	// Mutate the candidate by dropping a required entry from one schema. Removing
	// a required property is consumer-breaking (the response can no longer promise
	// it), so oasdiff must classify this as an ERR-level change.
	mutatedYaml := filepath.Join(tmp, "mutated.yaml")
	mutatedSchema, err := dropOneRequired(baseYaml, mutatedYaml)
	if err != nil {
		t.Fatalf("dropOneRequired: %v", err)
	}
	t.Logf("dropped a required property from schema %q", mutatedSchema)

	code, err = runBreakingCheck(breakingCheckOpts{basePath: baseYaml, candidatePath: mutatedYaml})
	if err != nil {
		t.Fatalf("runBreakingCheck (mutated): %v", err)
	}
	if code != 1 {
		t.Fatalf("expected exit code 1 for a dropped required property, got %d", code)
	}

	// Narrow an enum by dropping a member. This exercises the converter's enum
	// emission through the real generator: if the projection stopped emitting
	// enum constraints, the baseline would carry none, the diff would see no
	// enum to narrow, and the gate would silently miss the change.
	narrowedYaml := filepath.Join(tmp, "narrowed.yaml")
	narrowedSchema, err := dropOneEnumValue(baseYaml, narrowedYaml)
	if err != nil {
		t.Fatalf("dropOneEnumValue: %v", err)
	}
	t.Logf("dropped an enum value from schema %q", narrowedSchema)

	code, err = runBreakingCheck(breakingCheckOpts{basePath: baseYaml, candidatePath: narrowedYaml})
	if err != nil {
		t.Fatalf("runBreakingCheck (narrowed): %v", err)
	}
	if code != 1 {
		t.Fatalf("expected exit code 1 for a narrowed enum, got %d", code)
	}
}

// dropOneRequired reads the OpenAPI document at src, removes the first entry from
// the first component schema that declares a non-empty required list, writes the
// result to dst, and returns the name of the mutated schema. It fails if no such
// schema exists, so the test cannot silently pass on an empty diff.
func dropOneRequired(src, dst string) (string, error) {
	data, err := os.ReadFile(src)
	if err != nil {
		return "", fmt.Errorf("reading %q: %w", src, err)
	}
	var doc map[string]interface{}
	if err := yaml.Unmarshal(data, &doc); err != nil {
		return "", fmt.Errorf("unmarshaling %q: %w", src, err)
	}

	components, ok := doc["components"].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("no components map in %q", src)
	}
	schemas, ok := components["schemas"].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("no components.schemas map in %q", src)
	}

	// Iterate deterministically and skip experimental schemas: the gate exempts
	// them (x-status experimental → x-stability-level alpha), so a break there is
	// not flagged. The test must mutate a stable schema to exercise detection.
	names := make([]string, 0, len(schemas))
	for name := range schemas {
		names = append(names, name)
	}
	sort.Strings(names)

	for _, name := range names {
		schema, ok := schemas[name].(map[string]interface{})
		if !ok {
			continue
		}
		if status, _ := schema["x-status"].(string); status == "experimental" {
			continue
		}
		required, ok := schema["required"].([]interface{})
		if !ok || len(required) == 0 {
			continue
		}
		schema["required"] = required[1:]
		out, err := yaml.Marshal(doc)
		if err != nil {
			return "", fmt.Errorf("marshaling mutated doc: %w", err)
		}
		if err := os.WriteFile(dst, out, 0644); err != nil {
			return "", fmt.Errorf("writing %q: %w", dst, err)
		}
		return name, nil
	}
	return "", fmt.Errorf("no stable schema with a non-empty required list found in %q", src)
}

// dropOneEnumValue reads the OpenAPI document at src, removes the first member
// from the first stable component schema that declares a non-empty enum, writes
// the result to dst, and returns the name of the mutated schema. Narrowing an
// enum is producer-breaking (a value the API used to accept is now rejected). It
// fails if no such schema exists, so the test cannot silently pass on an empty
// diff — which is exactly what happens if the converter stops emitting enums.
func dropOneEnumValue(src, dst string) (string, error) {
	data, err := os.ReadFile(src)
	if err != nil {
		return "", fmt.Errorf("reading %q: %w", src, err)
	}
	var doc map[string]interface{}
	if err := yaml.Unmarshal(data, &doc); err != nil {
		return "", fmt.Errorf("unmarshaling %q: %w", src, err)
	}

	components, ok := doc["components"].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("no components map in %q", src)
	}
	schemas, ok := components["schemas"].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("no components.schemas map in %q", src)
	}

	names := make([]string, 0, len(schemas))
	for name := range schemas {
		names = append(names, name)
	}
	sort.Strings(names)

	for _, name := range names {
		schema, ok := schemas[name].(map[string]interface{})
		if !ok {
			continue
		}
		if status, _ := schema["x-status"].(string); status == "experimental" {
			continue
		}
		enum, ok := schema["enum"].([]interface{})
		if !ok || len(enum) == 0 {
			continue
		}
		schema["enum"] = enum[1:]
		out, err := yaml.Marshal(doc)
		if err != nil {
			return "", fmt.Errorf("marshaling mutated doc: %w", err)
		}
		if err := os.WriteFile(dst, out, 0644); err != nil {
			return "", fmt.Errorf("writing %q: %w", dst, err)
		}
		return name, nil
	}
	return "", fmt.Errorf("no stable schema with a non-empty enum found in %q", src)
}
