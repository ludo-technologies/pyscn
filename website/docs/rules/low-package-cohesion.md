# low-package-cohesion

**Category**: Module Structure  
**Severity**: Warning  
**Triggered by**: `pyscn analyze`, `pyscn check --select deps`

## What it does

Flags a package whose internal cohesion score falls below `architecture.min_cohesion` (default `0.5`). Cohesion is measured as the ratio of actual intra-package imports to the number possible between its submodules — a package whose modules never import each other scores `0`.

## Why is this a problem?

A package is supposed to be a single concept that happens to be split across files. When the files don't reference each other, the package is just a folder of unrelated code sharing a namespace:

- **Misleading import paths.** `from myapp.utils import X` suggests a relationship between `X` and everything else in `utils`; low cohesion means that promise is empty.
- **No natural owner.** Nobody is responsible for "`utils`" as a whole, because there is no whole.
- **Grows without bound.** Miscellaneous packages accumulate unrelated helpers until they become a dumping ground.
- **Hides a missing abstraction.** Often the right move is not "keep adding", but to find the real concept that two of the submodules share and extract it.

## Example

```
myapp/utils/
    __init__.py
    string_utils.py     # slugify, truncate
    math_utils.py       # clamp, lerp
    io_utils.py         # atomic_write, read_json
```

None of these three modules imports any of the others. The `utils` package has zero cohesion.

## Use instead

Split the package into focused packages named after what they actually do:

```
myapp/text/          # slugify, truncate, and the helpers they share
myapp/geometry/      # clamp, lerp
myapp/fs/            # atomic_write, read_json
```

Or — if the contents really are unrelated one-off helpers — acknowledge that and stop pretending otherwise. Name the package `misc` or move each helper to the module that actually uses it, and exclude the dumping ground from cohesion checks.

## Options

| Option | Default | Description |
| --- | --- | --- |
| [`architecture.validate_cohesion`](../configuration/reference.md#architecture) | `true` | Set to `false` to disable this rule. |
| [`architecture.min_cohesion`](../configuration/reference.md#architecture) | `0.5` | Packages scoring below this are flagged. |
| [`architecture.enabled`](../configuration/reference.md#architecture) | `true` | Master switch for architecture analysis. |
| [`architecture.fail_on_violations`](../configuration/reference.md#architecture) | `false` | Non-zero exit code on violation. |

## References

- Package cohesion computation (`internal/analyzer/coupling_metrics.go`, `internal/analyzer/module_analyzer.go`).
- [Rule catalog](index.md) · [layer-violation](layer-violation.md)
