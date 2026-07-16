---
layout: page
title: Website Migration to Dedicated Repository
---

- **ADR:** 0022
- **Proposal Author(s):** @jpower432
- **Status:** Proposed

## Context

The Gemara project currently houses both the CUE schema definitions and the specification documentation website in the same repository under a single license. The website (`docs/`) includes the model specification, ADRs, tutorials, schema reference pages, and Jekyll infrastructure, all deployed to `gemara.openssf.org` via GitHub Pages.

Keeping the spec and code under a single license limits flexibility. Specification content and implementation code have different audiences, contribution models, and licensing norms. Additionally, the documentation needs versioning support aligned with schema releases, and the schema generation pipeline should consume published schemas from the CUE Registry rather than the local filesystem.

## Decision

Migrate the entire `docs/` directory and associated website infrastructure to a new repository (`gemaraproj/website`) using a big bang approach -- one coordinated cut across both repos.

**What moves to the new repo:**

- All content in `docs/` (model spec, ADRs, tutorials, schema reference, community pages)
- Jekyll configuration, custom layouts/includes, Gemfile
- The `jekyll-site.yml` GitHub Actions workflow
- The `cmd/` doc generation tooling (CUE-to-OpenAPI, OpenAPI-to-Markdown, lexicon, termlinker)
- The CNAME for `gemara.openssf.org`

**What stays in this repo:**

- CUE schema definitions
- Go test infrastructure
- CI workflow (`ci.yml`)
- A stub in `docs/` pointing to the new repo

**What changes:**

- The gendocs pipeline in the new repo pulls schemas from the CUE Registry instead of the local filesystem
- The new repo supports versioned documentation
- The new repo carries its own license independent of this repo

## Consequences

**Positive:**

- Spec content and implementation code can be licensed independently
- Documentation can be versioned independently from schema releases
- Cleaner separation of concerns between schema development and documentation
- Doc contributors don't need access to the schema repo

**Negative:**

- Schema reference generation becomes a cross-repo concern — the gendocs pipeline must pull from the CUE Registry rather than the local filesystem
- Brief site downtime possible during CNAME cutover from `gemaraproj/gemara` to `gemaraproj/website`
- Contributors familiar with the current repo layout need to be redirected

**Neutral:**

- `gemara.openssf.org` CNAME must be updated to point to GitHub Pages on `gemaraproj/website`
- Makefile doc targets (`serve`, `build`, `gendocs`, `test-links`, `cleanup`) move out of this repo
- Existing links to the site are unaffected since the domain stays the same

## Alternatives Considered

**Gradual Content Migration:** Stand up the new repo with infrastructure first, migrate content in phases (model, ADRs, tutorials, schema reference). Rejected because dual-maintenance across two repos during the transition period adds complexity and risk of broken cross-links, for a site small enough to move in one cut.

**Fork and Diverge:** Fork this repo, then strip non-doc content from the fork and non-doc content from the original. Preserves full git history for doc files. Rejected because it carries schema history baggage into the new repo and complicates the licensing story — a clean start is preferable when the point is licensing independence.
