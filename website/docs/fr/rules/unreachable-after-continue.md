# unreachable-after-continue

**Catégorie** : Code inatteignable  
**Sévérité** : Critique  
**Déclenchée par** : `pyscn analyze`, `pyscn check`

## Ce que fait cette règle

Signale les instructions qui apparaissent après une instruction `continue` à l'intérieur d'une boucle.

## Pourquoi est-ce un problème ?

`continue` saute directement à l'itération suivante. Toute instruction qui le suit dans le même bloc est ignorée à chaque itération, donc elle s'exécute zéro fois.

Causes typiques :

- **Logique réordonnée** — une garde a été convertie en `continue` et le travail qui suivait a été laissé en place.
- **Un effet de bord mal placé** — des mises à jour de compteur ou des logs qui auraient dû s'exécuter avant le saut.
- **Sémantique mal comprise** — l'auteur s'attendait à ce que `continue` se comporte comme `pass`.

Comme l'instruction est inatteignable, les tests ne peuvent pas la couvrir et le comportement voulu ne se produit jamais en silence.

## Exemple

```python
for order in orders:
    if order.status == "cancelled":
        continue
        metrics.record_skip(order.id)   # ← ne s'exécute jamais
    process(order)
```

## À utiliser à la place

Exécutez l'instruction avant `continue`, ou supprimez-la.

```python
for order in orders:
    if order.status == "cancelled":
        metrics.record_skip(order.id)
        continue
    process(order)
```

## Options

| Option | Valeur par défaut | Description |
| --- | --- | --- |
| [`dead_code.detect_after_continue`](../configuration/reference.md#dead_code) | `true` | Mettez à `false` pour désactiver cette règle. |
| [`dead_code.min_severity`](../configuration/reference.md#dead_code) | `"warning"` | `"critical"` ne conserve que ce type de constat ; `"info"` en remonte davantage. |
| [`dead_code.ignore_patterns`](../configuration/reference.md#dead_code) | `[]` | Expressions régulières comparées à la ligne source ; les correspondances sont supprimées. |

## Références

- Analyse d'accessibilité sur le graphe de flot de contrôle (`internal/analyzer/dead_code.go`).
- [Catalogue des règles](index.md) · [Code inatteignable après break](unreachable-after-break.md) · [Code inatteignable après return](unreachable-after-return.md)
