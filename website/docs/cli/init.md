# `pyscn init`

Generate a `.pyscn.toml` configuration file with every option documented inline.

```text
pyscn init [flags]
```

## What it does

Writes a commented TOML file with the most commonly-tuned sections:

- `[output]`, `[complexity]`, `[dead_code]`, `[clones]`, `[cbo]`, `[analysis]`, `[architecture]` (plus example `[[architecture.layers]]` and `[[architecture.rules]]`)
- Default values filled in
- Comments explaining each key

The generated file does **not** include every configurable section. Options for LCOM4 cohesion (`[lcom]`), module dependency analysis (`[dependencies]`), mock-data detection (`[mock_data]`), and DI anti-patterns (`[di]`) are valid but must be added manually. See the [Configuration Reference](../configuration/reference.md) for every key.

Once the file exists, every subsequent `pyscn analyze` / `pyscn check` run in this project (or any subdirectory) picks it up automatically.

## Flags

| Flag | Default | Description |
| --- | --- | --- |
| `-c, --config <path>` | `.pyscn.toml` | Output file path. |
| `-f, --force`         | off          | Overwrite an existing file. |

## Exit codes

| Code | Meaning |
| --- | --- |
| `0` | File written successfully. |
| `1` | File already exists (use `--force` to overwrite) or write failed. |

## Examples

```bash
# Create .pyscn.toml in the current directory
pyscn init

# Use a custom filename
pyscn init --config tools/pyscn.toml

# Overwrite an existing config
pyscn init --force
```

## What to edit first

After running `init`, the settings most projects end up tuning are:

| Setting | Typical tuning |
| --- | --- |
| `[complexity].max_complexity` | Set to `10`, `15`, or `20` depending on how strict you want CI to be. |
| `[dead_code].min_severity`     | Raise to `"critical"` if warnings are too noisy. |
| `[clones].similarity_threshold`| Lower to `0.80` to find more clones, raise to `0.90` to reduce noise. |
| `[analysis].exclude_patterns`  | Add generated code paths, migrations, etc. |

See the full [Configuration Reference](../configuration/reference.md).

## See also

- [Configuration Reference](../configuration/reference.md) — all options explained.
- [Configuration Examples](../configuration/examples.md) — strict CI, large codebase, minimal overrides.
