---
hide:
  - navigation
  - toc
---

# pyscn

Analyseur statique structurel pour Python. Détecte le code mort, la duplication, la complexité et les problèmes de couplage via l'analyse du flux de contrôle et d'arbres.

```bash
uvx pyscn@latest analyze .
```

## Fonctionnalités

- **33 règles** couvrant le code inatteignable, le code dupliqué, la complexité, la conception de classes, l'injection de dépendances, la structure des modules et les données factices.
- **Accessibilité basée sur le CFG** qui détecte le code mort après `return` / `raise` / `break` / `continue` ainsi que les branches inatteignables.
- **Détection de clones APTED + LSH** sur les quatre types de clones (identique, renommé, modifié, sémantique).
- **Métriques de couplage et de cohésion de classes** CBO / LCOM4.
- **Détection des imports circulaires** via l'algorithme SCC de Tarjan.
- **Score de santé** (0–100) avec décomposition par catégorie.
- **Prêt pour la CI** grâce à `pyscn check`, une sortie de type linter et des codes de sortie déterministes.
- **Agent Skills** et **serveur MCP** (`pyscn-mcp`) pour Claude Code, Cursor et autres agents de codage IA.

Écrit en Go. Plus de 100 000 lignes/s sur du matériel courant. Aucune dépendance d'exécution Python.

## Installation

```bash
uvx pyscn@latest <command>   # exécution sans installation (recommandé)
uv tool install pyscn        # installation via uv
pipx install pyscn           # installation via pipx
pip install pyscn            # installation via pip
```

Voir [Installation](getting-started/installation.md) pour toutes les options.

## Démarrage rapide

```bash
pyscn analyze .                         # analyse complète, rapport HTML
pyscn check --select complexity,deadcode src/   # garde-fou CI
pyscn init                              # génère .pyscn.toml
```

Voir [Démarrage rapide](getting-started/quick-start.md) et le [Catalogue de règles](rules/index.md).

## Intégration avec les agents IA

```bash
uvx add-skills ludo-technologies/pyscn
```

Installe des Agent Skills qui apprennent à Claude Code, Cursor, Codex, Gemini CLI et autres agents de codage quand et comment lancer chaque analyse. Voir [Agent Skills](integrations/skills.md), ou utilisez le [serveur MCP](integrations/mcp.md) pour des appels d'outils structurés.
