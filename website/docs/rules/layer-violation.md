# layer-violation

**Category**: Module Structure  
**Severity**: Configurable via `architecture.rules[].severity`  
**Triggered by**: `pyscn analyze`, `pyscn check --select deps`

## What it does

Flags an `import` statement when the source module's layer is not permitted to depend on the target module's layer, according to the `[[architecture.rules]]` you configured. Layers are assigned to modules by matching package name fragments defined in `[[architecture.layers]]`.

## Why is this a problem?

A layered architecture only pays off while the layers hold. A single shortcut from `presentation` into `infrastructure` is enough to:

- **Defeat testability.** The presentation layer can now only be exercised with a real database / HTTP client behind it.
- **Create hidden coupling.** Swapping the infrastructure implementation silently breaks UI code that was never supposed to know it existed.
- **Normalise the violation.** Once one shortcut exists, the next one is easier to justify.

The rule is the automated enforcement of the architecture diagram you already drew in a design doc.

## Example

Config:

```toml
[[architecture.layers]]
name = "presentation"
packages = ["api", "handlers"]

[[architecture.layers]]
name = "application"
packages = ["services", "usecases"]

[[architecture.layers]]
name = "infrastructure"
packages = ["repositories", "db"]

[[architecture.rules]]
from = "presentation"
allow = ["application"]
deny = ["infrastructure"]
```

Violating code:

```python
# myapp/api/orders.py  (presentation)
from myapp.repositories.orders import OrderRepository   # ← forbidden

def list_orders():
    return OrderRepository().all()
```

`presentation` reaches past `application` straight into `infrastructure`.

## Use instead

Route the call through the application layer:

```python
# myapp/services/orders.py  (application)
from myapp.repositories.orders import OrderRepository

def list_orders():
    return OrderRepository().all()
```

```python
# myapp/api/orders.py  (presentation)
from myapp.services.orders import list_orders

def get():
    return list_orders()
```

`api` now depends only on `services`, and the infrastructure is replaceable without touching the presentation layer.

## Options

| Option | Default | Description |
| --- | --- | --- |
| [`[[architecture.layers]]`](../configuration/reference.md#architecture) | — | Defines layers and the package fragments that belong to each. |
| [`[[architecture.rules]]`](../configuration/reference.md#architecture) | — | `from` / `allow` / `deny` / optional `severity` per rule. |
| [`architecture.validate_layers`](../configuration/reference.md#architecture) | `true` | Set to `false` to disable this rule. |
| [`architecture.strict_mode`](../configuration/reference.md#architecture) | `true` | In strict mode, anything not explicitly allowed is denied. |
| [`architecture.fail_on_violations`](../configuration/reference.md#architecture) | `false` | Non-zero exit code when a violation is found. |

With no layers configured, the analyzer runs in permissive mode and this rule produces no findings.

## References

- Layer resolution and rule evaluation (`internal/analyzer/module_analyzer.go`).
- [Rule catalog](index.md) · [low-package-cohesion](low-package-cohesion.md)
