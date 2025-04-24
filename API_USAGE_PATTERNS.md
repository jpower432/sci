# Patterns of API usage

## Evaluation Tool

Applications that evaluate controls can choose to consume just the layer4 API as in the builder pattern below or both the layer2 and layer4 APIs as in the load and build pattern below.

### Builder pattern

In this pattern, the application creates functions that create `layer4.ControlEvaluation` instances for every control in a given framework/control catalog. `revanite-io/pvtr-github-repo` follows this pattern, where code like:

```go
func OSPS_AC_01() (evaluation *layer4.ControlEvaluation) {
	evaluation = &layer4.ControlEvaluation{
		Control_Id:        "OSPS-AC-01",
		Remediation_Guide: "",
	}

	evaluation.AddAssessment(
		"OSPS-AC-01.01",
		"When a user attempts to access a sensitive resource in the project's version control system, the system MUST require the user to complete a multi-factor authentication process.",
		[]string{
			"Maturity Level 1",
			"Maturity Level 2",
			"Maturity Level 3",
		},
		[]layer4.AssessmentMethod{
			orgRequiresMFA,
		},
	)

	return
}
```

directly encodes and duplicates the control catalog details into the application.

### Load and build pattern

In this pattern, the application uses the layer2 API to load a layer2.Catalog from reference data in YAML and then iterates over all layer2.ControlFamily instances and layer2.Control instances in the catalog.

The application provides a function that returns a layer4.Assessment with at least one layer4.AssessmentMethod for each layer2.Control.AssessmentRequirement
