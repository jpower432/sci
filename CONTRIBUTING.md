# Contributing to Gemara

The project welcomes your contributions whether they be:

* reporting an [issue](https://github.com/gemaraproj/gemara/issues/new/choose)
* making a code contribution ([create a fork](https://github.com/gemaraproj/gemara/fork))
* updating our [docs](https://github.com/gemaraproj/gemara/blob/main/README.md)

## PR guidelines

All changes to the repository should be made via PR ([OSPS-AC-03](https://baseline.openssf.org/#osps-ac-03)).

PRs MUST meet the following criteria:

* Clear title that conforms to the [Conventional Commits spec](https://www.conventionalcommits.org/)
* Descriptive commit message
* DCO signoff (via `git commit -s` -- [OSPS-LE-01](https://baseline.openssf.org/#osps-le-01))
* All checks must pass ([OSPS-QA-04](https://baseline.openssf.org/#osps-qa-04))

### Useful make tasks when making schema changes

Use `cue fmt .` and `make cuefmtcheck` to ensure proper formatting and `make lintcue` to validate the syntax of your changes. If you forget to do this before opening a PR and your changes are invalid, the [CI workflow](.github/workflows/ci.yml) will fail and alert you.
