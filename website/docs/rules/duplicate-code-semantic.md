# duplicate-code-semantic

**Category**: Duplicate Code  
**Severity**: Warning  
**Triggered by**: `pyscn analyze`, `pyscn check --select clones`

## What it does

Flags code blocks that are syntactically different but compute the same result (Type-4 clones, similarity ≥ 0.65). Uses data-flow analysis to compare behaviour rather than structure.

## Why is this a problem?

Semantic clones are the duplication you don't notice during review. One function uses a loop, the other a comprehension; one builds a dict via `update`, the other via merge syntax. The code looks different, so it passes visual inspection, but both implementations are doing the same job.

The risks are the same as for any duplication — changes have to be made in multiple places — but there's an extra cost. Readers can't tell at a glance whether the two implementations agree on edge cases. Does the loop version also skip `None`? Does the comprehension raise on an empty input? The mental audit has to be done every time.

Collapsing to a single implementation removes the audit and the drift.

## Example

```python
def unique_emails(users):
    seen = set()
    result = []
    for u in users:
        if u.email not in seen:
            seen.add(u.email)
            result.append(u.email)
    return result

def distinct_emails(users):
    return list({u.email: None for u in users}.keys())
```

## Use instead

Pick one implementation and use it everywhere. Prefer the clearer version; if both have merit, keep one and document why.

```python
def unique_emails(users):
    """Return user emails in first-seen order, without duplicates."""
    return list(dict.fromkeys(u.email for u in users))
```

## Options

| Option | Default | Description |
| --- | --- | --- |
| [`clones.type4_threshold`](../configuration/reference.md#clones) | `0.65` | Minimum similarity for a pair to be reported as semantic. |
| [`clones.enable_dfa`](../configuration/reference.md#clones) | `true` | Enables the data-flow analysis that powers Type-4 detection. |
| [`clones.similarity_threshold`](../configuration/reference.md#clones) | `0.65` | Global floor applied before per-type thresholds. |
| [`clones.enabled_clone_types`](../configuration/reference.md#clones) | `["type1","type2","type4"]` | Include `"type4"` to keep this rule active. |
| [`clones.min_lines`](../configuration/reference.md#clones) | `5` | Minimum fragment size in lines. |

## References

- Clone detection implementation (`internal/analyzer/clone_detector.go`, `internal/analyzer/apted.go`).
- [Rule catalog](index.md) · [Identical clones](duplicate-code-identical.md) · [Renamed clones](duplicate-code-renamed.md) · [Modified clones](duplicate-code-modified.md)
