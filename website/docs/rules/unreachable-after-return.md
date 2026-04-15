# unreachable-after-return

**Category**: Unreachable Code  
**Severity**: Critical  
**Triggered by**: `pyscn analyze`, `pyscn check`

## What it does

Flags statements that appear after a `return` statement in the same code block.

## Why is this a problem?

Code placed after `return` never executes. This is usually one of:

- **A leftover from refactoring** — the `return` was moved up and the code below was forgotten.
- **A bug** — the programmer expected the code to run but a control-flow change made it unreachable.
- **A misplaced cleanup** — something that should run before returning.

In every case, the code is dead: it takes up reading time, it isn't covered by tests (because it can't be), and if it hides a bug, that bug will never be reported from user behaviour.

## Example

```python
def charge(order):
    if order.total <= 0:
        return None
        log.debug("zero-value charge")   # ← never executes
    ...
```

## Use instead

Move the statement above the `return`, or delete it if it's no longer needed.

```python
def charge(order):
    if order.total <= 0:
        log.debug("zero-value charge")
        return None
    ...
```

## Options

| Option | Default | Description |
| --- | --- | --- |
| [`dead_code.detect_after_return`](../configuration/reference.md#dead_code) | `true` | Set to `false` to disable this rule. |
| [`dead_code.min_severity`](../configuration/reference.md#dead_code) | `"warning"` | `"critical"` keeps only this kind of finding; `"info"` surfaces more. |
| [`dead_code.ignore_patterns`](../configuration/reference.md#dead_code) | `[]` | Regex patterns matched against the source line; matches are suppressed. |

## References

- Control-flow graph reachability analysis (`internal/analyzer/dead_code.go`).
- [Rule catalog](index.md) · [Unreachable branch](unreachable-branch.md) · [Unreachable after raise](unreachable-after-raise.md)
