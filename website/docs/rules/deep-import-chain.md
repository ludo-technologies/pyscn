# deep-import-chain

**Category**: Module Structure  
**Severity**: Info  
**Triggered by**: `pyscn analyze`, `pyscn check --select deps`

## What it does

Reports the longest acyclic import chain in the project when its depth exceeds the expected depth for a project of its size. pyscn uses `log₂(module_count) + 1` as the reference — a 64-module project is expected to have chains no longer than 7.

A chain is a path through the module dependency graph: `a → b → c → …`, where each arrow is an `import`.

## Why is this a problem?

Deep chains indicate poor layering. Every additional link is a module that must be loaded, parsed, and initialised before the bottom of the chain is usable, and every link is a place where an unrelated change can ripple downward.

Symptoms of a too-deep chain:

- **Slow startup.** Importing the leaf module triggers a cascade of top-level side effects.
- **Fragile tests.** A unit test for the leaf pulls in the full chain and breaks when anything upstream changes.
- **Hidden coupling.** Modules in the middle of the chain often exist only as pass-throughs, masking the real dependency.
- **Hard to reason about.** There is no single "level" at which the code lives.

## Example

```
myapp.cli
  → myapp.commands
    → myapp.services
      → myapp.orchestrator
        → myapp.workers
          → myapp.adapters
            → myapp.drivers
```

Seven levels to reach the driver. In practice the CLI layer doesn't need to know that workers exist, and workers don't need to know about the CLI — but a change to `drivers` can force re-testing every layer above it.

## Use instead

Introduce a facade at the boundary so upper layers talk to one module, not a chain:

```
myapp.cli
  → myapp.commands
    → myapp.services        # single entry point
        (internally wires orchestrator / workers / adapters / drivers)
```

Or flatten: if `services`, `orchestrator`, and `workers` are all doing coordination, merge them into one layer and let it depend directly on `adapters`.

## Options

| Option | Default | Description |
| --- | --- | --- |
| [`dependencies.find_long_chains`](../configuration/reference.md#dependencies) | `true` | Set to `false` to disable this rule. |
| [`dependencies.enabled`](../configuration/reference.md#dependencies) | `false` | Opt-in for `pyscn check`; always on for `pyscn analyze`. |

There is no explicit depth threshold — pyscn compares the longest chain against `log₂(module_count) + 1` and reports when it is exceeded.

## References

- Longest-path search over the module DAG (`internal/analyzer/module_analyzer.go`, `internal/analyzer/coupling_metrics.go`).
- [Rule catalog](index.md) · [circular-import](circular-import.md)
