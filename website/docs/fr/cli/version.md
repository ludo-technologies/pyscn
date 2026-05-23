# `pyscn version`

Affiche les informations de version et de build.

```text
pyscn version [flags]
```

## Options

| Option | Description |
| --- | --- |
| `-s, --short` | Affiche uniquement le numéro de version, pour utilisation dans des scripts. |

## Exemples

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

## Utilisation dans des scripts

```bash
# Sauvegarder pour utilisation ultérieure
VERSION=$(pyscn version --short)

# Se prémunir contre des versions incompatibles
if [ "$(pyscn version --short)" != "0.2.0" ]; then
  echo "Expected pyscn 0.2.0" >&2
  exit 1
fi
```

`pyscn version` se termine toujours avec le code `0`.
