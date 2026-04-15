# Installation

## Requirements

- Python 3.8–3.13 (launcher only; pyscn has no Python runtime dependencies)
- Linux / macOS / Windows, x86_64 or arm64

## Install

| Method | Command |
| --- | --- |
| pip | `pip install pyscn` |
| pipx | `pipx install pyscn` |
| uv | `uv tool install pyscn` |
| uvx (no install) | `uvx pyscn@latest <command>` |
| Go | `go install github.com/ludo-technologies/pyscn/cmd/pyscn@latest` |

`pipx`, `uv tool install`, and `uvx` isolate pyscn from project dependencies. The Go install path does not include `pyscn-mcp`.

Pre-built binaries are attached to every [GitHub release](https://github.com/ludo-technologies/pyscn/releases).

## Verify

```bash
pyscn version
pyscn version --short    # just the version number
```

## Upgrade

```bash
pip install --upgrade pyscn
pipx upgrade pyscn
uv tool upgrade pyscn
```

## Uninstall

```bash
pip uninstall pyscn
pipx uninstall pyscn
uv tool uninstall pyscn
```
