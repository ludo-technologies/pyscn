# Python Packaging

pyscn is distributed on PyPI as a wheel containing a native Go binary. The Python layer is a stdlib launcher; there are no Python runtime dependencies.

## Which installer?

| Tool | Good for | Notes |
| --- | --- | --- |
| `uvx` (recommended) | One-off runs, CI | Runs without installing; caches after first call. |
| `uv tool install` | Persistent tool management | Fast, isolated. |
| `pipx` | Persistent CLI install | Isolated from project deps. |
| `pip` | Installing into a venv | No isolation. |

CI: `uvx pyscn@latest check .`. Local dev: `uv tool install pyscn` or `pipx install pyscn`.

## Platform support

| OS | Architectures |
| --- | --- |
| Linux | x86_64, arm64 |
| macOS | x86_64, arm64 |
| Windows | x86_64, arm64 |

Python 3.8–3.13.

## Packages

| Package | Contains | Install when |
| --- | --- | --- |
| `pyscn` | CLI + MCP server | You want the CLI. |
| `pyscn-mcp` | MCP server only | You only want the MCP server. |

## Versioning

[PEP 440](https://peps.python.org/pep-0440/), matching the Git tag:

- `0.1.0` — stable
- `0.2.0.dev1` — development
- `0.2.0b1` — beta

Pin for reproducibility:

```bash
pip install pyscn==0.2.0
```

## Containers

```dockerfile
FROM python:3.12-slim
RUN pip install --no-cache-dir pyscn
ENTRYPOINT ["pyscn"]
```

## Wheel contents

```
pyscn-0.2.0-py3-none-manylinux_2_17_x86_64.whl
├── pyscn/
│   ├── __init__.py
│   ├── __main__.py        # CLI launcher
│   ├── mcp_main.py        # MCP launcher
│   └── bin/
│       └── pyscn          # Go binary
```

The launcher detects OS + architecture and `exec`s the matching binary.

## Releases

Versions publish on Git tags starting with `v`:

```bash
git tag -a v0.2.0 -m "Release v0.2.0"
git push origin v0.2.0
```

GitHub Actions cross-compiles, packages platform-specific wheels, runs `twine check`, smoke-tests across OS × Python, publishes to PyPI, and creates a GitHub release. See the [Releases page](https://github.com/ludo-technologies/pyscn/releases).

## Alternatives to PyPI

- `go install github.com/ludo-technologies/pyscn/cmd/pyscn@latest` (Go 1.22+; does not install `pyscn-mcp`).
- Binary downloads from GitHub Releases.

## See also

- [Installation](../getting-started/installation.md)
- [CI/CD Integration](ci-cd.md)
