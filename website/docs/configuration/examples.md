# Configuration Examples

Copy-paste starting points for common scenarios.

## Minimal override

Just a few strict thresholds; everything else stays at default.

```toml
# .pyscn.toml
[complexity]
max_complexity = 15

[dead_code]
min_severity = "critical"
```

## Strict CI gate

Fail the build on any quality regression. Pair with `pyscn check`.

```toml
[complexity]
max_complexity = 10

[dead_code]
min_severity = "warning"
detect_after_return = true
detect_after_raise = true
detect_unreachable_branches = true

[clones]
# Only flag near-identical code
similarity_threshold = 0.90
min_lines = 15

[cbo]
medium_threshold = 7

[dependencies]
enabled = true
detect_cycles = true
```

Run:

```bash
pyscn check --select complexity,deadcode,deps --max-cycles 0 src/
```

## Legacy codebase grace period

You're adopting pyscn on an older project — you want signal without an immediate deluge of failures.

```toml
[complexity]
max_complexity = 25    # allow existing complexity

[dead_code]
min_severity = "critical"   # only the worst

[clones]
min_lines = 20              # only long-form duplicates
similarity_threshold = 0.90

[analysis]
exclude_patterns = [
  "legacy/**",     # quarantine old code
  "**/_archive/*",
  "generated/**",
]
```

Tighten the thresholds gradually over time.

## Large codebase (10k+ files)

Optimize for throughput. LSH auto-enables, but crank up parallelism.

```toml
[clones]
lsh_enabled = true
max_goroutines = 16
max_memory_mb = 2048
batch_size = 500
timeout_seconds = 600
min_lines = 15           # fewer, more meaningful candidates

[analysis]
exclude_patterns = [
  "**/test_*.py", "**/*_test.py",
  "**/migrations/**",
  "**/__generated__/**",
  "**/node_modules/**",
  ".venv/**", "venv/**",
]
```

## Clean architecture validation

Enforce a layered architecture: presentation → application → domain, infrastructure at the edge.

```toml
[architecture]
enabled = true
strict_mode = true
fail_on_violations = true

[[architecture.layers]]
name = "presentation"
packages = ["api", "routers", "handlers", "views"]

[[architecture.layers]]
name = "application"
packages = ["services", "usecases"]

[[architecture.layers]]
name = "domain"
packages = ["models", "entities", "core"]

[[architecture.layers]]
name = "infrastructure"
packages = ["repositories", "db", "adapters", "clients"]

[[architecture.rules]]
from = "presentation"
allow = ["application", "domain"]
deny = ["infrastructure"]

[[architecture.rules]]
from = "application"
allow = ["domain", "infrastructure"]
deny = ["presentation"]

[[architecture.rules]]
from = "domain"
deny = ["presentation", "application", "infrastructure"]
```

## Data-heavy ML / research codebase

High complexity is expected in notebooks-turned-modules. Focus on duplication and dead code.

```toml
[complexity]
max_complexity = 30    # data pipelines are naturally branchy

[dead_code]
min_severity = "critical"

[clones]
# Research code often has near-identical experiment variants;
# raise thresholds so you don't get flooded
min_lines = 20
similarity_threshold = 0.85

[analysis]
exclude_patterns = [
  "notebooks/**",
  "experiments/**/*.ipynb",
]
```

## Coexisting with `pyproject.toml`

If you already have a `pyproject.toml`, you can put pyscn config there instead of creating a new file:

```toml
# pyproject.toml
[project]
name = "my-package"
# ... other project metadata

[tool.pyscn.complexity]
max_complexity = 15

[tool.pyscn.dead_code]
min_severity = "critical"

[tool.pyscn.clones]
similarity_threshold = 0.85
```

!!! note
    `.pyscn.toml` takes priority over `pyproject.toml` if both exist. Pick one to avoid confusion.

## See also

- [Config File Format](format.md)
- [Configuration Reference](reference.md)
