# Config File Format

pyscn reads configuration from **TOML**. You can keep settings in a dedicated `.pyscn.toml` file or in a `[tool.pyscn]` section of your existing `pyproject.toml`.

## File discovery

When you run `pyscn analyze` or `pyscn check`, pyscn walks up from the target path looking for:

1. `.pyscn.toml` (highest priority)
2. `pyproject.toml` with a `[tool.pyscn]` section

The first file found is used. Parent directories are searched until one matches or the filesystem root is reached. If neither file is found, built-in defaults are used.

You can also pass an explicit path:

```bash
pyscn analyze --config ./configs/strict.toml src/
```

This bypasses discovery.

## Priority order

When a setting appears in multiple places, later wins:

1. **Built-in defaults** (lowest)
2. **`pyproject.toml` → `[tool.pyscn]`**
3. **`.pyscn.toml`**
4. **CLI flags** (highest)

CLI flags are only considered if they were **explicitly set** — unchanged defaults don't override config values.

## Two file styles

=== ".pyscn.toml"

    ```toml
    [complexity]
    max_complexity = 15

    [dead_code]
    min_severity = "critical"
    ```

=== "pyproject.toml"

    ```toml
    [tool.pyscn.complexity]
    max_complexity = 15

    [tool.pyscn.dead_code]
    min_severity = "critical"
    ```

If both files exist in the same directory, `.pyscn.toml` wins.

## Generating a starter file

```bash
pyscn init
```

This writes a fully-commented `.pyscn.toml` with every option, its default, and a short description. Edit the values you care about and delete (or leave alone) the rest.

```bash
pyscn init --force   # overwrite existing
pyscn init --config tools/pyscn.toml   # custom path
```

## Validation

pyscn validates configuration on load and exits with code `2` if anything is wrong. Common validation rules:

- Complexity thresholds must satisfy `low ≥ 1` and `medium > low`.
- Output format must be one of `text`, `json`, `yaml`, `csv`, `html`.
- Dead-code severity must be `info`, `warning`, or `critical`.
- Clone similarity thresholds must be in `[0.0, 1.0]`.
- At least one include pattern must be specified.

## Environment variables

pyscn does **not** read configuration from environment variables. The MCP server uses one exception: `PYSCN_CONFIG` can point to a config file.

## Next steps

- [Reference](reference.md) — every key, documented.
- [Examples](examples.md) — strict CI, large codebase, minimal overrides.
