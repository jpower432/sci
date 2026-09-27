// SPDX-License-Identifier: Apache-2.0

package schema_test

import (
	"os"
	"os/exec"
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
	"bool": true, "any": true, "byte": true, "uint": true, "uint64": true,
}

// consumerDefinedTypes are the definitions this schema tags @go(-): the generator
// names them as field types without declaring them, so every Go consumer hand-writes
// them. go-gemara models them as typed ints with their own marshalling, which is why
// the schema suppresses a string declaration.
//
// The list is written out rather than derived, so that adding to it is a deliberate
// act. A new entry means every Go consumer fails to compile until it defines the type,
// and a release that adds one unannounced breaks them all.
var consumerDefinedTypes = map[string]bool{
	"ArtifactType": true, "AssessmentStep": true, "ComplianceStatus": true,
	"ConfidenceLevel": true, "Determination": true, "Disposition": true,
	"EnforcementStep": true, "EntityType": true, "EntryType": true,
	"EvidenceType": true, "FindingLifecycle": true, "GuidanceType": true,
	"Lifecycle": true, "MethodType": true, "ModType": true, "ModeType": true,
	"Opinion": true, "RelationshipType": true, "Result": true,
	"RiskAppetite": true, "Severity": true,
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

// valueTypedOptionalStructs are optional struct-valued fields that project as a Go
// value rather than a pointer. encoding/json ignores omitempty on a struct, so each
// one emits {} when absent — and {} fails validation wherever the struct has a
// required field, which every one of these does. They predate the chain redesign and
// are listed so the class cannot grow silently: fixing them means value to pointer,
// which is itself a source-breaking change for consumers, so it belongs in its own
// change rather than being smuggled into this one.
var valueTypedOptionalStructs = map[string]bool{
	"Actor.Contact": true, "Resource.Owner": true,
	"Policy.ImplementationPlan": true, "Policy.Risks": true,
	"Scope.Out": true, "AcceptedRisk.Scope": true, "Risk.Owner": true,
}

// TestGoProjection generates the Go types and checks the projection consumers compile
// against. The OpenAPI projection is gated heavily — golden file, schema count,
// oasdiff against v1 — and the Go projection was gated not at all, which is how three
// fields reached this branch typed `any` with a TODO. Each was governed by a rule
// stated on the field it forbade, which puts bottom in that field's type, so the
// generator cannot name it and every consumer loses the field.
func TestGoProjection(t *testing.T) {
	cueBin, err := exec.LookPath("cue")
	if err != nil {
		t.Skip("cue is not on PATH; CI installs it before the Go steps")
	}

	root, err := filepath.Abs("..")
	if err != nil {
		t.Fatalf("resolve schema directory: %v", err)
	}
	out := filepath.Join(t.TempDir(), "cue_types_gen.go")

	cmd := exec.Command(cueBin, "exp", "gengotypes", "--outfile", out, ".")
	cmd.Dir = root
	if combined, err := cmd.CombinedOutput(); err != nil {
		if strings.Contains(string(combined), "unknown command") {
			t.Skipf("this cue does not support gengotypes: %s", strings.TrimSpace(string(combined)))
		}
		t.Fatalf("gengotypes: %v\n%s", err, combined)
	}

	data, err := os.ReadFile(out)
	if err != nil {
		t.Fatalf("read generated types: %v", err)
	}
	generated := string(data)

	// A field the generator cannot type degrades to `any` with a TODO marker, which
	// takes the field away from every consumer while still validating in CUE.
	for _, line := range strings.Split(generated, "\n") {
		if strings.Contains(line, "IncompleteKind") || strings.Contains(line, "TODO") {
			t.Errorf("a field degraded in the Go projection: %s\n"+
				"\tthis is what a rule stated on the field it forbids does — state it on the field it requires instead",
				strings.TrimSpace(line))
		}
	}

	// `any` is legitimate only where the schema asks for it.
	pinnedAny := map[string]bool{}
	files, err := filepath.Glob(filepath.Join(root, "*.cue"))
	if err != nil {
		t.Fatalf("list CUE files: %v", err)
	}
	for _, f := range files {
		src, err := os.ReadFile(f)
		if err != nil {
			t.Fatalf("read %s: %v", f, err)
		}
		for _, m := range regexp.MustCompile(`@go\((\w+),[^)]*type=any`).FindAllStringSubmatch(string(src), -1) {
			pinnedAny[m[1]] = true
		}
	}
	for _, m := range regexp.MustCompile(`\n\t(\w+) any`).FindAllStringSubmatch(generated, -1) {
		if !pinnedAny[m[1]] {
			t.Errorf("field %s projects as any without @go(type=any) in the schema", m[1])
		}
	}

	// An optional struct-valued field must project as a pointer, or it cannot be
	// omitted: encoding/json writes {} and the document fails its own schema.
	structTypes := map[string]bool{}
	for _, m := range regexp.MustCompile(`\ntype (\w+) struct`).FindAllStringSubmatch(generated, -1) {
		structTypes[m[1]] = true
	}
	owner := ""
	for _, line := range strings.Split(generated, "\n") {
		if m := regexp.MustCompile(`^type (\w+) struct`).FindStringSubmatch(line); m != nil {
			owner = m[1]
		}
		m := regexp.MustCompile("^\\t(\\w+) (\\w+) `json:\"[^\"]+,omitempty\"").FindStringSubmatch(line)
		if m == nil || !structTypes[m[2]] {
			continue
		}
		if field := owner + "." + m[1]; !valueTypedOptionalStructs[field] {
			t.Errorf("%s is an optional %s and projects as a value, so it cannot be omitted:\n"+
				"\tencoding/json ignores omitempty on a struct, so a producer emits {} and the document fails validation —\n"+
				"\tadd optional=nillable to its @go attribute",
				field, m[2])
		}
	}

	// Every type named as a field type is either generated here or one a consumer
	// defines. A new one of the latter breaks every consumer's build.
	declared := map[string]bool{}
	for _, m := range regexp.MustCompile(`\ntype (\w+) `).FindAllStringSubmatch(generated, -1) {
		declared[m[1]] = true
	}
	for _, m := range regexp.MustCompile("\n\t\\w+ (?:\\[\\])?\\*?(\\w+)(?: `|\n)").FindAllStringSubmatch(generated, -1) {
		name := m[1]
		if declared[name] || goBuiltins[name] || consumerDefinedTypes[name] {
			continue
		}
		t.Errorf("field type %s is neither generated nor one consumers already define:\n"+
			"\tif it is a new enum, every Go consumer must hand-write it before this schema is usable —\n"+
			"\tadd it to consumerDefinedTypes to say that cost is intended, or drop @go(-) so it is generated",
			name)
	}
}
