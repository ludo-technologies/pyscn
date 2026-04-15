# unreachable-branch

**Category**: Unreachable Code  
**Severity**: Warning  
**Triggered by**: `pyscn analyze`, `pyscn check`

## What it does

Flags an `if`, `elif`, or `else` branch that cannot be taken because every preceding branch terminates with `return`, `raise`, `break`, or `continue`.

## Why is this a problem?

When each earlier branch already exits the function or loop, the remaining branch is logically dead. The guard still looks meaningful to a reader, which hides the real control flow.

This usually indicates:

- **Redundant conditions** — the `else` exists only because the `if` used to fall through.
- **A subtle bug** — the author expected the later branch to run in some case, but the earlier exits make it impossible.
- **Stale defensive code** — a fallback that can no longer be reached.

Tests cannot cover the branch, and reviewers waste time reasoning about a path that never runs.

## Example

```python
def classify(payment):
    if payment.amount < 0:
        raise ValueError("negative amount")
    elif payment.amount == 0:
        return "empty"
    else:
        return "normal"
    return "unknown"   # ← unreachable branch
```

## Use instead

Remove the dead branch, or restructure the earlier branches so the fallback is actually reachable.

```python
def classify(payment):
    if payment.amount < 0:
        raise ValueError("negative amount")
    if payment.amount == 0:
        return "empty"
    return "normal"
```

## Options

| Option | Default | Description |
| --- | --- | --- |
| [`dead_code.detect_unreachable_branches`](../configuration/reference.md#dead_code) | `true` | Set to `false` to disable this rule. |
| [`dead_code.min_severity`](../configuration/reference.md#dead_code) | `"warning"` | Raise to `"critical"` to hide these findings; lower to `"info"` to surface more. |
| [`dead_code.ignore_patterns`](../configuration/reference.md#dead_code) | `[]` | Regex patterns matched against the source line; matches are suppressed. |

## References

- Control-flow graph reachability analysis (`internal/analyzer/dead_code.go`).
- [Rule catalog](index.md) · [Unreachable after return](unreachable-after-return.md) · [Unreachable after raise](unreachable-after-raise.md)
