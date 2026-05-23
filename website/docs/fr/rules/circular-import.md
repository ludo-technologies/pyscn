# circular-import

**Catégorie** : Structure des modules  
**Sévérité** : Configurable selon la taille du cycle (Low / Medium / High / Critical)  
**Déclenchée par** : `pyscn analyze`, `pyscn check --select circular`

## Ce qu'elle fait

Signale les groupes de modules qui forment un cycle d'imports — le module A importe B (directement ou transitivement) et B importe A. Les cycles sont détectés en exécutant l'algorithme des composantes fortement connexes de Tarjan sur le graphe de dépendances des modules.

La sévérité est attribuée en fonction de la taille du cycle et du fan-in de ses membres :

| Membres du cycle | Sévérité |
| --- | --- |
| 2 | Low |
| 3 – 5 | Medium |
| 6 – 9 | High |
| 10+, ou tout membre avec un fan-in > 10 | Critical |

## Pourquoi est-ce un problème ?

Un import circulaire signifie que deux modules ou plus ne peuvent pas être compris, testés ou livrés indépendamment. Concrètement :

- **Erreurs à l'import.** Python initialise partiellement les modules pendant les imports circulaires ; l'accès à un attribut sur le module à moitié chargé lève `ImportError` ou `AttributeError` selon l'ordre des instructions.
- **Couplage fort.** Les membres du cycle partagent un unique « module logique » réparti sur plusieurs fichiers. Une modification dans l'un tend à imposer une modification dans tous les autres.
- **Refactorisation bloquée.** Vous ne pouvez pas déplacer, renommer ou supprimer un membre du cycle sans toucher aux autres.
- **Aggravation avec la croissance du cycle.** Un cycle à 2 modules est une nuisance ; un cycle à 10 modules est un échec architectural — d'où l'échelle de sévérité.

## Exemple

```python
# myapp/orders.py
from myapp.billing import Invoice

class Order:
    def invoice(self) -> Invoice:
        return Invoice(self)
```

```python
# myapp/billing.py
from myapp.orders import Order

class Invoice:
    def __init__(self, order: "Order"):
        self.order = order
```

`orders` importe `billing` pour le type de retour ; `billing` importe `orders` pour le paramètre du constructeur. Exécuter l'un ou l'autre fichier au niveau supérieur déclenche le cycle.

## À utiliser à la place

Extrayez les types partagés dans un troisième module afin que les deux en dépendent au lieu de dépendre l'un de l'autre :

```python
# myapp/domain.py
class Order: ...
class Invoice: ...
```

```python
# myapp/orders.py
from myapp.domain import Order, Invoice
```

```python
# myapp/billing.py
from myapp.domain import Order, Invoice
```

Si l'arête inverse n'est nécessaire que pour les annotations de type, protégez-la avec `TYPE_CHECKING` afin qu'elle ne soit pas évaluée à l'exécution :

```python
# myapp/billing.py
from __future__ import annotations
from typing import TYPE_CHECKING

if TYPE_CHECKING:
    from myapp.orders import Order

class Invoice:
    def __init__(self, order: "Order"): ...
```

## Options

| Option | Défaut | Description |
| --- | --- | --- |
| [`dependencies.detect_cycles`](../configuration/reference.md#dependencies) | `true` | Mettre à `false` pour désactiver cette règle. |
| [`dependencies.cycle_reporting`](../configuration/reference.md#dependencies) | `"summary"` | `all`, `critical` ou `summary` — contrôle le nombre de cycles affichés dans le rapport. |
| [`dependencies.max_cycles_to_show`](../configuration/reference.md#dependencies) | `10` | Plafond des cycles rapportés. |
| `--max-cycles N` (check) | `0` | Fait échouer la commande `check` lorsque le nombre de cycles dépasse `N`. |
| `--allow-circular-deps` (check) | off | Rétrograde les cycles en avertissements plutôt qu'en échecs. |

## Références

- Implémentation Tarjan SCC (`internal/analyzer/circular_detector.go`), construction du graphe de modules (`internal/analyzer/module_analyzer.go`).
- [Catalogue des règles](index.md) · [deep-import-chain](deep-import-chain.md)
