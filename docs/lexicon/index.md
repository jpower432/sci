---
layout: page
title: The Gemara Lexicon
---

While the **[Model](../model)** describes the structure and relationships of GRC activities, the Lexicon provides the vocabulary. It establishes term meanings within the Gemara context, ensuring consistent definitions when teams discuss "controls," "evaluations," or "policies."

The Lexicon helps teams:
- **Agree on terminology** across activities and teams
- **Communicate effectively** across organizational boundaries
- **Understand relationships** between concepts in the Model

## Available Versions

- **[In development version](./dev)** - Development version (unreleased)

## Contributing

The Lexicon evolves based on community needs:

- **Term needs clarification?** Open an issue or discussion
- **Propose a new definition?** Submit a PR
- **Found an inconsistency?** Report it

See the [Contributing Guide](https://github.com/ossf/gemara/blob/main/CONTRIBUTING.md) for details.

## Relationship to Other Components

### [The Model](../model)
Provides the structural framework. Terms are organized by their relationship to the Model's layers.

### [The Implementation](../implementation)
Uses Lexicon definitions to inform schema design and SDK documentation. Refer to the Lexicon for precise term meanings.

## Using the Lexicon

**For GRC Professionals:** Use the Lexicon to ensure consistent terminology in discussions and documentation.

**For Developers:** Refer to the Lexicon when working with the Implementation to understand precise term meanings in schemas and APIs.

## Releases

The Lexicon is maintained in YAML format (`lexicon.yaml`) and compiled into markdown for documentation.

New lexicon versions are released manually. Releases can be either:
- **Development version** (`dev.md`) - for ongoing work and testing
- **CalVer version** (`YY-MM-DD.md`, e.g., `25-01-15.md`) - for official releases

To release a new version, compile the lexicon using the `buildlexicon` tool:

```bash
go run ./cmd/buildlexicon --lexicon lexicon.yaml --output docs/lexicon/VERSION.md --version VERSION
```

Where `VERSION` is either `dev` or a CalVer date in `YY-MM-DD` format.

