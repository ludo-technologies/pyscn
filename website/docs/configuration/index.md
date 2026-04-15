# Configuration

pyscn reads configuration from `.pyscn.toml` or `[tool.pyscn]` in `pyproject.toml`. CLI flags override config; config overrides built-in defaults.

- **[Config File Format](format.md)** — Discovery rules, file locations, priority order.
- **[Reference](reference.md)** — Every configuration key, type, default, and effect.
- **[Examples](examples.md)** — Strict CI, large codebase, minimal overrides.

Generate a commented starter file with:

```bash
pyscn init
```
