# unreachable-after-return

**Catégorie** : Code inatteignable  
**Sévérité** : Critique  
**Déclenchée par** : `pyscn analyze`, `pyscn check`

## Ce que fait cette règle

Signale les instructions qui apparaissent après une instruction `return` dans le même bloc de code.

## Pourquoi est-ce un problème ?

Le code placé après `return` ne s'exécute jamais. Il s'agit généralement de l'un des cas suivants :

- **Un reliquat de refactoring** — le `return` a été remonté et le code en dessous a été oublié.
- **Un bug** — le développeur s'attendait à ce que le code s'exécute, mais une modification du flot de contrôle l'a rendu inatteignable.
- **Un nettoyage mal placé** — quelque chose qui aurait dû s'exécuter avant le retour.

Dans tous les cas, le code est mort : il consomme du temps de lecture, il n'est pas couvert par les tests (parce qu'il ne peut pas l'être), et s'il dissimule un bug, ce bug ne sera jamais signalé par le comportement des utilisateurs.

## Exemple

```python
def charge(order):
    if order.total <= 0:
        return None
        log.debug("zero-value charge")   # ← ne s'exécute jamais
    ...
```

## À utiliser à la place

Déplacez l'instruction au-dessus du `return`, ou supprimez-la si elle n'est plus nécessaire.

```python
def charge(order):
    if order.total <= 0:
        log.debug("zero-value charge")
        return None
    ...
```

## Options

| Option | Valeur par défaut | Description |
| --- | --- | --- |
| [`dead_code.detect_after_return`](../configuration/reference.md#dead_code) | `true` | Mettez à `false` pour désactiver cette règle. |
| [`dead_code.min_severity`](../configuration/reference.md#dead_code) | `"warning"` | `"critical"` ne conserve que ce type de constat ; `"info"` en remonte davantage. |
| [`dead_code.ignore_patterns`](../configuration/reference.md#dead_code) | `[]` | Expressions régulières comparées à la ligne source ; les correspondances sont supprimées. |

## Références

- Analyse d'accessibilité sur le graphe de flot de contrôle (`internal/analyzer/dead_code.go`).
- [Catalogue des règles](index.md) · [Branche inatteignable](unreachable-branch.md) · [Code inatteignable après raise](unreachable-after-raise.md)
