// SPDX-License-Identifier: Apache-2.0

package schema_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/cuecontext"
	"cuelang.org/go/cue/load"
	cuejson "cuelang.org/go/encoding/json"
	cueyaml "cuelang.org/go/encoding/yaml"
)

var schemaValue cue.Value
var schemaCtx *cue.Context

func TestMain(m *testing.M) {
	schemaCtx = cuecontext.New()
	ctx := schemaCtx

	schemaDir, err := filepath.Abs("..")
	if err != nil {
		panic("failed to resolve schema directory: " + err.Error())
	}

	cfg := &load.Config{
		Dir: schemaDir,
	}
	instances := load.Instances([]string{"."}, cfg)
	if len(instances) != 1 {
		panic("expected exactly one CUE instance")
	}

	schemaValue = ctx.BuildInstance(instances[0])
	if schemaValue.Err() != nil {
		panic("failed to build CUE schema: " + schemaValue.Err().Error())
	}

	os.Exit(m.Run())
}

func TestSchemaValidation(t *testing.T) {
	tests := []struct {
		name        string
		file        string
		definition  string
		wantErr     bool
		errContains string
	}{
		// ControlCatalog — positive
		{"valid control catalog YAML", "./test-data/good-ccc.yaml", "#ControlCatalog", false, ""},
		{"valid control catalog JSON", "./test-data/good-ccc.json", "#ControlCatalog", false, ""},
		{"valid OSPS baseline", "./test-data/good-osps.yml", "#ControlCatalog", false, ""},
		{"valid lifecycle catalog", "./test-data/good-lifecycle.yaml", "#ControlCatalog", false, ""},
		{"valid nested control catalog", "./test-data/nested-good-ccc.yaml", "#ControlCatalog", false, ""},
		{"control catalog extending a declared reference", "./test-data/good-control-catalog-extends.yaml", "#ControlCatalog", false, ""},

		// GuidanceCatalog — positive
		{"valid AI governance framework", "./test-data/good-aigf.yaml", "#GuidanceCatalog", false, ""},
		// PrinciplesCatalog — positive
		{"valid AIGF principles catalog", "./test-data/good-aigf-principles.yaml", "#PrincipleCatalog", false, ""},

		// VectorCatalog — positive
		{"valid AIGF vector catalog", "./test-data/good-aigf-vectors.yaml", "#VectorCatalog", false, ""},
		{"threats with vectors", "./test-data/good-threat-catalog.yaml", "#ThreatCatalog", false, ""},
		{"valid capability catalog", "./test-data/good-capability-catalog.yaml", "#CapabilityCatalog", false, ""},
		{"vector mapping", "./test-data/good-vector-owasp-mapping.yaml", "#MappingDocument", false, ""},

		// AI agent capability catalog and ATR mappings (authored by ATR, validated against Gemara)
		{"valid AI agent capability catalog", "../examples/ai-agent/ai-agent-capability-catalog.yaml", "#CapabilityCatalog", false, ""},
		{"valid ATR categories to capabilities mapping", "../examples/ai-agent/atr-categories-to-capabilities-mapping.yaml", "#MappingDocument", false, ""},

		// A complete audit at the scale of a real baseline, with the policy it was
		// conducted against (examples/osps-level-2)
		{"OSPS Level 2 adherence policy", "../examples/osps-level-2/policy.yaml", "#Policy", false, ""},
		{"scan behind the OSPS baseline audit", "../examples/osps-level-2/evaluation-log.yaml", "#EvaluationLog", false, ""},
		{"complete audit of the OSPS baseline", "../examples/osps-level-2/audit-log.yaml", "#AuditLog", false, ""},
		{"gate actions taken after the OSPS scan", "../examples/osps-level-2/enforcement-log.yaml", "#EnforcementLog", false, ""},

		// RiskCatalog — positive
		{"valid risk catalog", "./test-data/good-risk-catalog.yaml", "#RiskCatalog", false, ""},

		// RiskCatalog — negative
		{"risk catalog with duplicate rank", "./test-data/bad-risk-catalog-duplicate-rank.yaml", "#RiskCatalog", true, "_uniqueRiskRanks"},

		// Policy — positive
		{"valid policy", "./test-data/good-policy.yaml", "#Policy", false, ""},
		{"valid security policy", "./test-data/good-security-policy.yml", "#Policy", false, ""},

		// Policy — negative (identity of EntryMapping targets)
		{"policy with duplicate assessment plan ids", "./test-data/bad-policy-duplicate-plan-id.yaml", "#Policy", true, "_uniquePlanIds"},
		{"policy with method id colliding across sites", "./test-data/bad-policy-duplicate-method-id.yaml", "#Policy", true, "_uniqueMethodIds"},
		{"policy with duplicate parameter ids", "./test-data/bad-policy-duplicate-parameter-id.yaml", "#Policy", true, "_uniqueParameterIds"},
		{"policy binding two plans to one requirement", "./test-data/bad-policy-duplicate-plan-requirement.yaml", "#Policy", true, "_uniquePlanRequirements"},
		{"policy plan naming an undeclared requirement catalog", "./test-data/bad-policy-plan-undeclared-requirement.yaml", "#Policy", true, "assessment-plans-0-requirement"},

		// ControlCatalog — negative
		{"invalid YAML", "./test-data/bad.yaml", "#ControlCatalog", true, "metadata: conflicting values"},
		{"invalid JSON", "./test-data/bad.json", "#ControlCatalog", true, "field not allowed"},
		{"controls without groups", "./test-data/bad-no-groups.yaml", "#ControlCatalog", true, "_groupValidation"},
		{"control catalog without a version", "./test-data/bad-control-catalog-missing-version.yaml", "#ControlCatalog", true, "metadata.version"},
		{"control with duplicate assessment requirement ids", "./test-data/bad-control-catalog-duplicate-requirement-id.yaml", "#ControlCatalog", true, "_uniqueRequirementIds"},
		{"control catalog extending an undeclared reference", "./test-data/bad-control-catalog-undeclared-extends.yaml", "#ControlCatalog", true, "_refValidation"},

		// MappingDocument — positive
		{"valid mapping document", "./test-data/good-mapping-document.yaml", "#MappingDocument", false, ""},
		{"valid AIGF NIST 800-53 mapping", "./test-data/good-aigf-nist-mapping.yaml", "#MappingDocument", false, ""},

		// MappingDocument — negative
		{"invalid mapping document without mapping-references", "./test-data/bad-mapping-document.yaml", "#MappingDocument", true, "\"mapping-references\".0.id"},
		{"mapping missing targets for non-no-match relationship", "./test-data/bad-mapping-no-target.yaml", "#MappingDocument", true, "targets.0.\"entry-id\""},

		// Lexicon — positive
		{"valid lexicon", "./test-data/good-lexicon.yaml", "#Lexicon", false, ""},

		// Lexicon — negative
		{"lexicon with duplicate term ids", "./test-data/bad-lexicon-duplicate-term-id.yaml", "#Lexicon", true, "_uniqueTermIds"},

		// GuidanceCatalog — negative
		{"retired guideline with recommendations", "./test-data/bad-lifecycle.yaml", "#GuidanceCatalog", true, "a retired guideline recommends nothing"},

		// EvaluationLog — positive
		{"valid PVTR baseline scan", "./test-data/pvtr-baseline-scan.yaml", "#EvaluationLog", false, ""},
		{"assessments that never ran omit start", "./test-data/good-evaluation-log-unstarted.yaml", "#EvaluationLog", false, ""},

		// EvaluationLog — negative
		{"executed assessment missing start", "./test-data/bad-evaluation-log-missing-start.yaml", "#EvaluationLog", true, "\"assessment-logs\".0.start"},
		{"evaluation log citing an undeclared reference", "./test-data/bad-evaluation-log-undeclared-reference.yaml", "#EvaluationLog", true, "_refValidation"},
		{"evidence with an inline payload and a download-url", "./test-data/bad-evaluation-log-payload-with-download-url.yaml", "#EvaluationLog", true, "inline payload cannot also have a source download-url"},
		{"evidence source with a download-url and no digest", "./test-data/bad-evaluation-log-download-url-without-digest.yaml", "#EvaluationLog", true, "source.digest"},
		{"log evaluating the same control twice", "./test-data/bad-evaluation-log-duplicate-control.yaml", "#EvaluationLog", true, "_uniqueEvaluatedControls"},
		{"control evaluation assessing the same requirement twice", "./test-data/bad-evaluation-log-duplicate-requirement.yaml", "#EvaluationLog", true, "_uniqueAssessments"},

		// EnforcementLog — positive
		{"valid enforcement log", "./test-data/good-enforcement-log.yaml", "#EnforcementLog", false, ""},

		// EnforcementLog — negative
		{"enforcement action with invalid disposition", "./test-data/bad-enforcement-log.yaml", "#EnforcementLog", true, "disposition"},
		{"enforcement action missing log reference", "./test-data/bad-enforcement-missing-log.yaml", "#EnforcementLog", true, "log"},
		{"clear action carrying a justification", "./test-data/bad-enforcement-clear-failed.yaml", "#EnforcementLog", true, "nothing to justify"},
		{"enforced action whose justification justifies nothing", "./test-data/bad-enforcement-empty-justification.yaml", "#EnforcementLog", true, "justifies itself"},
		{"enforcement finding the gate never rated", "./test-data/bad-enforcement-finding-missing-severity.yaml", "#EnforcementLog", true, "findings.0.severity"},
		{"enforcement log with duplicate action ids", "./test-data/bad-enforcement-duplicate-action-id.yaml", "#EnforcementLog", true, "_uniqueActionIds"},
		{"enforcement log with duplicate finding ids across actions", "./test-data/bad-enforcement-duplicate-finding-id.yaml", "#EnforcementLog", true, "_uniqueFindingIds"},

		// AuditLog — positive
		{"valid audit log", "./test-data/good-audit-log.yaml", "#AuditLog", false, ""},
		{"audit log evidence mapping with both coordinate and entry-id", "./test-data/good-audit-log-coordinate-and-entry-id.yaml", "#AuditLog", false, ""},
		{"audit that passed with nothing to call out", "./test-data/good-audit-log-clean.yaml", "#AuditLog", false, ""},

		// AuditLog — negative
		{"audit log missing summary policy coverage and results", "./test-data/bad-audit-log.yaml", "#AuditLog", true, "summary: incomplete value"},
		{"audit log evidence source with invalid digest format", "./test-data/bad-audit-log-invalid-digest.yaml", "#AuditLog", true, "digest"},
		{"audit citing an undeclared policy", "./test-data/bad-audit-log-undeclared-policy.yaml", "#AuditLog", true, "_refValidation"},
		{"audit finding naming a requirement with no verification entry", "./test-data/bad-audit-log-finding-unknown-requirement.yaml", "#AuditLog", true, "_requirementValidation"},
		{"audit finding citing an undeclared risk reference", "./test-data/bad-audit-log-dangling-risk.yaml", "#AuditLog", true, "_refValidation"},
		{"verification entry omitting evidence-present for a requirement it audited", "./test-data/bad-audit-log-verification-missing-evidence-present.yaml", "#AuditLog", true, "verifications.0.outcome"},
		{"not compliant finding stating no severity", "./test-data/bad-audit-log-not-compliant-missing-severity.yaml", "#AuditLog", true, "findings.0.severity"},
		{"named plan with no record of how it was followed", "./test-data/bad-audit-log-plan-id-without-verification.yaml", "#AuditLog", true, "plan.bound"},
		{"plan verification naming no plan-id", "./test-data/bad-audit-log-plan-verification-without-plan-id.yaml", "#AuditLog", true, "effective.\"plan-id\""},
		{"method entry not saying whether the method was used", "./test-data/bad-audit-log-method-missing-used.yaml", "#AuditLog", true, "methods.0.used"},
		{"entry claiming evidence is present while naming none", "./test-data/bad-audit-log-evidence-present-without-evidence.yaml", "#AuditLog", true, "verifications.0.\"evidence-present\""},
		{"plan verification enumerating no methods", "./test-data/bad-audit-log-plan-no-methods.yaml", "#AuditLog", true, "plan.methods"},
		{"audit log evidence with neither payload nor source", "./test-data/bad-audit-log-evidence-neither.yaml", "#AuditLog", true, "evidence.0.source"},
		{"audit log target uri with no scheme", "./test-data/bad-audit-log-uri-no-scheme.yaml", "#AuditLog", true, "target.uri"},
		{"audit log mapping reference url with a non-alphabetic scheme", "./test-data/bad-audit-log-url-invalid-scheme.yaml", "#AuditLog", true, ".url: invalid value"},
		{"audit log mapping reference url with no scheme", "./test-data/bad-audit-log-url-no-scheme.yaml", "#AuditLog", true, ".url: invalid value"},

		// CapabilityCatalog — negative
		{"capability with invalid group", "./test-data/bad-capability-invalid-group.yaml", "#CapabilityCatalog", true, "_groupValidation"},

		// ThreatCatalog — negative
		{"threat with invalid group", "./test-data/bad-threat-invalid-group.yaml", "#ThreatCatalog", true, "_groupValidation"},

		// PrincipleCatalog — negative
		{"principle with invalid group", "./test-data/bad-principle-invalid-group.yaml", "#PrincipleCatalog", true, "_groupValidation"},

		// ControlCatalog — negative (group validation)
		{"control with invalid group", "./test-data/bad-control-invalid-group.yaml", "#ControlCatalog", true, "_groupValidation"},

		// ControlCatalog — edge cases
		{"empty nested catalog", "./test-data/nested-empty.yaml", "#ControlCatalog", false, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := os.ReadFile(tt.file)
			if err != nil {
				t.Fatalf("read %s: %v", tt.file, err)
			}

			def := schemaValue.LookupPath(cue.ParsePath(tt.definition))
			if def.Err() != nil {
				t.Fatalf("lookup %s: %v", tt.definition, def.Err())
			}

			var validationErr error
			switch {
			case strings.HasSuffix(tt.file, ".json"):
				validationErr = cuejson.Validate(data, def)
			case strings.HasSuffix(tt.file, ".yaml"), strings.HasSuffix(tt.file, ".yml"):
				validationErr = cueyaml.Validate(data, def)
			default:
				t.Fatalf("unsupported file extension: %s", tt.file)
			}

			if tt.wantErr && validationErr == nil {
				t.Error("expected validation error, got nil")
			}
			if !tt.wantErr && validationErr != nil {
				t.Errorf("unexpected validation error: %v", validationErr)
			}
			if tt.errContains != "" && validationErr != nil {
				if !strings.Contains(validationErr.Error(), tt.errContains) {
					t.Errorf("error %q does not contain %q", validationErr.Error(), tt.errContains)
				}
			}
		})
	}
}
