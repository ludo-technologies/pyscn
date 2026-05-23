# `pyscn init`

Génère un fichier de configuration `.pyscn.toml` avec toutes les options documentées en commentaires.

```text
pyscn init [flags]
```

## Ce qu'elle fait

Écrit un fichier TOML commenté avec les sections les plus couramment ajustées :

- `[output]`, `[complexity]`, `[dead_code]`, `[clones]`, `[cbo]`, `[analysis]`, `[architecture]` (plus des exemples `[[architecture.layers]]` et `[[architecture.rules]]`)
- Valeurs par défaut renseignées
- Commentaires expliquant chaque clé

Le fichier généré n'inclut **pas** toutes les sections configurables. Les options pour la cohésion LCOM4 (`[lcom]`), l'analyse des dépendances entre modules (`[dependencies]`), la détection de données factices (`[mock_data]`) et les anti-patterns DI (`[di]`) sont valides mais doivent être ajoutées manuellement. Consultez la [Référence de configuration](../configuration/reference.md) pour toutes les clés.

Une fois le fichier en place, chaque exécution ultérieure de `pyscn analyze` / `pyscn check` dans ce projet (ou tout sous-répertoire) le détecte automatiquement.

## Options

| Option | Défaut | Description |
| --- | --- | --- |
| `-c, --config <path>` | `.pyscn.toml` | Chemin du fichier de sortie. |
| `-f, --force`         | off          | Écrase un fichier existant. |

## Codes de sortie

| Code | Signification |
| --- | --- |
| `0` | Fichier écrit avec succès. |
| `1` | Le fichier existe déjà (utilisez `--force` pour écraser) ou l'écriture a échoué. |

## Exemples

```bash
# Créer .pyscn.toml dans le répertoire courant
pyscn init

# Utiliser un nom de fichier personnalisé
pyscn init --config tools/pyscn.toml

# Écraser une configuration existante
pyscn init --force
```

## Que modifier en premier

Après avoir exécuté `init`, les réglages que la plupart des projets finissent par ajuster sont :

| Réglage | Ajustement habituel |
| --- | --- |
| `[complexity].max_complexity` | À fixer à `10`, `15` ou `20` selon le niveau de rigueur souhaité en CI. |
| `[dead_code].min_severity`     | Passer à `"critical"` si les avertissements sont trop bruyants. |
| `[clones].similarity_threshold`| Abaisser à `0.80` pour trouver plus de clones, monter à `0.90` pour réduire le bruit. |
| `[analysis].exclude_patterns`  | Ajouter les chemins de code généré, migrations, etc. |

Consultez la [Référence de configuration](../configuration/reference.md) complète.

## Voir aussi

- [Référence de configuration](../configuration/reference.md) — toutes les options expliquées.
- [Exemples de configuration](../configuration/examples.md) — CI stricte, gros codebase, surcharges minimales.
