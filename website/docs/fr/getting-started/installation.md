# Installation

## Prérequis

- Python 3.8–3.13 (lanceur uniquement ; pyscn n'a aucune dépendance d'exécution Python)
- Linux / macOS / Windows, x86_64 ou arm64

## Installation

| Méthode | Commande | Remarques |
| --- | --- | --- |
| **uvx** (recommandé) | `uvx pyscn@latest <command>` | S'exécute sans installation ; mise en cache après le premier appel. |
| uv tool | `uv tool install pyscn` | Installation persistante, isolée des dépendances du projet. |
| pipx | `pipx install pyscn` | Installation persistante, isolée des dépendances du projet. |
| pip | `pip install pyscn` | Installe dans l'environnement courant. |
| Go | `go install github.com/ludo-technologies/pyscn/cmd/pyscn@latest` | N'installe pas `pyscn-mcp`. |

`uvx` est la voie la plus rapide pour un usage ponctuel et fonctionne bien en CI. Utilisez `uv tool install` ou `pipx` pour un usage local répété sans polluer les dépendances du projet.

Des binaires précompilés sont joints à chaque [release GitHub](https://github.com/ludo-technologies/pyscn/releases).

## Vérification

```bash
pyscn version
pyscn version --short    # uniquement le numéro de version
```

## Mise à jour

```bash
uv tool upgrade pyscn        # si installé avec uv tool
pipx upgrade pyscn           # si installé avec pipx
pip install --upgrade pyscn  # si installé avec pip
```

`uvx pyscn@latest` résout toujours vers la dernière version, donc aucune étape de mise à jour n'est nécessaire.

## Désinstallation

```bash
uv tool uninstall pyscn
pipx uninstall pyscn
pip uninstall pyscn
```
