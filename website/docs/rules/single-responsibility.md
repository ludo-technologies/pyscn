# single-responsibility

**Category**: Module Structure  
**Severity**: Warning (Error when the module is also a hub with very high fan-in/out)  
**Triggered by**: `pyscn analyze`, `pyscn check --select deps`

## What it does

Flags a module that mixes more than `architecture.max_responsibilities` (default `3`) distinct dependency concerns, or that acts as a fan-in/fan-out hub for more concerns than the rest of the project does on average.

A "concern" is inferred from the names of the module's neighbors: for each module the analyzer imports or that imports it, it takes the first segment of the neighbor's path that is not part of the current module's path and not a generic catch-all (`base`, `common`, `helpers`, `node`, `shared`, `util`, `utils`). Those segments are deduplicated; the count is the number of responsibilities pyscn attributes to the module.

A module is reported when either condition holds:

- It owns more than `max_responsibilities` distinct concerns.
- Both its fan-in (number of importers) and fan-out (number of imports) are above the project's `mean + standard deviation`, and it owns more than one concern.

## Why is this a problem?

The Single Responsibility Principle is about *axes of change*. A module that participates in several unrelated dependency clusters has multiple reasons to change:

- **Edits ripple.** Touching one concern forces re-reading and re-testing the rest, because they all share the same module boundary.
- **Imports lie.** `from myapp.core import X` tells the reader nothing — `core` is doing several jobs.
- **Hubs become bottlenecks.** A module that everyone imports *and* that imports everything is a single point of contention for changes, reviews, and merges.
- **Hides a missing seam.** When two concerns keep ending up in the same file, the right fix is usually a new module that names the relationship between them.

## Example

```
myapp/core.py
```

```python
# myapp/core.py
from myapp.routers import user_router, order_router
from myapp.services import billing_service, notification_service
from myapp.repositories import user_repo, order_repo
from myapp.telemetry import metrics, tracing

# ...glue code that pulls everything together...
```

`core` mixes four concerns (`routers`, `services`, `repositories`, `telemetry`), and it is imported by both routers and services elsewhere — fan-in and fan-out are both high. pyscn flags it as overloaded.

## Use instead

Split the module along the concerns that already exist. Each new module should name *one* axis of change.

```
myapp/wiring/web.py          # router-level wiring
myapp/wiring/services.py     # service-level wiring
myapp/wiring/persistence.py  # repository wiring
myapp/wiring/observability.py
```

Or, if the module is a legitimate composition root, narrow its scope: it should only *wire* the parts together, not also implement business rules, define types, or own telemetry.

## Options

| Option | Default | Description |
| --- | --- | --- |
| [`architecture.validate_responsibility`](../configuration/reference.md#architecture) | `true` | Set to `false` to disable this rule. |
| [`architecture.max_responsibilities`](../configuration/reference.md#architecture) | `3` | Modules owning more concerns than this are flagged. |
| [`architecture.enabled`](../configuration/reference.md#architecture) | `true` | Master switch for architecture analysis. |
| [`architecture.fail_on_violations`](../configuration/reference.md#architecture) | `false` | Non-zero exit code on violation. |

## References

- Responsibility inference and severity rules: `service/responsibility_analysis.go`.
- Martin, R. C. *Agile Software Development: Principles, Patterns, and Practices*, 2002 (Chapter 8 — SRP).
- [Rule catalog](index.md) · [low-package-cohesion](low-package-cohesion.md) · [layer-violation](layer-violation.md)
