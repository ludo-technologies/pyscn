# placeholder-comment

**Category**: Mock Data  
**Severity**: Info  
**Triggered by**: `pyscn check --select mockdata`

## What it does

Flags comments containing unfinished-work markers: `TODO`, `FIXME`, `XXX`, `HACK`, `BUG`, `NOTE`.

## Why is this a problem?

A `# TODO` in source is a promise the author made to their future self, with no due date and no reviewer. Most codebases accumulate them faster than they're cleared. Each one is a small piece of hidden scope — the reader has to decide whether it's still relevant, whether it blocks a change, and whether someone is actually tracking it.

This rule doesn't argue that every marker is a bug. It surfaces the list so you can decide, per project, whether to clear them, turn them into tracked issues, or accept them with an explicit policy.

## Example

```python
def process_order(order):
    # TODO: handle refunds
    ...
```

## Use instead

Either implement the work, or convert the marker into a tracked issue link so the intent lives somewhere with a close state.

```python
def process_order(order):
    # Refunds are handled by the billing service: see issue #1423.
    ...
```

## Options

| Option | Default | Description |
| --- | --- | --- |
| [`mock_data.enabled`](../configuration/reference.md#mock_data) | `false` | Opt-in. |
| [`mock_data.min_severity`](../configuration/reference.md#mock_data) | `"info"` | Raise to `"warning"` to exclude this rule. |
| [`mock_data.ignore_tests`](../configuration/reference.md#mock_data) | `true` | Skip test files. |
| [`mock_data.ignore_patterns`](../configuration/reference.md#mock_data) | `[]` | Regex patterns matched against file paths. |

## References

- Implementation: `internal/analyzer/mock_data_detector.go`.
- [Rule catalog](index.md) · [mock-keyword-in-code](mock-keyword-in-code.md)
