# placeholder-phone-number

**Category**: Mock Data  
**Severity**: Warning  
**Triggered by**: `pyscn check --select mockdata`

## What it does

Flags phone numbers in string literals that follow obviously fake patterns: all zeros (`000-0000-0000`), sequential digits (`123-456-7890`, `012-345-6789`), or long runs of repeated digits.

## Why is this a problem?

A placeholder phone number is the kind of value that survives from the first draft of a form and is never revisited. It validates, it formats, and it round-trips through the database — so nothing breaks until a real user sees it on a confirmation screen or a support agent tries to call it.

## Example

```python
default_phone = "000-0000-0000"
```

## Use instead

Leave the field empty, require it from the caller, or pull it from configuration. A phone number that isn't known should be absent, not faked.

```python
default_phone: str | None = None
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
- [Rule catalog](index.md) · [placeholder-uuid](placeholder-uuid.md) · [repetitive-string-literal](repetitive-string-literal.md)
