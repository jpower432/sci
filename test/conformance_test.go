// SPDX-License-Identifier: Apache-2.0

package schema_test

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"cuelang.org/go/cue"
	cueyaml "cuelang.org/go/encoding/yaml"
)

// TestChainConformance implements the cross-document rules from docs/tool-spec.md
// against examples/osps-level-2, which is a complete chain: the policy, the scan
// that evaluated it, and the audit conducted against both. Those rules cannot live
// in CUE — validation sees one document at a time — so without something like this
// they are prose that nothing checks.
//
// Cadence and freshness are deliberately not checked here. Both need a clock, and a
// test that reads the wall clock either rots or goes green for the wrong reason; the
// tool spec states them as obligations on a tool that has one.
func TestChainConformance(t *testing.T) {
	catalog := loadYAML(t, "./test-data/good-osps.yml")
	policy := loadYAML(t, "../examples/osps-level-2/policy.yaml")
	evalLog := loadYAML(t, "../examples/osps-level-2/evaluation-log.yaml")
	audit := loadYAML(t, "../examples/osps-level-2/audit-log.yaml")

	// The catalog's requirements, and the subset the policy governs.
	catalogReqs := map[string]cue.Value{}
	for _, c := range listAt(catalog, "controls") {
		for _, r := range listAt(c, `"assessment-requirements"`) {
			catalogReqs[strAt(r, "id")] = r
		}
	}
	excluded := map[string]bool{}
	for _, imp := range listAt(policy, "imports.catalogs") {
		for _, e := range stringsAt(imp, "exclusions") {
			excluded[e] = true
		}
	}
	governed := map[string]bool{}
	for id := range catalogReqs {
		if !excluded[id] {
			governed[id] = true
		}
	}

	// The policy's plans, by id, and the methods each accepts.
	type plan struct {
		requirement string
		methods     map[string]cue.Value
	}
	plans := map[string]plan{}
	for _, p := range listAt(policy, `adherence."assessment-plans"`) {
		methods := map[string]cue.Value{}
		for _, m := range listAt(p, `"evaluation-methods"`) {
			methods[strAt(m, "id")] = m
		}
		plans[strAt(p, "id")] = plan{requirement: strAt(p, `requirement."entry-id"`), methods: methods}
	}
	if len(plans) == 0 {
		t.Fatal("the policy declares no assessment plans; the rest of this test would pass vacuously")
	}

	// Every assessment in the scan, keyed by the requirement it assessed. An
	// assessment log has no id of its own, so a citation resolves by requirement.
	assessments := map[string][]cue.Value{}
	for _, e := range listAt(evalLog, "evaluations") {
		for _, a := range listAt(e, `"assessment-logs"`) {
			rid := strAt(a, `requirement."entry-id"`)
			assessments[rid] = append(assessments[rid], a)
		}
	}

	verifications := listAt(audit, "verifications")

	t.Run("rule 6: the audit covers exactly the requirements the policy governs", func(t *testing.T) {
		verified := map[string]bool{}
		for _, v := range verifications {
			verified[strAt(v, `requirement."entry-id"`)] = true
		}
		for id := range governed {
			if !verified[id] {
				t.Errorf("requirement %s is governed by the policy and has no verification entry", id)
			}
		}
		for id := range verified {
			if !governed[id] {
				t.Errorf("requirement %s has a verification entry but the policy does not govern it", id)
			}
		}
	})

	t.Run("rule 7: the effective requirement is derivable from the catalog", func(t *testing.T) {
		for _, v := range verifications {
			rid := strAt(v, `requirement."entry-id"`)
			r, ok := catalogReqs[rid]
			if !ok {
				continue // reported by rule 6
			}
			if got, want := collapse(strAt(v, "effective.text")), collapse(strAt(r, "text")); got != want {
				t.Errorf("%s: effective text is not the catalog's\n  audit:   %s\n  catalog: %s", rid, got, want)
			}
			if got, want := stringsAt(v, "effective.applicability"), stringsAt(r, "applicability"); !equal(got, want) {
				t.Errorf("%s: effective applicability %v is not the catalog's %v", rid, got, want)
			}
		}
	})

	t.Run("rule 8: a named plan is the policy's and governs that requirement", func(t *testing.T) {
		named := 0
		for _, v := range verifications {
			pid := strAt(v, `effective."plan-id"`)
			if pid == "" {
				continue
			}
			named++
			rid := strAt(v, `requirement."entry-id"`)
			p, ok := plans[pid]
			if !ok {
				t.Errorf("%s: names plan %s, which the policy does not declare", rid, pid)
				continue
			}
			if p.requirement != rid {
				t.Errorf("%s: names plan %s, which governs %s", rid, pid, p.requirement)
			}
		}
		if named == 0 {
			t.Error("no verification entry names a plan; rule 8 passed vacuously")
		}
	})

	t.Run("rule 2: every method named is one its plan accepts", func(t *testing.T) {
		// in the audit's plan verification
		for _, v := range verifications {
			pid := strAt(v, `effective."plan-id"`)
			p, ok := plans[pid]
			if !ok {
				continue
			}
			for _, m := range listAt(v, "plan.methods") {
				mid := strAt(m, `"method-id"`)
				if _, ok := p.methods[mid]; !ok {
					t.Errorf("%s: verifies method %s, which plan %s does not accept",
						strAt(v, `requirement."entry-id"`), mid, pid)
				}
			}
		}
		// and in the scan that ran them
		for rid, as := range assessments {
			for _, a := range as {
				pid := strAt(a, `plan."entry-id"`)
				if pid == "" {
					continue
				}
				p, ok := plans[pid]
				if !ok {
					t.Errorf("%s: assessment cites plan %s, which the policy does not declare", rid, pid)
					continue
				}
				mid := strAt(a, `execution."method-id"`)
				if _, ok := p.methods[mid]; !ok {
					t.Errorf("%s: assessment ran method %s, which plan %s does not accept", rid, mid, pid)
				}
			}
		}
	})

	t.Run("rule 1: an assessment's plan governs the requirement it assessed", func(t *testing.T) {
		for rid, as := range assessments {
			for _, a := range as {
				pid := strAt(a, `plan."entry-id"`)
				if p, ok := plans[pid]; ok && p.requirement != rid {
					t.Errorf("%s: assessed under plan %s, which governs %s", rid, pid, p.requirement)
				}
			}
		}
	})

	t.Run("rule 3: the executor is the one the plan named", func(t *testing.T) {
		checked := 0
		for rid, as := range assessments {
			for _, a := range as {
				p, ok := plans[strAt(a, `plan."entry-id"`)]
				if !ok {
					continue
				}
				m, ok := p.methods[strAt(a, `execution."method-id"`)]
				if !ok {
					continue
				}
				want := strAt(m, "executor.id")
				if want == "" {
					continue // the plan names no executor, so anything may run it
				}
				checked++
				if got := strAt(a, `execution.executor.id`); got != want {
					t.Errorf("%s: ran as executor %q where the plan names %q", rid, got, want)
				}
			}
		}
		if checked == 0 {
			t.Error("no plan method names an executor; rule 3 passed vacuously")
		}
	})

	t.Run("rule 4: parameter values are ones the plan declares", func(t *testing.T) {
		checked := 0
		for rid, as := range assessments {
			for _, a := range as {
				p, ok := plans[strAt(a, `plan."entry-id"`)]
				if !ok {
					continue
				}
				declared := map[string][]string{}
				for _, decl := range listAt(policyPlan(policy, strAt(a, `plan."entry-id"`)), "parameters") {
					declared[strAt(decl, "id")] = stringsAt(decl, `"accepted-values"`)
				}
				_ = p
				for _, pv := range listAt(a, "execution.parameters") {
					id, value := strAt(pv, "id"), strAt(pv, "value")
					accepted, ok := declared[id]
					if !ok {
						t.Errorf("%s: used parameter %s, which its plan does not declare", rid, id)
						continue
					}
					checked++
					if len(accepted) > 0 && !contains(accepted, value) {
						t.Errorf("%s: parameter %s used %q, not among %v", rid, id, value, accepted)
					}
				}
			}
		}
		if checked == 0 {
			t.Error("no assessment records a parameter value; rule 4 passed vacuously")
		}
	})

	t.Run("rule 9: citations resolve to assessments that exist", func(t *testing.T) {
		cited := 0
		for _, v := range verifications {
			rid := strAt(v, `requirement."entry-id"`)
			for _, c := range listAt(v, "assessments") {
				cited++
				entry := strAt(c, `"entry-id"`)
				as, ok := assessments[entry]
				if !ok {
					t.Errorf("%s: cites assessment %q, which the scan does not contain", rid, entry)
					continue
				}
				if entry != rid {
					t.Errorf("%s: cites an assessment of %s", rid, entry)
				}
				// An assessment log has no id, so entry-id names the requirement. Where
				// several assessments share it — a plan requiring more than one method —
				// the citation must say which, or it does not identify anything.
				coord := strAt(c, "coordinate")
				if len(as) > 1 && coord == "" {
					t.Errorf("%s: %d assessments share this requirement and the citation names no coordinate",
						rid, len(as))
					continue
				}
				if coord == "" {
					continue
				}
				want, ok := strings.CutPrefix(coord, "execution.method-id=")
				if !ok {
					continue // some other reader hint
				}
				found := false
				for _, a := range as {
					if strAt(a, `execution."method-id"`) == want {
						found = true
					}
				}
				if !found {
					t.Errorf("%s: coordinate names method %s, which no assessment of it ran", rid, want)
				}
			}
		}
		if cited == 0 {
			t.Error("the audit cites no assessments; rule 9 passed vacuously")
		}
	})

	t.Run("findings cite evaluation entries that exist", func(t *testing.T) {
		for _, f := range listAt(audit, "findings") {
			entry := strAt(f, `log."entry-id"`)
			if entry == "" {
				continue
			}
			if _, ok := assessments[entry]; !ok {
				t.Errorf("finding %s cites evaluation entry %q, which the scan does not contain",
					strAt(f, "id"), entry)
			}
		}
	})

	t.Run("an outcome a reader would question is explained by a finding", func(t *testing.T) {
		explained := map[string]bool{}
		for _, f := range listAt(audit, "findings") {
			if r := strAt(f, `requirement."entry-id"`); r != "" {
				explained[r] = true
			}
		}
		for _, v := range verifications {
			outcome := strAt(v, "outcome")
			rid := strAt(v, `requirement."entry-id"`)
			if (outcome == "Not Compliant" || outcome == "Undetermined") && !explained[rid] {
				t.Errorf("%s is %s and no finding names it", rid, outcome)
			}
		}
	})

	t.Run("an outcome is consistent with the checks beneath it", func(t *testing.T) {
		for _, v := range verifications {
			rid := strAt(v, `requirement."entry-id"`)
			checks := []string{strAt(v, `"evidence-present"`), strAt(v, `"evidence-fresh"`), strAt(v, "plan.bound")}
			for _, m := range listAt(v, "plan.methods") {
				for _, k := range []string{`allowed`, `used`, `"cadence-met"`, `"executor-matched"`, `"parameters-matched"`} {
					checks = append(checks, strAt(m, k))
				}
			}
			if strAt(v, "outcome") == "Compliant" && contains(checks, "Not Satisfied") {
				t.Errorf("%s is Compliant with a check that came out Not Satisfied", rid)
			}
		}
	})

	t.Run("the scan's aggregate follows Result precedence", func(t *testing.T) {
		order := []string{"Failed", "Unknown", "Needs Review", "Passed", "Not Applicable"}
		var seen []string
		for _, e := range listAt(evalLog, "evaluations") {
			var inner []string
			for _, a := range listAt(e, `"assessment-logs"`) {
				inner = append(inner, strAt(a, "result"))
			}
			if got, want := strAt(e, "result"), strongest(order, inner); got != want {
				t.Errorf("control evaluation %s reports %s where its assessments roll up to %s",
					strAt(e, `control."entry-id"`), got, want)
			}
			seen = append(seen, inner...)
		}
		if got, want := strAt(evalLog, "result"), strongest(order, seen); got != want {
			t.Errorf("the log reports %s where its assessments roll up to %s", got, want)
		}
	})
}

func loadYAML(t *testing.T, path string) cue.Value {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	f, err := cueyaml.Extract(path, data)
	if err != nil {
		t.Fatalf("parse %s: %v", path, err)
	}
	v := schemaCtx.BuildFile(f)
	if v.Err() != nil {
		t.Fatalf("build %s: %v", path, v.Err())
	}
	return v
}

func policyPlan(policy cue.Value, id string) cue.Value {
	for _, p := range listAt(policy, `adherence."assessment-plans"`) {
		if strAt(p, "id") == id {
			return p
		}
	}
	return cue.Value{}
}

func strAt(v cue.Value, path string) string {
	s, err := v.LookupPath(cue.ParsePath(path)).String()
	if err != nil {
		return ""
	}
	return s
}

func listAt(v cue.Value, path string) []cue.Value {
	it, err := v.LookupPath(cue.ParsePath(path)).List()
	if err != nil {
		return nil
	}
	var out []cue.Value
	for it.Next() {
		out = append(out, it.Value())
	}
	return out
}

func stringsAt(v cue.Value, path string) []string {
	var out []string
	it, err := v.LookupPath(cue.ParsePath(path)).List()
	if err != nil {
		return nil
	}
	for it.Next() {
		s, err := it.Value().String()
		if err != nil {
			return nil
		}
		out = append(out, s)
	}
	return out
}

func collapse(s string) string { return strings.Join(strings.Fields(s), " ") }

func equal(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func contains(haystack []string, needle string) bool {
	for _, h := range haystack {
		if h == needle {
			return true
		}
	}
	return false
}

// strongest returns the first value of order that appears in got, which is how
// every Gemara roll-up aggregates.
func strongest(order, got []string) string {
	for _, want := range order {
		if contains(got, want) {
			return want
		}
	}
	return fmt.Sprintf("no value of %v", order)
}
