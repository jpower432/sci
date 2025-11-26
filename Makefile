all: tidy test testcov cuefmtcheck lintcue cuegen dirtycheck lintinsights

tidy:
	@echo "  >  Tidying go.mod ..."
	@go mod tidy
	@echo "  >  Tidying cue.mod ..."
	@cd schemas && cue mod tidy

test:
	@echo "  >  Running tests ..."
	@go vet ./...
	@go test ./...

testcov:
	@echo "Running tests and generating coverage output ..."
	@go test ./... -coverprofile coverage.out -covermode count
	@sleep 2 # Sleeping to allow for coverage.out file to get generated
	@echo "Current test coverage : $(shell go tool cover -func=coverage.out | grep total | grep -Eo '[0-9]+\.[0-9]+') %"

# Verify CUE formatting in ./schemas
cuefmtcheck:
	@echo "  >  Verifying CUE formatting in ./schemas ..."
	@cue fmt --check --files ./schemas

lint:
	@echo "  >  Linting Go files ..."
	@golangci-lint run

lintcue:
	@echo "  >  Linting CUE files (with module support) ..."
	@cd schemas && cue eval ./layer1/layer1.cue --all-errors --verbose
	@cd schemas && cue eval ./layer2/layer2.cue --all-errors --verbose
	@cd schemas && cue eval ./layer3/layer3.cue --all-errors --verbose
	@cd schemas && cue eval ./layer4/layer4.cue --all-errors --verbose

cuegen: tidy
	@echo "  >  Generating types from cue schema ..."
	@cd schemas && cue exp gengotypes ./layer3/layer3.cue
	@mv schemas/cue_types_gen.go layer3/generated_types.go
	@cd schemas && cue exp gengotypes ./layer4/layer4.cue
	@mv schemas/cue_types_gen.go layer4/generated_types.go
	@mv schemas/common/cue_types_gen.go common/generated_types.go
	@mv schemas/layer1/cue_types_gen.go layer1/generated_types.go
	@mv schemas/layer2/cue_types_gen.go layer2/generated_types.go
	@go build -o utils/fix_generated_types utils/fix_generated_types.go
	@utils/fix_generated_types common/generated_types.go
	@utils/fix_generated_types layer1/generated_types.go
	@utils/fix_generated_types layer2/generated_types.go
	@utils/fix_generated_types layer3/generated_types.go
	@utils/fix_generated_types layer4/generated_types.go
	@rm utils/fix_generated_types

dirtycheck:
	@echo "  >  Checking for uncommitted changes ..."
	@if [ -n "$$(git status --porcelain)" ]; then \
		echo "  >  Uncommitted changes to generated files found!"; \
		echo "  >  Run make cuegen and commit the results."; \
		exit 1; \
	else \
		echo "  >  No uncommitted changes to generated files found."; \
	fi

oscalgenerate:
	@echo "  >  Generating OSCAL testdata from Gemara artifacts..."
	@mkdir -p artifacts
	@go run ./utils/oscal catalog ./layer2/test-data/good-osps.yml --output ./artifacts/catalog.json
	@go run ./utils/oscal  guidance ./layer1/test-data/good-aigf.yaml --catalog-output ./artifacts/guidance.json --profile-output ./artifacts/profile.json

lintinsights:
	@echo "  >  Linting security-insights.yml ..."
	@curl -O --silent https://raw.githubusercontent.com/ossf/security-insights-spec/refs/tags/v2.1.0/schema.cue
	@cue vet -d '#SecurityInsights' security-insights.yml schema.cue
	@rm schema.cue
	@echo "  >  Linting security-insights.yml complete."

PHONY: tidy test testcov lintcue cuegen dirtycheck lintinsights
