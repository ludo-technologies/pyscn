# Quick Start

## Run an analysis

```bash
pyscn analyze .
```

Writes an HTML report to `.pyscn/reports/analyze_YYYYMMDD_HHMMSS.html` and opens it in the default browser.

## Select output format

```bash
pyscn analyze --json .
pyscn analyze --yaml .
pyscn analyze --csv .
pyscn analyze --no-open .       # suppress browser open
```

## Run specific analyzers

```bash
pyscn analyze --select complexity .
pyscn analyze --select complexity,deadcode .
pyscn analyze --skip-clones .
```

See [`analyze`](../cli/analyze.md) for all flags.

## CI quality gate

```bash
pyscn check .                              # exit 0 pass, 1 fail
pyscn check --max-complexity 15 src/
pyscn check --select complexity,deadcode,deps src/
```

See [`check`](../cli/check.md) and [CI/CD Integration](../integrations/ci-cd.md).

## Generate a config file

```bash
pyscn init
```

Creates `.pyscn.toml` with every option commented. See the [Configuration Reference](../configuration/reference.md).
