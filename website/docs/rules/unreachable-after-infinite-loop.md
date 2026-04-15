# unreachable-after-infinite-loop

**Category**: Unreachable Code  
**Severity**: Warning  
**Triggered by**: `pyscn analyze`, `pyscn check`

## What it does

Flags statements that follow a loop with no reachable exit, such as `while True:` without a `break` or `return`.

## Why is this a problem?

If a loop has no path out, execution never proceeds past it. Anything written after the loop is dead.

This is usually one of:

- **A forgotten exit condition** — the loop was meant to terminate but the `break` was lost in a refactor.
- **A misplaced cleanup** — shutdown or teardown code sitting after a worker loop that never returns.
- **A copy-paste error** — post-loop logic left over from a previous version of the function.

The reader expects the trailing code to run eventually. It does not.

## Example

```python
def run_worker(queue):
    while True:
        job = queue.get()
        job.run()
    queue.close()   # ← never executes
```

## Use instead

Give the loop a reachable exit, or remove the unreachable tail.

```python
def run_worker(queue):
    while not queue.closed:
        job = queue.get()
        job.run()
    queue.close()
```

## Options

| Option | Default | Description |
| --- | --- | --- |
| [`dead_code.enabled`](../configuration/reference.md#dead_code) | `true` | This rule has no dedicated toggle; it is controlled by `dead_code.enabled` and CFG analysis. |
| [`dead_code.min_severity`](../configuration/reference.md#dead_code) | `"warning"` | Raise to `"critical"` to hide these findings; lower to `"info"` to surface more. |
| [`dead_code.ignore_patterns`](../configuration/reference.md#dead_code) | `[]` | Regex patterns matched against the source line; matches are suppressed. |

## References

- Control-flow graph reachability analysis (`internal/analyzer/dead_code.go`).
- [Rule catalog](index.md) · [Unreachable after return](unreachable-after-return.md) · [Unreachable branch](unreachable-branch.md)
