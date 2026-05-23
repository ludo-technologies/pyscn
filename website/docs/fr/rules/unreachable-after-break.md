# unreachable-after-break

**Catégorie** : Code inatteignable  
**Sévérité** : Critique  
**Déclenchée par** : `pyscn analyze`, `pyscn check`

## Ce que fait cette règle

Signale les instructions qui apparaissent après une instruction `break` à l'intérieur d'une boucle.

## Pourquoi est-ce un problème ?

`break` quitte immédiatement la boucle englobante. Tout ce qui se trouve après lui dans le même bloc s'exécute zéro fois.

Causes fréquentes :

- **Un incrément ou une mise à jour d'accumulateur mal placés** — l'auteur souhaitait qu'ils s'exécutent à la dernière itération.
- **Du log ou du nettoyage résiduel** — déplacé sous le `break` lors d'un refactoring.
- **Une confusion sur le flot de contrôle** — l'auteur pensait que `break` ne sautait qu'une partie de l'itération.

Le code est inatteignable, donc les tests ne peuvent pas le couvrir et les bugs qu'il contient ne se manifesteront jamais.

## Exemple

```python
for user in users:
    if user.id == target_id:
        break
        user.last_seen = now()   # ← ne s'exécute jamais
```

## À utiliser à la place

Effectuez le travail avant le `break`, ou supprimez l'instruction morte.

```python
for user in users:
    if user.id == target_id:
        user.last_seen = now()
        break
```

## Options

| Option | Valeur par défaut | Description |
| --- | --- | --- |
| [`dead_code.detect_after_break`](../configuration/reference.md#dead_code) | `true` | Mettez à `false` pour désactiver cette règle. |
| [`dead_code.min_severity`](../configuration/reference.md#dead_code) | `"warning"` | `"critical"` ne conserve que ce type de constat ; `"info"` en remonte davantage. |
| [`dead_code.ignore_patterns`](../configuration/reference.md#dead_code) | `[]` | Expressions régulières comparées à la ligne source ; les correspondances sont supprimées. |

## Références

- Analyse d'accessibilité sur le graphe de flot de contrôle (`internal/analyzer/dead_code.go`).
- [Catalogue des règles](index.md) · [Code inatteignable après continue](unreachable-after-continue.md) · [Code inatteignable après return](unreachable-after-return.md)
