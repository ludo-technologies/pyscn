---
hide:
  - navigation
  - toc
---

<div class="pyscn-hero" markdown="1">

<div class="pyscn-hero__copy" markdown="1">

<p class="pyscn-hero__eyebrow">Analyse statique structurelle pour Python</p>

# pyscn

<p class="pyscn-hero__lede">pyscn lit Python comme un compilateur — graphes de flux de contrôle, arbres syntaxiques, graphes d'imports. C'est ainsi qu'il détecte ce que les linters ligne par ligne ne voient pas : du code mort laissé après un <code>return</code>, de la logique dupliquée sous un autre nom, et des modules pris dans des cycles silencieux.</p>

```bash
uvx pyscn@latest analyze .
```

[Commencer :material-arrow-right:](getting-started/quick-start.md){ .md-button .md-button--primary } [Voir sur GitHub :fontawesome-brands-github:](https://github.com/ludo-technologies/pyscn){ .md-button }

<p class="pyscn-hero__meta">Binaire Go · aucune dépendance runtime Python · 100 000+ lignes/s · 33 règles</p>

</div>

--8<-- "includes/cfg-diagram.html"

</div>

## Ce qu'il détecte

<div class="grid cards" markdown>

-   :material-source-branch:{ .lg .middle } __Code inatteignable__

    ---

    L'analyse d'accessibilité basée sur le CFG détecte le code mort laissé après `return`, `raise`, `break`, `continue`, ou après une branche toujours vraie.

-   :material-content-duplicate:{ .lg .middle } __Code dupliqué__

    ---

    La distance d'édition d'arbres APTED combinée au LSH détecte quatre types de clones : identique, renommé, modifié, sémantique.

-   :material-gauge:{ .lg .middle } __Complexité__

    ---

    Complexité cyclomatique par fonction, avec des seuils ajustables par projet.

-   :material-shape-outline:{ .lg .middle } __Conception de classes__

    ---

    Les métriques de couplage CBO et de cohésion LCOM4 révèlent les classes qui font trop, ou trop peu, ensemble.

-   :material-sync:{ .lg .middle } __Imports circulaires__

    ---

    L'algorithme SCC de Tarjan détecte les cycles d'imports avant qu'ils ne provoquent une `ImportError` à l'exécution.

-   :material-sitemap:{ .lg .middle } __Structure des modules__

    ---

    Le clustering de Leiden sur le graphe d'imports révèle quels modules devraient vraiment être ensemble — et lesquels non.

</div>

## Installation

=== "uvx (recommandé)"

    ```bash
    uvx pyscn@latest analyze .
    ```

    Exécute la dernière version sans rien installer.

=== "uv"

    ```bash
    uv tool install pyscn
    ```

=== "pipx"

    ```bash
    pipx install pyscn
    ```

=== "pip"

    ```bash
    pip install pyscn
    ```

Voir [Installation](getting-started/installation.md) pour toutes les options.

## Démarrage rapide

```bash
pyscn analyze .                                  # analyse complète, rapport HTML
pyscn check --select complexity,deadcode src/    # garde-fou CI
pyscn init                                       # génère .pyscn.toml
```

Voir [Démarrage rapide](getting-started/quick-start.md) et le [Catalogue de règles](rules/index.md).

## Intégration avec les agents IA

```bash
uvx add-skills ludo-technologies/pyscn
```

Installe des Agent Skills qui apprennent à Claude Code, Cursor, Codex, Gemini CLI et autres agents de codage quand et comment lancer chaque analyse. Voir [Agent Skills](integrations/skills.md), ou utilisez le [serveur MCP](integrations/mcp.md) pour des appels d'outils structurés.
