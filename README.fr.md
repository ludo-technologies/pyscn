<div align="center">

[English](README.md) | [日本語](README.ja.md) | [简体中文](README.zh-CN.md) | [Français](README.fr.md)

<br>

<picture>
  <source media="(prefers-color-scheme: dark)" srcset="assets/logo.svg">
  <source media="(prefers-color-scheme: light)" srcset="assets/logo-light.svg">
  <img alt="pyscn" src="assets/logo-light.svg" width="320">
</picture>

**Un analyseur de qualité de code pour les vibe coders Python.**

Vous développez avec Cursor, Claude ou ChatGPT ? pyscn effectue une analyse structurelle pour maintenir la maintenabilité de votre codebase.

[![Article](https://img.shields.io/badge/dev.to-Article-0A0A0A?style=flat-square&logo=dev.to)](https://dev.to/daisukeyoda/pyscn-the-code-quality-analyzer-for-vibe-coders-18hk)
[![PyPI](https://img.shields.io/pypi/v/pyscn?style=flat-square&logo=pypi)](https://pypi.org/project/pyscn/)
[![Downloads](https://img.shields.io/pypi/dm/pyscn?style=flat-square&logo=pypi&label=downloads)](https://pypi.org/project/pyscn/)
[![Go](https://img.shields.io/github/go-mod/go-version/ludo-technologies/pyscn?style=flat-square&logo=go)](https://go.dev/)
[![License](https://img.shields.io/github/license/ludo-technologies/pyscn?style=flat-square)](LICENSE)

*Vous travaillez avec JavaScript/TypeScript ? Découvrez [jscan](https://github.com/ludo-technologies/jscan)*

</div>

## Démarrage rapide

```bash
# Lancer l'analyse sans installation
uvx pyscn@latest analyze .
# ou
pipx run pyscn analyze .
```

## Démo

https://github.com/user-attachments/assets/71d7a126-9c5e-4254-99f4-f2cdedd526ad

## Fonctionnalités

- 🔍 **Détection de code mort basée sur le CFG** – Trouve le code inatteignable après des chaînes if-elif-else exhaustives
- 📋 **Détection de clones multi-algorithmes (Type 1-4)** – Identifie les opportunités de refactorisation avec accélération LSH
- 🔗 **Métriques de couplage (CBO)** – Suivez la qualité architecturale et les dépendances entre modules
- 📊 **Analyse de complexité cyclomatique** – Repérez les fonctions à découper

**100 000+ lignes/s** • Construit avec Go + tree-sitter

## Intégration MCP

Exécutez les analyses pyscn directement depuis vos assistants de codage IA via le Model Context Protocol (MCP). Le serveur `pyscn-mcp` intégré expose les mêmes outils que le CLI à Claude Code, Cursor, ChatGPT et autres clients MCP.

### Cas d'usage MCP

Vous pouvez interagir avec pyscn via vos outils de codage IA :

1. « Analyse la qualité du code du répertoire app/ »

2. « Trouve le code dupliqué et aide-moi à le refactoriser »

3. « Montre-moi le code complexe et aide-moi à le simplifier »

### Configuration Claude Code

**Option 1 : Installation via le marketplace de plugins (recommandé)**

```bash
claude plugin marketplace add ludo-technologies/pyscn
claude plugin install pyscn-mcp@pyscn-marketplace
```

**Option 2 : Configuration MCP manuelle**

```bash
claude mcp add pyscn-mcp uvx -- pyscn-mcp
```

### Configuration Cursor / Claude Desktop

Ajoutez à vos paramètres MCP (`~/.config/claude-desktop/config.json` ou les paramètres Cursor) :

```json
{
  "mcpServers": {
    "pyscn-mcp": {
      "command": "uvx",
      "args": ["pyscn-mcp"],
      "env": {
        "PYSCN_CONFIG": "/path/to/.pyscn.toml"
      }
    }
  }
}
```

Des instructions comme « Analyse la qualité du code » déclenchent pyscn via MCP.

Consultez `mcp/README.md` pour les guides de configuration et `docs/MCP_INTEGRATION.md` pour les détails d'architecture.

## Installation

```bash
# Installer avec pipx (recommandé)
pipx install pyscn

# Ou avec uv
uv tool install pyscn
```

<details>
<summary>Autres méthodes d'installation</summary>

### Compilation depuis les sources
```bash
git clone https://github.com/ludo-technologies/pyscn.git
cd pyscn
make build
```

### Go install
```bash
go install github.com/ludo-technologies/pyscn/cmd/pyscn@latest
```

</details>

## Commandes courantes

### `pyscn analyze`
Lance une analyse complète avec rapport HTML
```bash
pyscn analyze .                              # Toutes les analyses avec rapport HTML
pyscn analyze --json .                       # Générer un rapport JSON
pyscn analyze --select complexity .          # Analyse de complexité uniquement
pyscn analyze --select deps .                # Analyse de dépendances uniquement
pyscn analyze --select complexity,deps,deadcode . # Analyses multiples
```

### `pyscn check`
Garde-fou qualité rapide pour la CI
```bash
pyscn check .                         # Vérification rapide réussite/échec
pyscn check --max-complexity 15 .     # Seuils personnalisés
pyscn check --max-cycles 0 .          # Autoriser uniquement 0 dépendance cyclique
pyscn check --select deps .           # Vérifier uniquement les dépendances circulaires
pyscn check --allow-circular-deps .   # Autoriser les dépendances circulaires (avertissement uniquement)
```

### `pyscn init`
Créer un fichier de configuration
```bash
pyscn init                         # Générer .pyscn.toml
```

> 💡 Exécutez `pyscn --help` ou `pyscn <command> --help` pour toutes les options

## Configuration

Créez un fichier `.pyscn.toml` ou ajoutez `[tool.pyscn]` à votre `pyproject.toml` :

```toml
# .pyscn.toml
[complexity]
max_complexity = 15

[dead_code]
min_severity = "warning"

[output]
directory = "reports"
```

> ⚙️ Exécutez `pyscn init` pour générer un fichier de configuration complet avec toutes les options disponibles

## Pyscn Bot (GitHub App)

[Pyscn Bot](https://github.com/marketplace/pyscn-bot) surveille automatiquement la qualité de votre code Python.

### Fonctionnalités

- **Revue de code sur les PR** - Revue automatique sur chaque pull request
- **Audit hebdomadaire du code** - Analyse l'ensemble du dépôt et crée des issues pour les problèmes architecturaux

---

## Documentation

📖 **[Site de documentation pyscn](https://ludo-technologies.github.io/pyscn/fr/)** — installation, catalogue de règles, référence CLI, configuration, spécification des sorties

Pour les contributeurs : **[Guide de développement](docs/DEVELOPMENT.md)** • **[Architecture](docs/ARCHITECTURE.md)** • **[Tests](docs/TESTING.md)**

## Support entreprise

Pour un support commercial, des intégrations personnalisées ou des services de conseil, contactez-nous à contact@ludo-tech.org

## Licence

MIT License — voir [LICENSE](LICENSE)

---

*Construit avec ❤️ en utilisant Go et tree-sitter*