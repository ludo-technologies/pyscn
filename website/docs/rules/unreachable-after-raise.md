# unreachable-after-raise

**Category**: Unreachable Code  
**Severity**: Critical  
**Triggered by**: `pyscn analyze`, `pyscn check`

## What it does

Flags statements that appear after a `raise` statement in the same code block.

## Why is this a problem?

A `raise` unconditionally unwinds the stack. Any statement following it in the same block is never executed.

This usually points to:

- **Stale cleanup** — code that was supposed to run before the exception was thrown.
- **A refactor artefact** — the `raise` replaced an earlier branch and surrounding lines were left behind.
- **A logic bug** — the author assumed execution would continue past the `raise`.

Dead code after `raise` will never be exercised by tests and will never surface in production, so any bug it hides is silent.

## Example

```python
def withdraw(account, amount):
    if amount > account.balance:
        raise InsufficientFunds(account.id)
        account.balance -= amount   # ← never executes
    account.balance -= amount
```

## Use instead

Move the statement above the `raise`, or delete it.

```python
def withdraw(account, amount):
    if amount > account.balance:
        raise InsufficientFunds(account.id)
    account.balance -= amount
```

## Options

| Option | Default | Description |
| --- | --- | --- |
| [`dead_code.detect_after_raise`](../configuration/reference.md#dead_code) | `true` | Set to `false` to disable this rule. |
| [`dead_code.min_severity`](../configuration/reference.md#dead_code) | `"warning"` | `"critical"` keeps only this kind of finding; `"info"` surfaces more. |
| [`dead_code.ignore_patterns`](../configuration/reference.md#dead_code) | `[]` | Regex patterns matched against the source line; matches are suppressed. |

## References

- Control-flow graph reachability analysis (`internal/analyzer/dead_code.go`).
- [Rule catalog](index.md) · [Unreachable after return](unreachable-after-return.md) · [Unreachable branch](unreachable-branch.md)
