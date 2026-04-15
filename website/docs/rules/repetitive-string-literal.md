# repetitive-string-literal

**Category**: Mock Data  
**Severity**: Info  
**Triggered by**: `pyscn check --select mockdata`

## What it does

Flags string literals of length 4 to 20 with highly repetitive character patterns: `aaaa`, `1111`, `xxxxxxxx`, and similar single-character or two-character runs.

## Why is this a problem?

A string like `"aaaaaaaaaaaaaaaa"` is almost never a real value — it's the shape of something typed to get past a validator while the developer was wiring something else up. Left in production, it becomes an API key, a hash input, or a token that looks like data and passes length checks but has no meaning.

The rule is length-bounded (4–20 characters) to avoid flagging intentional fills like padding constants or test vectors that legitimately need repeated characters.

## Example

```python
api_key = "aaaaaaaaaaaaaaaa"
```

## Use instead

Read secrets and tokens from configuration or a secret store. Don't bake placeholder values into the source.

```python
import os

api_key = os.environ["SERVICE_API_KEY"]
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
- [Rule catalog](index.md) · [test-credential-in-code](test-credential-in-code.md) · [placeholder-uuid](placeholder-uuid.md)
