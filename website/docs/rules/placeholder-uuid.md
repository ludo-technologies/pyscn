# placeholder-uuid

**Category**: Mock Data  
**Severity**: Warning  
**Triggered by**: `pyscn check --select mockdata`

## What it does

Flags UUID-shaped string literals with very low entropy: the nil UUID (`00000000-0000-0000-0000-000000000000`), all-ones, all-`f`, or long runs of repeated characters that nevertheless parse as a UUID.

## Why is this a problem?

The nil UUID is a legitimate value in a few contexts, but in most application code it's a leftover from a stub `DEFAULT_USER_ID = "00..."` that was meant to be replaced. Because it parses and serialises like any other UUID, foreign-key lookups, log lines, and audit trails all accept it — and rows start collapsing onto the same "user" without anything raising an error.

## Example

```python
DEFAULT_USER_ID = "00000000-0000-0000-0000-000000000000"
```

## Use instead

Generate a real UUID at the point of use, or require the caller to supply one. Don't carry a sentinel value that looks like data.

```python
import uuid

def new_user_id() -> uuid.UUID:
    return uuid.uuid4()
```

## Options

| Option | Default | Description |
| --- | --- | --- |
| [`mock_data.enabled`](../configuration/reference.md#mock_data) | `false` | Opt-in. |
| [`mock_data.min_severity`](../configuration/reference.md#mock_data) | `"info"` | Raise to `"warning"` to keep only this-level findings. |
| [`mock_data.ignore_tests`](../configuration/reference.md#mock_data) | `true` | Skip test files. |
| [`mock_data.ignore_patterns`](../configuration/reference.md#mock_data) | `[]` | Regex patterns matched against file paths. |

## References

- Implementation: `internal/analyzer/mock_data_detector.go`.
- [Rule catalog](index.md) · [placeholder-phone-number](placeholder-phone-number.md) · [repetitive-string-literal](repetitive-string-literal.md)
