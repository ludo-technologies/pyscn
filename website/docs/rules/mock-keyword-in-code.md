# mock-keyword-in-code

**Category**: Mock Data  
**Severity**: Info (in strings) / Warning (in identifiers)  
**Triggered by**: `pyscn check --select mockdata`

## What it does

Flags identifiers and string literals that contain common placeholder keywords: `mock`, `fake`, `dummy`, `test`, `sample`, `example`, `placeholder`, `stub`, `fixture`, `temp`, `foo`, `bar`, `baz`, `lorem`, `ipsum`.

## Why is this a problem?

These words are what you type while you're still figuring something out. They're fine in a notebook, a test file, or a five-minute spike — but when one survives into production it usually means a stub was never replaced. The name `foo` in a checked-in module is almost never what the author meant to ship.

Matches in **identifiers** are treated as warnings because a bound name (`foo = get_user()`) changes behaviour. Matches in **string literals** are info-level because a leftover `"fake_user"` is more often cosmetic than broken — but still worth reviewing before a release.

## Example

```python
def create_user():
    name = "fake_user"    # string literal matches
    foo = get_user()      # identifier `foo` matches
    return foo
```

## Use instead

Remove the placeholder. Use real data, read from configuration, or move the stub into a test fixture where it belongs.

```python
def create_user(name: str):
    user = get_user(name)
    return user
```

## Options

| Option | Default | Description |
| --- | --- | --- |
| [`mock_data.enabled`](../configuration/reference.md#mock_data) | `false` | Opt-in; the whole category is off by default. |
| [`mock_data.keywords`](../configuration/reference.md#mock_data) | *(built-in list)* | Override the keyword list for this rule. |
| [`mock_data.min_severity`](../configuration/reference.md#mock_data) | `"info"` | `"warning"` keeps only identifier matches. |
| [`mock_data.ignore_tests`](../configuration/reference.md#mock_data) | `true` | Skip files that look like tests. |
| [`mock_data.ignore_patterns`](../configuration/reference.md#mock_data) | `[]` | Regex patterns matched against file paths; matches are suppressed. |

## References

- Implementation: `internal/analyzer/mock_data_detector.go`.
- [Rule catalog](index.md) · [mock-domain-in-string](mock-domain-in-string.md) · [test-credential-in-code](test-credential-in-code.md) · [placeholder-comment](placeholder-comment.md)
