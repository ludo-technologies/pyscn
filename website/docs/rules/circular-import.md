# circular-import

**Category**: Module Structure  
**Severity**: Configurable by cycle size (Low / Medium / High / Critical)  
**Triggered by**: `pyscn analyze`, `pyscn check --select circular`

## What it does

Flags groups of modules that form an import cycle — module A imports B (directly or transitively) and B imports A. Cycles are found by running Tarjan's strongly-connected-components algorithm over the module dependency graph.

Severity is assigned from the cycle's size and the fan-in of its members:

| Cycle members | Severity |
| --- | --- |
| 2 | Low |
| 3 – 5 | Medium |
| 6 – 9 | High |
| 10+, or any member with fan-in > 10 | Critical |

## Why is this a problem?

A circular import means two or more modules cannot be understood, tested, or released independently. Concretely:

- **Import-time errors.** Python partially initialises modules during circular imports; attribute access on the half-loaded module raises `ImportError` or `AttributeError` depending on statement order.
- **Tight coupling.** The cycle's members share a single "logical module" split across files. A change in one tends to force a change in all of them.
- **Blocked refactoring.** You cannot move, rename, or delete any member of the cycle without touching the others.
- **Worse as the cycle grows.** A 2-module cycle is a nuisance; a 10-module cycle is an architectural failure — hence the severity ramp.

## Example

```python
# myapp/orders.py
from myapp.billing import Invoice

class Order:
    def invoice(self) -> Invoice:
        return Invoice(self)
```

```python
# myapp/billing.py
from myapp.orders import Order

class Invoice:
    def __init__(self, order: "Order"):
        self.order = order
```

`orders` imports `billing` for the return type; `billing` imports `orders` for the constructor parameter. Running either file top-level triggers the cycle.

## Use instead

Extract the shared types to a third module so both depend on it instead of on each other:

```python
# myapp/domain.py
class Order: ...
class Invoice: ...
```

```python
# myapp/orders.py
from myapp.domain import Order, Invoice
```

```python
# myapp/billing.py
from myapp.domain import Order, Invoice
```

If the back-edge is only needed for type annotations, guard it with `TYPE_CHECKING` so it isn't evaluated at runtime:

```python
# myapp/billing.py
from __future__ import annotations
from typing import TYPE_CHECKING

if TYPE_CHECKING:
    from myapp.orders import Order

class Invoice:
    def __init__(self, order: "Order"): ...
```

## Options

| Option | Default | Description |
| --- | --- | --- |
| [`dependencies.detect_cycles`](../configuration/reference.md#dependencies) | `true` | Set to `false` to disable this rule. |
| [`dependencies.cycle_reporting`](../configuration/reference.md#dependencies) | `"summary"` | `all`, `critical`, or `summary` — controls how many cycles appear in the report. |
| [`dependencies.max_cycles_to_show`](../configuration/reference.md#dependencies) | `10` | Cap on reported cycles. |
| `--max-cycles N` (check) | `0` | Fail the `check` command when cycle count exceeds `N`. |
| `--allow-circular-deps` (check) | off | Demote cycles to warnings instead of failures. |

## References

- Tarjan SCC implementation (`internal/analyzer/circular_detector.go`), module graph construction (`internal/analyzer/module_analyzer.go`).
- [Rule catalog](index.md) · [deep-import-chain](deep-import-chain.md)
