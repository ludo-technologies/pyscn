# Format du fichier de configuration

pyscn lit la configuration au format **TOML**. Vous pouvez conserver vos réglages dans un fichier `.pyscn.toml` dédié ou dans une section `[tool.pyscn]` de votre `pyproject.toml` existant.

## Découverte des fichiers

Lorsque vous exécutez `pyscn analyze` ou `pyscn check`, pyscn remonte depuis le chemin cible à la recherche de :

1. `.pyscn.toml` (priorité la plus haute)
2. `pyproject.toml` contenant une section `[tool.pyscn]`

Le premier fichier trouvé est utilisé. Les répertoires parents sont parcourus jusqu'à trouver une correspondance ou atteindre la racine du système de fichiers. Si aucun fichier n'est trouvé, les valeurs par défaut intégrées sont utilisées.

Vous pouvez également passer un chemin explicite :

```bash
pyscn analyze --config ./configs/strict.toml src/
```

Cela contourne la découverte.

## Ordre de priorité

Lorsqu'un réglage apparaît à plusieurs endroits, le dernier l'emporte :

1. **Valeurs par défaut intégrées** (priorité la plus basse)
2. **`pyproject.toml` → `[tool.pyscn]`**
3. **`.pyscn.toml`**
4. **Options en ligne de commande** (priorité la plus haute)

Les options en ligne de commande ne sont prises en compte que si elles ont été **explicitement définies** — les valeurs par défaut inchangées ne surchargent pas les valeurs de configuration.

## Deux styles de fichier

=== ".pyscn.toml"

    ```toml
    [complexity]
    max_complexity = 15

    [dead_code]
    min_severity = "critical"
    ```

=== "pyproject.toml"

    ```toml
    [tool.pyscn.complexity]
    max_complexity = 15

    [tool.pyscn.dead_code]
    min_severity = "critical"
    ```

Si les deux fichiers existent dans le même répertoire, `.pyscn.toml` l'emporte.

## Générer un fichier de démarrage

```bash
pyscn init
```

Cela écrit un fichier `.pyscn.toml` entièrement commenté avec chaque option, sa valeur par défaut et une courte description. Modifiez les valeurs qui vous intéressent et supprimez (ou laissez tel quel) le reste.

```bash
pyscn init --force   # écraser l'existant
pyscn init --config tools/pyscn.toml   # chemin personnalisé
```

## Validation

pyscn valide la configuration au chargement et se termine avec le code `2` en cas de problème. Règles de validation courantes :

- Les seuils de complexité doivent satisfaire `low ≥ 1` et `medium > low`.
- Le format de sortie doit être l'un de `text`, `json`, `yaml`, `csv`, `html`.
- La sévérité du code mort doit être `info`, `warning` ou `critical`.
- Les seuils de similarité des clones doivent être dans `[0.0, 1.0]`.
- Au moins un motif d'inclusion doit être spécifié.

## Variables d'environnement

pyscn ne lit **pas** la configuration depuis des variables d'environnement. Le serveur MCP fait une exception : `PYSCN_CONFIG` peut pointer vers un fichier de configuration.

## Étapes suivantes

- [Référence](reference.md) — chaque clé, documentée.
- [Exemples](examples.md) — CI stricte, gros codebase, surcharges minimales.
