# too-many-constructor-parameters

**Catégorie** : Injection de dépendances  
**Sévérité** : Avertissement  
**Déclenché par** : `pyscn analyze`, `pyscn check --select di`

## Ce que fait cette règle

Signale une méthode `__init__` dont le nombre de paramètres (en excluant `self`) dépasse `di.constructor_param_threshold` (par défaut `5`).

## Pourquoi est-ce un problème ?

Une signature de constructeur longue est le symptôme d'une classe qui a pris trop de responsabilités. Chaque paramètre est un collaborateur que la classe connaît, et chaque site d'appel doit tous les fournir — y compris les tests, qui finissent par construire des fixtures élaborées juste pour instancier l'objet.

Avec le temps, ajouter une dépendance de plus paraît anodin, donc la liste s'allonge. Les lecteurs perdent de vue ce dont la classe a réellement besoin par rapport à ce qui est accessoire, et réordonner ou ajouter des valeurs par défaut aux paramètres devient risqué.

Lorsque vous voyez plus de cinq dépendances, la classe accomplit généralement deux ou trois tâches qui pourraient être séparées, ou plusieurs paramètres devraient être regroupés en un seul objet.

## Exemple

```python
class OrderService:
    def __init__(
        self,
        user_repo,
        order_repo,
        payment_gateway,
        inventory,
        notifier,
        audit_log,
        clock,
    ):
        self.user_repo = user_repo
        self.order_repo = order_repo
        ...
```

## À utiliser à la place

Regroupez les collaborateurs liés en un objet de paramètres, ou scindez la classe selon ses responsabilités.

```python
@dataclass
class OrderDependencies:
    users: UserRepository
    orders: OrderRepository
    payments: PaymentGateway
    inventory: Inventory
    notifier: Notifier

class OrderService:
    def __init__(self, deps: OrderDependencies, clock: Clock):
        self.deps = deps
        self.clock = clock
```

## Options

| Option | Défaut | Description |
| --- | --- | --- |
| [`di.enabled`](../configuration/reference.md#di) | `false` | Doit être `true` pour qu'`analyze` exécute les règles DI. `check --select di` l'active implicitement. |
| [`di.min_severity`](../configuration/reference.md#di) | `"warning"` | Augmentez à `"error"` pour masquer cette règle, ou abaissez à `"info"` pour en faire ressortir davantage. |
| [`di.constructor_param_threshold`](../configuration/reference.md#di) | `5` | Nombre de paramètres au-dessus duquel `__init__` est signalé. |

## Références

- Détection de sur-injection par constructeur (`internal/analyzer/constructor_analyzer.go`).
- [Catalogue des règles](index.md) · [Dépendance par instanciation concrète](concrete-instantiation-dependency.md) · [Patron de localisateur de services](service-locator-pattern.md)
