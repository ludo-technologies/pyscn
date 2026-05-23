# Empaquetage Python

pyscn est distribué sur PyPI sous forme de wheel contenant un binaire Go natif. La couche Python est un lanceur basé sur la bibliothèque standard ; il n'y a aucune dépendance Python d'exécution.

## Quel installateur choisir ?

| Outil | Adapté à | Notes |
| --- | --- | --- |
| `uvx` (recommandé) | Exécutions ponctuelles, CI | S'exécute sans installation ; mise en cache après le premier appel. |
| `uv tool install` | Gestion d'outils persistante | Rapide et isolé. |
| `pipx` | Installation persistante d'un CLI | Isolé des dépendances du projet. |
| `pip` | Installation dans un venv | Pas d'isolation. |

CI : `uvx pyscn@latest check .`. Développement local : `uv tool install pyscn` ou `pipx install pyscn`.

## Plateformes prises en charge

| OS | Architectures |
| --- | --- |
| Linux | x86_64, arm64 |
| macOS | x86_64, arm64 |
| Windows | x86_64, arm64 |

Python 3.8–3.13.

## Paquets

| Paquet | Contient | À installer quand |
| --- | --- | --- |
| `pyscn` | CLI + serveur MCP | Vous voulez le CLI. |
| `pyscn-mcp` | Serveur MCP uniquement | Vous ne voulez que le serveur MCP. |

## Versionnage

[PEP 440](https://peps.python.org/pep-0440/), correspondant à l'étiquette Git :

- `0.1.0` — stable
- `0.2.0.dev1` — développement
- `0.2.0b1` — bêta

Figez la version pour la reproductibilité :

```bash
pip install pyscn==0.2.0
```

## Conteneurs

```dockerfile
FROM python:3.12-slim
RUN pip install --no-cache-dir pyscn
ENTRYPOINT ["pyscn"]
```

## Contenu du wheel

```
pyscn-0.2.0-py3-none-manylinux_2_17_x86_64.whl
├── pyscn/
│   ├── __init__.py
│   ├── __main__.py        # CLI launcher
│   ├── mcp_main.py        # MCP launcher
│   └── bin/
│       └── pyscn          # Go binary
```

Le lanceur détecte l'OS et l'architecture puis appelle `exec` sur le binaire correspondant.

## Versions

Les versions sont publiées sur les étiquettes Git commençant par `v` :

```bash
git tag -a v0.2.0 -m "Release v0.2.0"
git push origin v0.2.0
```

GitHub Actions compile en croisé, fabrique des wheels spécifiques à chaque plateforme, exécute `twine check`, lance des tests de fumée à travers la matrice OS × Python, publie sur PyPI et crée une release GitHub. Voir la [page des Releases](https://github.com/ludo-technologies/pyscn/releases).

## Alternatives à PyPI

- `go install github.com/ludo-technologies/pyscn/cmd/pyscn@latest` (Go 1.22+ ; n'installe pas `pyscn-mcp`).
- Téléchargements binaires depuis GitHub Releases.

## Voir aussi

- [Installation](../getting-started/installation.md)
- [Intégration CI/CD](ci-cd.md)
