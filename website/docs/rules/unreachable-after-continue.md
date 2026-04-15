# unreachable-after-continue

**Category**: Unreachable Code  
**Severity**: Critical  
**Triggered by**: `pyscn analyze`, `pyscn check`

## What it does

Flags statements that appear after a `continue` statement inside a loop.

## Why is this a problem?

`continue` jumps straight to the next iteration. Any statement following it in the same block is skipped on every iteration, so it runs zero times.

Typical causes:

- **Reordered logic** — a guard was converted to `continue` and trailing work was left behind.
- **A misplaced side effect** — counter updates or logging that should run before skipping.
- **Misunderstood semantics** — the author expected `continue` to behave like `pass`.

Because the statement is unreachable, tests cannot cover it and the intended behaviour silently never happens.

## Example

```python
for order in orders:
    if order.status == "cancelled":
        continue
        metrics.record_skip(order.id)   # ← never executes
    process(order)
```

## Use instead

Run the statement before `continue`, or drop it.

```python
for order in orders:
    if order.status == "cancelled":
        metrics.record_skip(order.id)
        continue
    process(order)
```

## Options

| Option | Default | Description |
| --- | --- | --- |
| [`dead_code.detect_after_continue`](../configuration/reference.md#dead_code) | `true` | Set to `false` to disable this rule. |
| [`dead_code.min_severity`](../configuration/reference.md#dead_code) | `"warning"` | `"critical"` keeps only this kind of finding; `"info"` surfaces more. |
| [`dead_code.ignore_patterns`](../configuration/reference.md#dead_code) | `[]` | Regex patterns matched against the source line; matches are suppressed. |

## References

- Control-flow graph reachability analysis (`internal/analyzer/dead_code.go`).
- [Rule catalog](index.md) · [Unreachable after break](unreachable-after-break.md) · [Unreachable after return](unreachable-after-return.md)
