# concrete-instantiation-dependency

**Catégorie** : Injection de dépendances  
**Sévérité** : Avertissement  
**Déclenché par** : `pyscn analyze`, `pyscn check --select di`

## Ce que fait cette règle

Signale une classe qui construit un collaborateur concret à l'intérieur de `__init__` (par exemple `self.repo = SqlUserRepository()`) au lieu de le recevoir comme paramètre.

## Pourquoi est-ce un problème ?

Lorsqu'une classe construit ses propres collaborateurs, elle s'approprie également leur configuration, leur durée de vie et leurs dépendances transitives. `OrderService()` exige soudainement une chaîne de connexion à la base de données, un client HTTP et des identifiants — tous obtenus implicitement — parce que `SqlUserRepository()` et `StripeGateway()` en ont besoin.

Les tests le ressentent en premier. Il n'existe aucune jointure pour intercaler un faux : chaque test doit soit lancer le vrai collaborateur, soit monkey-patcher la classe en cours de construction. Les tests d'intégration et les tests unitaires se confondent, et la suite de tests ralentit.

Passer le collaborateur rend la dépendance explicite, garde la classe centrée sur sa propre logique et permet à différents sites d'appel de câbler différentes implémentations — un vrai dépôt en production, un dépôt en mémoire dans les tests.

## Exemple

```python
class OrderService:
    def __init__(self):
        self.repo = SqlOrderRepository(DATABASE_URL)
        self.payments = StripeGateway(api_key=STRIPE_KEY)

    def place(self, order):
        self.repo.save(order)
        self.payments.charge(order)
```

## À utiliser à la place

Recevez les collaborateurs via `__init__` et construisez-les une seule fois à la racine de composition.

```python
class OrderService:
    def __init__(self, repo, payments):
        self.repo = repo
        self.payments = payments

    def place(self, order):
        self.repo.save(order)
        self.payments.charge(order)
```

## Options

| Option | Défaut | Description |
| --- | --- | --- |
| [`di.enabled`](../configuration/reference.md#di) | `false` | Doit être `true` pour qu'`analyze` exécute les règles DI. |
| [`di.min_severity`](../configuration/reference.md#di) | `"warning"` | Augmentez à `"error"` pour supprimer cette règle. |

## Références

- Détection des dépendances concrètes (`internal/analyzer/concrete_dependency_detector.go`).
- [Catalogue des règles](index.md) · [Dépendance à une annotation de type concrète](concrete-type-hint-dependency.md) · [Patron de localisateur de services](service-locator-pattern.md)
