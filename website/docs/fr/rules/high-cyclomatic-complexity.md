# high-cyclomatic-complexity

**Catégorie** : Complexité  
**Sévérité** : Configurable selon un seuil  
**Déclenchée par** : `pyscn analyze`, `pyscn check`

## Ce que fait cette règle

Signale les fonctions dont la complexité cyclomatique de McCabe dépasse le seuil configuré. Chaque `if`, `elif`, `for`, `while`, `except`, `match case` et clause booléenne à l'intérieur d'une compréhension ajoute un au compteur. Une fonction linéaire commence à 1.

pyscn ne compte pas les opérateurs de court-circuit `and` / `or` comme des branches séparées.

## Pourquoi est-ce un problème ?

Un nombre élevé de branches signifie :

- **Plus de chemins à lire** — chaque relecteur doit simuler mentalement chaque branche pour comprendre ce que fait la fonction.
- **Plus de chemins à tester** — une couverture de branches complète exige un test par chemin ; en pratique, la plupart des fonctions très ramifiées sont sous-testées.
- **Une densité de défauts plus élevée** — des études empiriques depuis McCabe (1976) corrèlent la complexité avec le taux de bugs.
- **Plus difficile à modifier en toute sécurité** — une petite édition dans une branche peut silencieusement casser une autre.

Les fonctions au-dessus d'environ 10 assument généralement plusieurs tâches qui pourraient être nommées et séparées.

## Exemple

```python
def price_for(user, cart, coupon, region):
    total = 0
    for item in cart:
        if item.category == "book":
            if region == "EU":
                total += item.price * 0.95
            elif region == "US":
                total += item.price
            else:
                total += item.price * 1.10
        elif item.category == "food":
            if user.is_student:
                total += item.price * 0.90
            else:
                total += item.price
        else:
            total += item.price
    if coupon:
        if coupon.kind == "percent":
            total *= 1 - coupon.value
        elif coupon.kind == "fixed":
            total -= coupon.value
    if total < 0:
        total = 0
    return total
```

Complexité cyclomatique : 13.

## À utiliser à la place

Extrayez la tarification par article et la gestion des coupons, et remplacez la condition imbriquée par une table de dispatch.

```python
REGION_BOOK_MULTIPLIER = {"EU": 0.95, "US": 1.00}

def _book_price(item, region):
    return item.price * REGION_BOOK_MULTIPLIER.get(region, 1.10)

def _food_price(item, user):
    return item.price * (0.90 if user.is_student else 1.00)

PRICERS = {"book": _book_price, "food": _food_price}

def _item_price(item, user, region):
    pricer = PRICERS.get(item.category)
    return pricer(item, region) if item.category == "book" else \
           pricer(item, user)   if item.category == "food" else \
           item.price

def _apply_coupon(total, coupon):
    if coupon is None:
        return total
    if coupon.kind == "percent":
        return total * (1 - coupon.value)
    return total - coupon.value

def price_for(user, cart, coupon, region):
    subtotal = sum(_item_price(i, user, region) for i in cart)
    return max(0, _apply_coupon(subtotal, coupon))
```

Chaque fonction utilitaire a désormais une complexité de 1 à 3 et une responsabilité unique. Les gardes (`if coupon is None: return`) aplatissent les branches restantes.

## Options

| Option | Valeur par défaut | Description |
| --- | --- | --- |
| [`complexity.max_complexity`](../configuration/reference.md#complexity) | `0` | Limite stricte appliquée par `pyscn check`. `0` signifie qu'aucune application n'est faite dans `analyze` ; `pyscn check --max-complexity` utilise `10` si non défini. |
| [`complexity.low_threshold`](../configuration/reference.md#complexity) | `9` | Les fonctions à cette valeur ou en dessous sont signalées comme à faible risque. |
| [`complexity.medium_threshold`](../configuration/reference.md#complexity) | `19` | Au-dessus de cette valeur, une fonction est à risque élevé. |
| [`complexity.min_complexity`](../configuration/reference.md#complexity) | `1` | Les fonctions en dessous de cette valeur sont omises du rapport. |

## Métriques associées

pyscn calcule deux autres mesures liées à la complexité à côté du compteur de McCabe. Elles apparaissent dans le rapport HTML et la sortie JSON, mais ne sont *pas* appliquées par `complexity.max_complexity` :

- **Complexité cognitive** (style SonarQube). Pénalise plus lourdement l'imbrication et les ruptures du flot linéaire que McCabe — utile pour repérer les fonctions qui *paraissent* difficiles à suivre même lorsque leur complexité cyclomatique est modérée. Apparaît sous le nom `CognitiveComplexity` par fonction dans la sortie JSON.
- **Métriques de code brutes** par fichier (SLOC, LLOC, lignes de commentaires, lignes vides, densité de commentaires). Voir [`raw_metrics`](../output/schemas.md) dans la référence des schémas.

## Références

- McCabe, T. J. *A Complexity Measure.* IEEE Transactions on Software Engineering, 1976.
- Construction du graphe de flot de contrôle et comptage cyclomatique : `internal/analyzer/complexity.go`, `internal/analyzer/complexity_analyzer.go`, `internal/analyzer/cfg_builder.go`.
- [Catalogue des règles](index.md) · [too-many-constructor-parameters](too-many-constructor-parameters.md)
