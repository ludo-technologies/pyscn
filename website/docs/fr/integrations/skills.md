# Intégration Agent Skills

pyscn est livré avec 4 Agent Skills qui apprennent aux agents de codage quand et comment lancer chaque analyse, sans avoir besoin de configurer un serveur MCP.

## Installation

```bash
uvx add-skills ludo-technologies/pyscn
```

Installe les Skills dans votre projet. Compatible avec Claude Code, Cursor, Codex, Gemini CLI et [de nombreux autres agents](https://github.com/ludo-technologies/add-skills) (ajoutez `--agent cursor` pour cibler un agent en particulier, `--global` pour tous les projets).

## Liste des Skills

| Skill | À utiliser quand |
| --- | --- |
| `health-check` | « Ce code est-il sain ? », un aperçu de la qualité, une comparaison avant/après |
| `refactoring` | Trouver des cibles de refactoring — code dupliqué, points chauds de complexité, code mort |
| `architecture-review` | Structure des modules, couplage, dépendances circulaires, fichiers à revoir ensemble |
| `cli-analysis` | Garde-fous CI/CD, rapports partageables, configuration du projet |

Chaque Skill exécute en interne `uvx pyscn@latest <command>` — aucune installation préalable n'est nécessaire.

## Exemples d'invites

> Analyse la qualité du code du répertoire app/

> Trouve le code dupliqué et aide-moi à le refactoriser

> Montre-moi le code complexe et aide-moi à le simplifier

> Vérifie s'il y a des dépendances circulaires avant que je merge

## Plugin Claude Code

Installe le serveur MCP et les Skills ensemble :

```bash
claude plugin marketplace add ludo-technologies/pyscn
claude plugin install pyscn-mcp@pyscn-marketplace
```

## Skills vs. MCP

Les Agent Skills apprennent à un agent quand recourir à pyscn et quelle commande CLI exécuter ; elles ne nécessitent aucun serveur et fonctionnent avec tout agent prenant en charge le format Skill. Le [serveur MCP](mcp.md) expose au contraire les mêmes analyses comme des appels d'outils structurés — utilisez-le si vous voulez des résultats JSON typés directement intégrés au client plutôt qu'une sortie shell. Les deux sont complémentaires et peuvent être installés ensemble via le plugin Claude Code ci-dessus.

## Voir aussi

- [Intégration MCP](mcp.md)
- [Référence CLI](../cli/index.md)
- [add-skills](https://github.com/ludo-technologies/add-skills)
