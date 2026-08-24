# SPDX-License-Identifier: Apache-2.0

all: tidy cuefmtcheck lintcue lintinsights test

#
# SCHEMA VALIDATION TESTS
#

test:
	@echo "  >  Running schema validation tests ..."
	@cd test && go test -v ./...
	@echo "  >  Schema validation tests complete."


#
# CUE DEVELOPMENT TOOLS
#

tidy:
	@echo "  >  Tidying cue.mod ..."
	@cue mod tidy

tidycheck: tidy
	@echo "  >  Checking CUE module tidiness ..."
	@if [ -n "$$(git status --porcelain cue.mod 2>/dev/null)" ]; then \
		echo "Error: cue.mod is not tidy. Please run 'make tidy' and commit the changes."; \
		git diff cue.mod; \
		exit 1; \
	fi
	@echo "  >  CUE module is tidy."

cuefmtcheck:
	@echo "  >  Verifying CUE formatting ..."
	@cue fmt --check --files .

lintcue:
	@echo "  >  Linting CUE files (with module support) ..."
	@cue eval . --all-errors --verbose

#
# SECURITY INSIGHTS VALIDATION
#
lintinsights:
	@echo "  >  Linting security-insights.yml ..."
	@curl -O --silent https://raw.githubusercontent.com/ossf/security-insights-spec/refs/tags/v2.1.0/schema.cue
	@cue vet -d '#SecurityInsights' security-insights.yml schema.cue
	@rm schema.cue
	@echo "  >  Linting security-insights.yml complete."

#
# GENERATE OPENAPI FROM SCHEMAS
#
# The website (github.com/gemaraproj/website) generates its reference pages
# by running the cmd/ CLI against a checkout of this repo; downstream type
# generation (e.g. gemara-react) consumes generated/openapi.yaml.
#

GENERATED_DIR := generated
OPENAPI_YAML := $(GENERATED_DIR)/openapi.yaml
MANIFEST_JSON := $(GENERATED_DIR)/schema-manifest.json

genopenapi:
	@echo "  >  Converting CUE schema to OpenAPI ..."
	@mkdir -p $(GENERATED_DIR)
	@cd cmd && go run . cue2openapi --schema .. --output ../$(OPENAPI_YAML) --manifest ../$(MANIFEST_JSON)
	@echo "  >  OpenAPI schema generation complete!"

#
# TEST BACKWARD COMPATIBILITY
#
# Repo whose releases hold the v1 baseline OpenAPI asset.
REPO ?= gemaraproj/gemara

breaking-check:
	@echo "  >  Checking for breaking schema changes ..."
	@tag=$$(gh api "repos/$(REPO)/releases" --paginate \
		--jq '.[] | select(.prerelease == false) | select(.tag_name | test("^v1\\.")) | select(any(.assets[]?; .name == "openapi.yaml")) | .tag_name' \
		| sort -V | tail -n1); \
	if [ -z "$$tag" ]; then \
		echo "  >  No v1 baseline release with an openapi.yaml asset found; skipping breaking-change check"; \
		exit 0; \
	fi; \
	echo "  >  Comparing against v1 baseline $$tag"; \
	tmp=$$(mktemp -d); \
	trap 'rm -rf "$$tmp"' EXIT; \
	gh release download "$$tag" --repo "$(REPO)" --pattern openapi.yaml --dir "$$tmp"; \
	( cd cmd && go run . breaking-check --schema .. --base "$$tmp/openapi.yaml" )
	@echo "  >  Backward compatibility check complete."

#
# REMOVE GENERATED FILES
#

cleanup:
	@echo "  >  Removing generated files..."
	@rm -rf $(GENERATED_DIR)
	@echo "  >  Cleanup complete!"

.PHONY: tidy tidycheck cuefmtcheck lintcue lintinsights test breaking-check cleanup genopenapi
