# mock-email-address

**Category**: Mock Data  
**Severity**: Warning  
**Triggered by**: `pyscn check --select mockdata`

## What it does

Flags email addresses whose domain part is reserved for testing: `test@example.com`, `admin@test.com`, `foo@localhost`, and similar.

## Why is this a problem?

A test-domain email in application code is almost always a leftover from a fixture, a tutorial, or a "fill this in later" stub. Unlike a malformed address, it passes validation and serialisation cleanly, so it silently ends up in database rows, notification queues, and "from" headers. The usual discovery path is a support ticket asking why no one received a reset link.

This rule complements [mock-domain-in-string](mock-domain-in-string.md) but treats the email form specifically so it can keep the domain list tight and the match precise.

## Example

```python
admin_email = "admin@example.com"
```

## Use instead

Read the address from configuration, or accept it as a parameter.

```python
admin_email = settings.admin_email
```

## Options

| Option | Default | Description |
| --- | --- | --- |
| [`mock_data.enabled`](../configuration/reference.md#mock_data) | `false` | Opt-in. |
| [`mock_data.domains`](../configuration/reference.md#mock_data) | *(RFC 2606 list)* | Domains considered placeholder; shared with `mock-domain-in-string`. |
| [`mock_data.min_severity`](../configuration/reference.md#mock_data) | `"info"` | Raise to `"warning"` for only this-level findings. |
| [`mock_data.ignore_tests`](../configuration/reference.md#mock_data) | `true` | Skip test files. |
| [`mock_data.ignore_patterns`](../configuration/reference.md#mock_data) | `[]` | Regex patterns matched against file paths. |

## References

- RFC 2606: *Reserved Top Level DNS Names.*
- Implementation: `internal/analyzer/mock_data_detector.go`.
- [Rule catalog](index.md) · [mock-domain-in-string](mock-domain-in-string.md) · [test-credential-in-code](test-credential-in-code.md)
