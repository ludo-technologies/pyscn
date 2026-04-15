# duplicate-code-renamed

**Category**: Duplicate Code  
**Severity**: Warning  
**Triggered by**: `pyscn analyze`, `pyscn check --select clones`

## What it does

Flags code blocks with the same structure but different identifiers or literals (Type-2 clones, similarity ≥ 0.75).

## Why is this a problem?

Renamed clones are what happens when someone copies a function, runs a find-and-replace on the variable names, and moves on. The structure is identical; only the nouns changed. The maintenance cost is the same as for textually identical clones — every change has to be applied in multiple places — but the duplication is harder to spot by eye because the words differ.

This is also a signal that the original code wasn't parameterized when it should have been. The thing that varies (a type, a field name, a constant) is a natural function argument.

## Example

```python
def total_for_orders(orders):
    total = 0
    for order in orders:
        if order.status == "paid":
            total += order.amount
    return total

def total_for_invoices(invoices):
    total = 0
    for invoice in invoices:
        if invoice.status == "settled":
            total += invoice.amount
    return total
```

## Use instead

Extract a generic helper that takes the varying predicate and field accessors as parameters.

```python
def total_where(items, is_active):
    return sum(item.amount for item in items if is_active(item))

def total_for_orders(orders):
    return total_where(orders, lambda o: o.status == "paid")

def total_for_invoices(invoices):
    return total_where(invoices, lambda i: i.status == "settled")
```

## Options

| Option | Default | Description |
| --- | --- | --- |
| [`clones.type2_threshold`](../configuration/reference.md#clones) | `0.75` | Minimum similarity for a pair to be reported as renamed. |
| [`clones.similarity_threshold`](../configuration/reference.md#clones) | `0.65` | Global floor applied before per-type thresholds. |
| [`clones.ignore_identifiers`](../configuration/reference.md#clones) | `true` | Treat differing variable names as equivalent when computing similarity. |
| [`clones.ignore_literals`](../configuration/reference.md#clones) | `true` | Treat differing numeric and string literals as equivalent. |
| [`clones.min_lines`](../configuration/reference.md#clones) | `5` | Minimum fragment size in lines. |
| [`clones.enabled_clone_types`](../configuration/reference.md#clones) | `["type1","type2","type4"]` | Include `"type2"` to keep this rule active. |

## References

- Clone detection implementation (`internal/analyzer/clone_detector.go`, `internal/analyzer/apted.go`).
- [Rule catalog](index.md) · [Identical clones](duplicate-code-identical.md) · [Modified clones](duplicate-code-modified.md) · [Semantic clones](duplicate-code-semantic.md)
