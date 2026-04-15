# unreachable-after-break

**Category**: Unreachable Code  
**Severity**: Critical  
**Triggered by**: `pyscn analyze`, `pyscn check`

## What it does

Flags statements that appear after a `break` statement inside a loop.

## Why is this a problem?

`break` exits the enclosing loop immediately. Anything after it in the same block runs zero times.

Common causes:

- **A misplaced increment or accumulator update** — the author intended it to run on the final iteration.
- **Leftover logging or cleanup** — moved below the `break` during a refactor.
- **Confused control flow** — the author expected `break` to skip only part of the iteration.

The code is unreachable, so tests cannot cover it and bugs inside it will never surface.

## Example

```python
for user in users:
    if user.id == target_id:
        break
        user.last_seen = now()   # ← never executes
```

## Use instead

Perform the work before the `break`, or remove the dead statement.

```python
for user in users:
    if user.id == target_id:
        user.last_seen = now()
        break
```

## Options

| Option | Default | Description |
| --- | --- | --- |
| [`dead_code.detect_after_break`](../configuration/reference.md#dead_code) | `true` | Set to `false` to disable this rule. |
| [`dead_code.min_severity`](../configuration/reference.md#dead_code) | `"warning"` | `"critical"` keeps only this kind of finding; `"info"` surfaces more. |
| [`dead_code.ignore_patterns`](../configuration/reference.md#dead_code) | `[]` | Regex patterns matched against the source line; matches are suppressed. |

## References

- Control-flow graph reachability analysis (`internal/analyzer/dead_code.go`).
- [Rule catalog](index.md) · [Unreachable after continue](unreachable-after-continue.md) · [Unreachable after return](unreachable-after-return.md)
