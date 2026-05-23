# duplicate-code-modified

**Catégorie** : Code dupliqué  
**Sévérité** : Info  
**Déclenchée par** : `pyscn analyze`, `pyscn check --select clones`

## Ce que fait cette règle

Signale les blocs de code qui partagent la majeure partie de leur structure mais comportent des instructions ajoutées, supprimées ou modifiées (clones de Type 3, similarité ≥ 0,70). Désactivée par défaut ; activez-la en ajoutant `"type3"` à `clones.enabled_clone_types`.

## Pourquoi est-ce un problème ?

Les quasi-doublons sont le type de clones le plus courant dans les bases de code réelles, et souvent le plus difficile à nettoyer. Deux fonctions font presque la même chose, mais l'une a une étape de validation supplémentaire ou un chemin d'erreur légèrement différent. La partie commune devrait vivre à un seul endroit ; les différences devraient être la seule chose qui varie entre les sites d'appel.

Laissés en l'état, les clones modifiés divergent : un bug est corrigé dans une copie mais pas dans l'autre, une nouvelle fonctionnalité est ajoutée à l'une mais oubliée dans la suivante. Les relecteurs cessent de faire confiance au fait qu'un code d'apparence similaire se comporte de la même manière.

Cette règle est en Info parce que le refactoring approprié est moins mécanique que pour les clones de Type 1 ou de Type 2 — parfois les différences sont précisément l'objectif, et fusionner nuirait à la clarté. Considérez les constats comme des candidats à examiner, pas comme des défauts automatiques.

## Exemple

```python
def export_users_csv(users, path):
    with open(path, "w") as f:
        writer = csv.writer(f)
        writer.writerow(["id", "name", "email"])
        for u in users:
            writer.writerow([u.id, u.name, u.email])

def export_orders_csv(orders, path):
    with open(path, "w") as f:
        writer = csv.writer(f)
        writer.writerow(["id", "total", "status"])
        for o in orders:
            if o.status != "draft":
                writer.writerow([o.id, o.total, o.status])
```

## À utiliser à la place

Extrayez la structure commune et paramétrez l'en-tête, la forme des lignes et tout filtre éventuel.

```python
def export_csv(rows, path, header, to_row, keep=lambda _: True):
    with open(path, "w") as f:
        writer = csv.writer(f)
        writer.writerow(header)
        for row in rows:
            if keep(row):
                writer.writerow(to_row(row))

def export_users_csv(users, path):
    export_csv(users, path, ["id", "name", "email"],
               lambda u: [u.id, u.name, u.email])

def export_orders_csv(orders, path):
    export_csv(orders, path, ["id", "total", "status"],
               lambda o: [o.id, o.total, o.status],
               keep=lambda o: o.status != "draft")
```

## Options

| Option | Valeur par défaut | Description |
| --- | --- | --- |
| [`clones.type3_threshold`](../configuration/reference.md#clones) | `0.70` | Similarité minimale pour qu'une paire soit signalée comme modifiée. |
| [`clones.enabled_clone_types`](../configuration/reference.md#clones) | `["type1","type2","type4"]` | Ajoutez `"type3"` pour activer cette règle. |
| [`clones.similarity_threshold`](../configuration/reference.md#clones) | `0.65` | Plancher global appliqué avant les seuils par type. |
| [`clones.min_lines`](../configuration/reference.md#clones) | `5` | Taille minimale du fragment en lignes. |
| [`clones.min_nodes`](../configuration/reference.md#clones) | `10` | Taille minimale du fragment en nœuds AST. |

## Références

- Implémentation de la détection de clones (`internal/analyzer/clone_detector.go`, `internal/analyzer/apted.go`).
- [Catalogue des règles](index.md) · [Clones identiques](duplicate-code-identical.md) · [Clones renommés](duplicate-code-renamed.md) · [Clones sémantiques](duplicate-code-semantic.md)
