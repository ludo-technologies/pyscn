# Architecture Style Presets

This document describes the **architecture style presets** that ship with pyscn:
named bundles of layer definitions and dependency rules that a project can opt
into with a single `style = "..."` setting instead of hand-writing the full
`[architecture.layers]` / `[architecture.rules]` configuration.

The feature was delivered across PRs #520–#523 (issue #335):

| PR    | Scope |
| ----- | ----- |
| #520  | Preset plumbing + the `layered` preset (the long-standing default, extracted into a named preset). |
| #521  | `hexagonal` and `clean` presets. |
| #522 / #523 | `mvc` preset + the warning-severity mechanism (`Warn` rules). #523 re-applies #522 cleanly on `main`. |

Implementation lives in `internal/config/architecture_presets.go` (the preset
definitions) and `service/system_analysis_service.go` (resolution and rule
evaluation). Domain types are in `domain/system_analysis.go`.

## Why presets

Layer-boundary analysis is only as good as the layer/rule configuration behind
it. Writing that configuration by hand is verbose and error-prone, and most
projects follow one of a handful of well-known architectural styles. Presets
encode those styles once so a project can declare its intent in one line:

```toml
[architecture]
style = "hexagonal"
```

A preset supplies both the **layers** (which package-name patterns map to which
layer) and the **rules** (which layer may depend on which). Users can still
override either: explicit `[[architecture.layers]]` / `[[architecture.rules]]`
take precedence over the preset (see [Resolution order](#resolution-order)).

## Available presets

### `layered` (default)

The backward-compatible baseline. Selecting `style = "layered"` reproduces the
behavior projects had before presets existed: the preset mirrors the
`[architecture]` section of `internal/config/default_config.toml.tmpl` exactly,
so the generated default config (which ships those layers/rules explicitly)
sees no change.

> **Note on the empty style.** `ArchitectureStylePreset("")` returns the
> `layered` definitions, but the resolver only consults a preset when `style`
> is non-empty. A project with **no** `style`, layers, or rules at all is
> auto-detected from its dependency graph, not silently given the `layered`
> preset — see [Resolution order](#resolution-order).

| Layer | Example packages |
| --- | --- |
| `presentation` | `router`, `handler`, `controller`, `view`, `api`, `web`, `rest`, `graphql` |
| `application` | `service`, `usecase`, `workflow`, `command`, `query`, `manager` |
| `domain` | `model`, `entity`, `schema`, `domain`, `core`, `aggregate`, `valueobject` |
| `infrastructure` | `repository`, `db`, `adapter`, `persistence`, `storage`, `cache`, `client` |

Rules (allow-based):

```
presentation   -> presentation, application, domain, infrastructure
application    -> application, domain, infrastructure
domain         -> domain, infrastructure        (deny: presentation, application)
infrastructure -> infrastructure, domain, application
```

> **Note:** `layered` intentionally allows `domain -> infrastructure`. This
> relaxes the strict Dependency Inversion Principle but preserves long-standing
> default behavior. Use `hexagonal` or `clean` for stricter DIP enforcement.

### `hexagonal`

Hexagonal / Onion architecture. Ports (interfaces) sit on the domain side;
adapters (implementations) sit on the infrastructure side. Enforces the
Dependency Inversion Principle: the domain depends on nothing outward.

| Layer | Example packages |
| --- | --- |
| `domain` | `domain`, `model`, `entity`, `core`, `aggregate`, `valueobject` |
| `ports` | `port`, `interface`, `contract`, `usecase`, `service` |
| `adapters` | `adapter`, `handler`, `controller`, `router`, `api`, `repository`, `db`, `client`, `web`, `rest`, `graphql` |

```
domain   -> domain                       (deny: ports, adapters)
ports    -> ports, domain                (deny: adapters)
adapters -> adapters, ports, domain
```

### `clean`

Clean Architecture's four concentric layers, innermost to outermost. The
dependency rule: source-code dependencies point only **inward** — each layer
may depend on itself and any inner layer, never an outer one.

| Layer | Example packages |
| --- | --- |
| `entities` | `entity`, `model`, `domain`, `core`, `aggregate`, `valueobject` |
| `use_cases` | `usecase`, `interactor`, `service`, `workflow`, `command`, `query` |
| `interface_adapters` | `adapter`, `controller`, `presenter`, `gateway`, `repository`, `router` |
| `frameworks` | `framework`, `infrastructure`, `db`, `web`, `api`, `external`, `client`, `persistence`, `ui` |

```
entities           -> entities                                                (deny: use_cases, interface_adapters, frameworks)
use_cases          -> use_cases, entities                                     (deny: interface_adapters, frameworks)
interface_adapters -> interface_adapters, use_cases, entities                 (deny: frameworks)
frameworks         -> frameworks, interface_adapters, use_cases, entities
```

### `mvc`

MVC / MVT. Template engines (MVT) are treated as views. Unlike the other
presets, `mvc` exercises the **warning** severity: a direct `view -> model`
dependency is permitted but discouraged rather than denied outright.

| Layer | Example packages |
| --- | --- |
| `model` | `model`, `entity`, `schema`, `domain` |
| `view` | `view`, `template`, `serializer`, `form` |
| `controller` | `controller`, `handler`, `router`, `route`, `viewset` |

```
model      -> model                      (deny: view, controller)
view       -> view, controller           (warn: model)        <-- discouraged, not denied
controller -> controller, model, view
```

## Rule semantics: `allow`, `deny`, `warn`

Each `LayerRule` (see `domain.LayerRule`) lists, for a given source layer
(`From`), the target layers it may reach. Three lists are checked in order by
`evaluateLayerEdge` (`service/system_analysis_service.go`):

1. **`Deny`** — takes precedence. A denied edge is an **error**-severity
   violation. Rule marker: `from !> to`.
2. **`Warn`** — checked *between* deny and allow. A warn-listed edge is
   permitted but emits a **warning**-severity violation, and the check stops
   here so the edge is not also reported as a disallowed error. Rule marker:
   `from ~> to`.
3. **`Allow`** — if a non-empty allow list is present and the target is not in
   it (and not already handled by deny/warn), the edge is an **error**.
   Rule marker: `from -> {allowed,...}`.

The `Warn` list (`Warn []string` on `domain.LayerRule`, `config.LayerRule`, and
`LayerRuleToml`) was added in PR #522/#523 specifically to model the "tolerated
but suboptimal" `view -> model` case. The same field is available to
hand-written configs, not just the `mvc` preset.

### Effect on the compliance score

Severity feeds the architecture compliance score via
`calculateComplianceWeighted`: an **error** is weighted 5 points, a **warning**
1 point, against the number of checked edges:

```
compliance = 1 - (5*errors + 1*warnings) / checked   (clamped to [0, 1])
```

So a `view ~> model` warning nudges the score down slightly rather than
counting as a full boundary breach. (In the #523 end-to-end check, a single
`view ~> model` warning yielded a compliance of `0.929`.)

## Resolution order

`resolveArchitectureRules` decides what layers/rules actually run:

1. **No config at all** (no layers, no rules, no style) → fully auto-detect the
   architecture from the dependency graph.
2. **A recognized `style`** → apply its preset layers/rules, then:
   - If the user defined their own `[[architecture.layers]]`, those win and the
     preset rules are *filtered* to the user's layer names.
   - If the user defined their own `[[architecture.rules]]`, they are *merged*
     on top of the preset — user rules win for any matching `From` value.
3. **An unrecognized `style` with no explicit config** → fall back to
   auto-detection.
4. **Explicit layers/rules without a style** → behave as before presets existed
   (load default rules filtered to the user's layers, or auto-detect layers for
   user-supplied rules).

The resolver clones the caller's `ArchitectureRules` before mutating, so the
incoming config object is never modified.

`ArchitectureStylePreset(style)` returns `(nil, nil)` for an unrecognized
style, which is how callers distinguish "fall back to auto-detect" from "apply
this preset."

## Configuration

In `.pyscn.toml` (or `pyproject.toml` under `[tool.pyscn.architecture]`):

```toml
[architecture]
enabled = true
style = "clean"        # one of: layered, hexagonal, clean, mvc (empty → auto-detect)
strict_mode = true
```

Mixing a preset with overrides — keep the `mvc` layers but tighten `view`:

```toml
[architecture]
style = "mvc"

# User rule wins for From = "view": now an error instead of a warning.
[[architecture.rules]]
from = "view"
allow = ["view", "controller"]
deny  = ["model"]
```

## Testing

Preset behavior is covered by:

- **Unit** (`internal/config`): `TestArchitectureStylePreset_*` —
  `HexagonalDomainIsolated`, `CleanEntitiesIsolated`, `MVCViewWarnsOnModel`,
  `AllStylesNonEmpty`.
- **Unit** (`service`): `TestEvaluateLayerEdge` Warn subtests.
- **Fixtures** (`testdata/`): `hexagonal_ports/`, `clean_layers/`, `mvc_app/`,
  each with one intentional violation.
- **Integration**: `TestArchitecture_HexagonalPreset`,
  `TestArchitecture_CleanPreset`, `TestArchitecture_MVCPreset` — each asserts
  the intended violation is flagged (the MVC case specifically as a *warning*)
  while allowed edges are not.

## See also

- [`dependency.md`](dependency.md) — module dependency graph construction and
  the metrics the architecture layer builds on.
- `internal/config/architecture_presets.go` — the canonical preset definitions.
