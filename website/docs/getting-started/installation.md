# Installation

## Requirements

- Python 3.8–3.13 (launcher only; pyscn has no Python runtime dependencies)
- Linux / macOS / Windows, x86_64 or arm64

## Install

| Method | Command | Notes |
| --- | --- | --- |
| **uvx** (recommended) | `uvx pyscn@latest <command>` | Runs without installing; caches after first call. |
| uv tool | `uv tool install pyscn` | Persistent install, isolated from project deps. |
| pipx | `pipx install pyscn` | Persistent install, isolated from project deps. |
| pip | `pip install pyscn` | Installs into the current environment. |
| Go | `go install github.com/ludo-technologies/pyscn/cmd/pyscn@latest` | Does not install `pyscn-mcp`. |

`uvx` is the fastest path for one-off use and works well in CI. Use `uv tool install` or `pipx` for repeated local use without polluting project dependencies.

Pre-built binaries are attached to every [GitHub release](https://github.com/ludo-technologies/pyscn/releases).

## Verify

```bash
pyscn version
pyscn version --short    # just the version number
```

## Upgrade

```bash
uv tool upgrade pyscn        # if installed with uv tool
pipx upgrade pyscn           # if installed with pipx
pip install --upgrade pyscn  # if installed with pip
```

`uvx pyscn@latest` always resolves to the latest version, so no upgrade step is needed.

## Uninstall

```bash
uv tool uninstall pyscn
pipx uninstall pyscn
pip uninstall pyscn
```
