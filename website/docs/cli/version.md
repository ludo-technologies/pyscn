# `pyscn version`

Print version and build information.

```text
pyscn version [flags]
```

## Flags

| Flag | Description |
| --- | --- |
| `-s, --short` | Print only the version number, for use in scripts. |

## Examples

```bash
$ pyscn version
pyscn v0.2.0
  commit:   a3671f4
  built:    2026-02-14T08:42:11Z
  go:       go1.24.6
  platform: darwin/arm64

$ pyscn version --short
0.2.0
```

## Use in scripts

```bash
# Save for later use
VERSION=$(pyscn version --short)

# Guard against incompatible versions
if [ "$(pyscn version --short)" != "0.2.0" ]; then
  echo "Expected pyscn 0.2.0" >&2
  exit 1
fi
```

`pyscn version` always exits with code `0`.
