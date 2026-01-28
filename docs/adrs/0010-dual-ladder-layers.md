---
layout: page
title: Dual-ladders Within Each Layer
---

- **ADR:** 0010
- **Proposal Author(s):** @eddie-knight
- **Status:** Accepted

## Context

We have been treating each layer as cumulative, with a single descriptor for the layer. In many cases, we have allowed for additional artificts with a single layer, such as a list of threats that lives alongside controls in a Layer 2 [Control Catalog](../model/02-definitions.html#control-catalog), or a list of Risks within a [Policy](../model/02-definitions.html#policy) Document.

Here is an overview of some common phrasing we've been using:

- Layer 1: [Guidance](../model/02-definitions.html#guidance) (informed by known attack/negligence vectors)
- Layer 2: Controls (catalog includes technology-specific threats)
- Layer 3: [Policy](../model/02-definitions.html#policy) (document links to organizational [risk](../model/02-definitions.html#risk) considerations)
- Layer 4: Sensitive Activities
- Layer 5: [Evaluation](../model/02-definitions.html#evaluation) (config scans and behavior scans)
- Layer 6: [Enforcement](../model/02-definitions.html#enforcement) (gates and remediation)
- Layer 7: [Audit](../model/02-definitions.html#audit) (manual and [continuous monitoring](../model/02-definitions.html#continuous-monitoring)

## Decision

We will allow for two primary artifacts within each layer. For current development conversations we will informally refer to these as "Red" and "Blue" until a better term is found, because these are _similar_ but not perfectly aligned with cybersecurity "Red Team" and "Blue Team" concepts.

The specific phrasing or terms may be refined or adjusted, but the meaning shall be roughly as follows:

| Layer | Red | Blue |
| 1 | [Vector](../model/02-definitions.html#vector) | [Guidance](../model/02-definitions.html#guidance) |
| 2 | [Threat](../model/02-definitions.html#threat) | [Control](../model/02-definitions.html#control) |
| 3 | [Risk](../model/02-definitions.html#risk) | [Policy](../model/02-definitions.html#policy) |
| 4 | [Risk](../model/02-definitions.html#risk) Actualization | Sensitive Activities |
| 5 | Behavioral [Evaluation](../model/02-definitions.html#evaluation) | [Intent Evaluation](../model/02-definitions.html#intent-evaluation) |
| 6 | [Remediative Enforcement](../model/02-definitions.html#remediative-enforcement) | [Preventive Enforcement](../model/02-definitions.html#preventive-enforcement) |
| 7 | Continuous [Compliance](../model/02-definitions.html#compliance) Monitoring | Point-in-Time [Audit](../model/02-definitions.html#audit) |

## Consequences

- All documentation and web content must be updated
- Adopters of previous versions will be disrupted by this change in terminology

## Alternatives Considered

We could not do this, but it leaves the relevant parts open to potential confusion.
