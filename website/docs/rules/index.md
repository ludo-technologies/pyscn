# Rule catalog

pyscn ships 32 rules across 7 categories. Every rule has a page that describes what it detects, why it's a problem, a bad example, and how to fix it.

Click a rule name to open its page.

## Unreachable Code

Dead code that can never execute. Detected through control-flow graph reachability analysis.

| Rule | Severity |
| ---- | -------- |
| [`unreachable-after-return`](unreachable-after-return.md) | Critical |
| [`unreachable-after-raise`](unreachable-after-raise.md) | Critical |
| [`unreachable-after-break`](unreachable-after-break.md) | Critical |
| [`unreachable-after-continue`](unreachable-after-continue.md) | Critical |
| [`unreachable-after-infinite-loop`](unreachable-after-infinite-loop.md) | Warning |
| [`unreachable-branch`](unreachable-branch.md) | Warning |

## Duplicate Code

Copy-paste or near-copy-paste code fragments across the project.

| Rule | Severity |
| ---- | -------- |
| [`duplicate-code-identical`](duplicate-code-identical.md) | Warning |
| [`duplicate-code-renamed`](duplicate-code-renamed.md) | Warning |
| [`duplicate-code-modified`](duplicate-code-modified.md) | Info (opt-in) |
| [`duplicate-code-semantic`](duplicate-code-semantic.md) | Warning |

## Complexity

Functions that are too branchy to test or reason about reliably.

| Rule | Severity |
| ---- | -------- |
| [`high-cyclomatic-complexity`](high-cyclomatic-complexity.md) | By threshold |

## Class Design

Classes that depend on too many things or do too many unrelated jobs.

| Rule | Severity |
| ---- | -------- |
| [`high-class-coupling`](high-class-coupling.md) | By threshold |
| [`low-class-cohesion`](low-class-cohesion.md) | By threshold |

## Dependency Injection

Constructor and collaborator patterns that hurt testability.

| Rule | Severity |
| ---- | -------- |
| [`too-many-constructor-parameters`](too-many-constructor-parameters.md) | Warning |
| [`global-state-dependency`](global-state-dependency.md) | Error |
| [`module-variable-dependency`](module-variable-dependency.md) | Warning |
| [`singleton-pattern-dependency`](singleton-pattern-dependency.md) | Warning |
| [`concrete-type-hint-dependency`](concrete-type-hint-dependency.md) | Info |
| [`concrete-instantiation-dependency`](concrete-instantiation-dependency.md) | Warning |
| [`service-locator-pattern`](service-locator-pattern.md) | Warning |

## Module Structure

Import graph problems: cycles, long chains, layer violations.

| Rule | Severity |
| ---- | -------- |
| [`circular-import`](circular-import.md) | By cycle size |
| [`deep-import-chain`](deep-import-chain.md) | Info |
| [`layer-violation`](layer-violation.md) | By architecture rule |
| [`low-package-cohesion`](low-package-cohesion.md) | Warning |

## Mock Data

Placeholder data accidentally shipped to production.

| Rule | Severity |
| ---- | -------- |
| [`mock-keyword-in-code`](mock-keyword-in-code.md) | Info / Warning |
| [`mock-domain-in-string`](mock-domain-in-string.md) | Warning |
| [`mock-email-address`](mock-email-address.md) | Warning |
| [`placeholder-phone-number`](placeholder-phone-number.md) | Warning |
| [`placeholder-uuid`](placeholder-uuid.md) | Warning |
| [`placeholder-comment`](placeholder-comment.md) | Info |
| [`repetitive-string-literal`](repetitive-string-literal.md) | Info |
| [`test-credential-in-code`](test-credential-in-code.md) | Warning |

## Selecting rules on the command line

Most users run all rules with `pyscn analyze`. For CI, filter by analyzer category:

```bash
pyscn check --select deadcode          # only unreachable-code rules
pyscn check --select clones            # only duplicate-code rules
pyscn check --select complexity        # only high-cyclomatic-complexity
pyscn check --select deps              # circular-import + deep-import-chain + layer-violation
pyscn check --select di                # all dependency-injection rules (opt-in)
pyscn check --select mockdata          # all mock-data rules (opt-in)
pyscn check --select complexity,deadcode,deps   # combine
```

See [`pyscn check`](../cli/check.md) for the full flag list.

## Severity meanings

| Severity | Intent |
| -------- | --- |
| **Critical** | Almost always a bug. Prefer fixing before merging. |
| **Error** | High-risk pattern. Usually should fail CI. |
| **Warning** | Worth reviewing. Default fail threshold for `pyscn check`. |
| **Info** | Informational. Surfaces only when `min_severity = "info"` or equivalent. |
| **By threshold** | Severity depends on a numeric threshold (see the rule's Options). |
