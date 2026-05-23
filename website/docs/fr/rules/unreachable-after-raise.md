# unreachable-after-raise

**Catégorie** : Code inatteignable  
**Sévérité** : Critique  
**Déclenchée par** : `pyscn analyze`, `pyscn check`

## Ce que fait cette règle

Signale les instructions qui apparaissent après une instruction `raise` dans le même bloc de code.

## Pourquoi est-ce un problème ?

Un `raise` déroule la pile de manière inconditionnelle. Toute instruction qui le suit dans le même bloc n'est jamais exécutée.

Cela indique généralement :

- **Un nettoyage obsolète** — du code qui était censé s'exécuter avant que l'exception soit levée.
- **Un artefact de refactoring** — le `raise` a remplacé une branche antérieure et les lignes voisines ont été laissées en place.
- **Un bug de logique** — l'auteur supposait que l'exécution se poursuivrait au-delà du `raise`.

Le code mort après `raise` ne sera jamais exercé par les tests et ne se manifestera jamais en production, donc tout bug qu'il dissimule reste silencieux.

## Exemple

```python
def withdraw(account, amount):
    if amount > account.balance:
        raise InsufficientFunds(account.id)
        account.balance -= amount   # ← ne s'exécute jamais
    account.balance -= amount
```

## À utiliser à la place

Déplacez l'instruction au-dessus du `raise`, ou supprimez-la.

```python
def withdraw(account, amount):
    if amount > account.balance:
        raise InsufficientFunds(account.id)
    account.balance -= amount
```

## Options

| Option | Valeur par défaut | Description |
| --- | --- | --- |
| [`dead_code.detect_after_raise`](../configuration/reference.md#dead_code) | `true` | Mettez à `false` pour désactiver cette règle. |
| [`dead_code.min_severity`](../configuration/reference.md#dead_code) | `"warning"` | `"critical"` ne conserve que ce type de constat ; `"info"` en remonte davantage. |
| [`dead_code.ignore_patterns`](../configuration/reference.md#dead_code) | `[]` | Expressions régulières comparées à la ligne source ; les correspondances sont supprimées. |

## Références

- Analyse d'accessibilité sur le graphe de flot de contrôle (`internal/analyzer/dead_code.go`).
- [Catalogue des règles](index.md) · [Code inatteignable après return](unreachable-after-return.md) · [Branche inatteignable](unreachable-branch.md)
