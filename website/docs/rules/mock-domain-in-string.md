# mock-domain-in-string

**Category**: Mock Data  
**Severity**: Warning  
**Triggered by**: `pyscn check --select mockdata`

## What it does

Flags string literals containing domains reserved for documentation and testing: `example.com`, `example.org`, `example.net`, `test.com`, `localhost`, `invalid`, `foo.com`, `bar.com`, and similar RFC 2606 / RFC 6761 names.

## Why is this a problem?

These domains exist precisely so that examples and tests can't collide with real traffic. That's useful while writing docs — and a problem once a literal ships. A hard-coded `example.com` URL in production is usually either:

- A placeholder that was meant to be replaced before release, or
- A config value that should never have been hard-coded in the first place.

In both cases the failure mode is silent: requests succeed (the domain resolves to a documentation page or nothing at all), no exception is raised, and the bug is only noticed when someone asks why signups aren't arriving.

## Example

```python
SIGNUP_URL = "https://example.com/signup"
```

## Use instead

Move the value to configuration, or inline the real URL.

```python
SIGNUP_URL = settings.signup_url
```

## Options

| Option | Default | Description |
| --- | --- | --- |
| [`mock_data.enabled`](../configuration/reference.md#mock_data) | `false` | Opt-in. |
| [`mock_data.domains`](../configuration/reference.md#mock_data) | *(RFC 2606 list)* | Override or extend the reserved domain list. |
| [`mock_data.min_severity`](../configuration/reference.md#mock_data) | `"info"` | Raise to `"warning"` to keep only this-level findings. |
| [`mock_data.ignore_tests`](../configuration/reference.md#mock_data) | `true` | Skip test files. |
| [`mock_data.ignore_patterns`](../configuration/reference.md#mock_data) | `[]` | Regex patterns matched against file paths. |

## References

- RFC 2606: *Reserved Top Level DNS Names.*
- RFC 6761: *Special-Use Domain Names.*
- Implementation: `internal/analyzer/mock_data_detector.go`.
- [Rule catalog](index.md) · [mock-email-address](mock-email-address.md) · [mock-keyword-in-code](mock-keyword-in-code.md)
