# test-credential-in-code

**Category**: Mock Data  
**Severity**: Warning  
**Triggered by**: `pyscn check --select mockdata`

## What it does

Flags string literals that look like obvious placeholder credentials: `password123`, `secret123`, `testpassword`, `token0`, `api_key_test`, and similar patterns built from credential-shaped words plus trivial suffixes.

## Why is this a problem?

These aren't real secrets — that's the problem. They're the values typed in while setting up a client, writing a first test, or filling a required field. Once checked in, they have two failure modes:

- **Embarrassment / correctness**: the literal is used as a default and ships to users, so the "default admin password" really is `password123`.
- **Missed rotation**: a real secret was meant to replace the placeholder before release; nobody noticed that it didn't.

pyscn is not a security scanner and doesn't try to detect high-entropy secrets. This rule catches the opposite case: low-entropy, obviously-fake credentials that should never have been in source in the first place.

## Example

```python
DEFAULT_PASSWORD = "password123"
```

## Use instead

Read credentials from the environment or a secret manager. If a default is genuinely required for local development, keep it in a separate config file that isn't loaded in production.

```python
import os

DEFAULT_PASSWORD = os.environ["APP_DEFAULT_PASSWORD"]
```

## Options

| Option | Default | Description |
| --- | --- | --- |
| [`mock_data.enabled`](../configuration/reference.md#mock_data) | `false` | Opt-in. |
| [`mock_data.min_severity`](../configuration/reference.md#mock_data) | `"info"` | Raise to `"warning"` to keep only this-level findings. |
| [`mock_data.ignore_tests`](../configuration/reference.md#mock_data) | `true` | Skip test files — test credentials belong in tests. |
| [`mock_data.ignore_patterns`](../configuration/reference.md#mock_data) | `[]` | Regex patterns matched against file paths. |

## References

- Implementation: `internal/analyzer/mock_data_detector.go`.
- [Rule catalog](index.md) · [repetitive-string-literal](repetitive-string-literal.md) · [mock-keyword-in-code](mock-keyword-in-code.md)
