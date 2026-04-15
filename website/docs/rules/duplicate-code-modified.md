# duplicate-code-modified

**Category**: Duplicate Code  
**Severity**: Info  
**Triggered by**: `pyscn analyze`, `pyscn check --select clones`

## What it does

Flags code blocks that share most of their structure but have added, removed, or modified statements (Type-3 clones, similarity ≥ 0.70). Disabled by default; enable by adding `"type3"` to `clones.enabled_clone_types`.

## Why is this a problem?

Near-duplicates are the most common kind of clone in real codebases and often the hardest to clean up. Two functions do almost the same thing, but one has an extra validation step or a slightly different error path. The shared part should live in one place; the differences should be the only thing that varies between call sites.

Left alone, modified clones drift: a bug gets fixed in one copy but not the other, a new feature is added to one but forgotten in the next. Reviewers stop trusting that similar-looking code behaves the same way.

This rule is Info because the right refactoring is less mechanical than for Type-1 or Type-2 clones — sometimes the differences are the whole point, and merging would hurt clarity. Treat findings as candidates worth a look, not automatic defects.

## Example

```python
def export_users_csv(users, path):
    with open(path, "w") as f:
        writer = csv.writer(f)
        writer.writerow(["id", "name", "email"])
        for u in users:
            writer.writerow([u.id, u.name, u.email])

def export_orders_csv(orders, path):
    with open(path, "w") as f:
        writer = csv.writer(f)
        writer.writerow(["id", "total", "status"])
        for o in orders:
            if o.status != "draft":
                writer.writerow([o.id, o.total, o.status])
```

## Use instead

Extract the common scaffolding and parameterize the header, row shape, and any filter.

```python
def export_csv(rows, path, header, to_row, keep=lambda _: True):
    with open(path, "w") as f:
        writer = csv.writer(f)
        writer.writerow(header)
        for row in rows:
            if keep(row):
                writer.writerow(to_row(row))

def export_users_csv(users, path):
    export_csv(users, path, ["id", "name", "email"],
               lambda u: [u.id, u.name, u.email])

def export_orders_csv(orders, path):
    export_csv(orders, path, ["id", "total", "status"],
               lambda o: [o.id, o.total, o.status],
               keep=lambda o: o.status != "draft")
```

## Options

| Option | Default | Description |
| --- | --- | --- |
| [`clones.type3_threshold`](../configuration/reference.md#clones) | `0.70` | Minimum similarity for a pair to be reported as modified. |
| [`clones.enabled_clone_types`](../configuration/reference.md#clones) | `["type1","type2","type4"]` | Add `"type3"` to enable this rule. |
| [`clones.similarity_threshold`](../configuration/reference.md#clones) | `0.65` | Global floor applied before per-type thresholds. |
| [`clones.min_lines`](../configuration/reference.md#clones) | `5` | Minimum fragment size in lines. |
| [`clones.min_nodes`](../configuration/reference.md#clones) | `10` | Minimum fragment size in AST nodes. |

## References

- Clone detection implementation (`internal/analyzer/clone_detector.go`, `internal/analyzer/apted.go`).
- [Rule catalog](index.md) · [Identical clones](duplicate-code-identical.md) · [Renamed clones](duplicate-code-renamed.md) · [Semantic clones](duplicate-code-semantic.md)
