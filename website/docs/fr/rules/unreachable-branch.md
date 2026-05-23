# unreachable-branch

**Catégorie** : Code inatteignable  
**Sévérité** : Avertissement  
**Déclenchée par** : `pyscn analyze`, `pyscn check`

## Ce que fait cette règle

Signale une branche `if`, `elif` ou `else` qui ne peut pas être empruntée parce que toutes les branches précédentes se terminent par `return`, `raise`, `break` ou `continue`.

## Pourquoi est-ce un problème ?

Lorsque chaque branche antérieure quitte déjà la fonction ou la boucle, la branche restante est logiquement morte. La garde semble pourtant pertinente pour le lecteur, ce qui masque le véritable flot de contrôle.

Cela indique généralement :

- **Des conditions redondantes** — le `else` n'existe que parce que le `if` avait l'habitude de retomber.
- **Un bug subtil** — l'auteur s'attendait à ce que la branche ultérieure s'exécute dans certains cas, mais les sorties précédentes la rendent impossible.
- **Du code défensif obsolète** — un repli qui ne peut plus être atteint.

Les tests ne peuvent pas couvrir la branche, et les relecteurs perdent du temps à raisonner sur un chemin qui ne s'exécute jamais.

## Exemple

```python
def classify(payment):
    if payment.amount < 0:
        raise ValueError("negative amount")
    elif payment.amount == 0:
        return "empty"
    else:
        return "normal"
    return "unknown"   # ← branche inatteignable
```

## À utiliser à la place

Supprimez la branche morte, ou restructurez les branches précédentes pour que le repli soit réellement atteignable.

```python
def classify(payment):
    if payment.amount < 0:
        raise ValueError("negative amount")
    if payment.amount == 0:
        return "empty"
    return "normal"
```

## Options

| Option | Valeur par défaut | Description |
| --- | --- | --- |
| [`dead_code.detect_unreachable_branches`](../configuration/reference.md#dead_code) | `true` | Mettez à `false` pour désactiver cette règle. |
| [`dead_code.min_severity`](../configuration/reference.md#dead_code) | `"warning"` | Passez à `"critical"` pour masquer ces constats ; abaissez à `"info"` pour en remonter davantage. |
| [`dead_code.ignore_patterns`](../configuration/reference.md#dead_code) | `[]` | Expressions régulières comparées à la ligne source ; les correspondances sont supprimées. |

## Références

- Analyse d'accessibilité sur le graphe de flot de contrôle (`internal/analyzer/dead_code.go`).
- [Catalogue des règles](index.md) · [Code inatteignable après return](unreachable-after-return.md) · [Code inatteignable après raise](unreachable-after-raise.md)
