# high-class-coupling

**Catégorie** : Conception de classes  
**Sévérité** : Configurable par seuil  
**Déclenché par** : `pyscn analyze`, `pyscn check`

## Ce que fait cette règle

Signale les classes qui dépendent de trop d'autres classes (la métrique Coupling Between Objects, ou CBO, de Chidamber & Kemerer). pyscn compte le nombre de classes distinctes qu'une classe référence via l'héritage, les annotations de type, l'instanciation directe, l'accès aux attributs sur des modules importés et les imports.

En clair : *trop de choses doivent être en place pour que cette classe fonctionne.*

## Pourquoi est-ce un problème ?

Une classe fortement couplée est difficile à vivre :

- **Difficile à tester** — la construire dans un test unitaire entraîne tout un réseau de collaborateurs, donc les tests deviennent des tests d'intégration ou reposent sur un mocking massif.
- **Difficile à modifier** — un changement de signature dans l'une des dépendances se répercute sur cette classe.
- **Difficile à réutiliser** — vous ne pouvez pas l'extraire vers un autre projet sans extraire tout son voisinage.
- **Signe d'une abstraction manquante** — la classe orchestre probablement des choses qui devraient se trouver derrière une interface plus petite.

## Exemple

```python
from billing.stripe_gateway import StripeGateway
from billing.paypal_gateway import PayPalGateway
from notifications.sendgrid import SendGridClient
from notifications.twilio import TwilioClient
from storage.s3 import S3Bucket
from storage.postgres import PostgresConnection
from audit.datadog import DatadogLogger
from auth.okta import OktaClient

class OrderService:
    def __init__(self):
        self.stripe = StripeGateway()
        self.paypal = PayPalGateway()
        self.email = SendGridClient()
        self.sms = TwilioClient()
        self.blobs = S3Bucket("orders")
        self.db = PostgresConnection()
        self.audit = DatadogLogger()
        self.auth = OktaClient()

    def place(self, user, cart): ...
```

`OrderService` est couplé à 8 classes concrètes de fournisseurs. Remplacer Stripe par Adyen, ou exécuter un test sans Postgres opérationnel, implique de modifier `OrderService`.

## À utiliser à la place

Dépendez de petits protocoles et injectez les collaborateurs via `__init__`. Le service ne sait plus quel fournisseur se trouve à l'autre extrémité.

```python
from typing import Protocol

class PaymentGateway(Protocol):
    def charge(self, amount: int, token: str) -> str: ...

class Notifier(Protocol):
    def notify(self, user_id: str, message: str) -> None: ...

class OrderRepository(Protocol):
    def save(self, order) -> None: ...

class OrderService:
    def __init__(
        self,
        payments: PaymentGateway,
        notifier: Notifier,
        repo: OrderRepository,
    ):
        self._payments = payments
        self._notifier = notifier
        self._repo = repo

    def place(self, user, cart): ...
```

Si une classe a réellement besoin de nombreux collaborateurs, scindez-la par responsabilité (par exemple `Checkout`, `Fulfillment`, `Receipt`) et laissez un orchestrateur les appeler.

## Options

| Option | Défaut | Description |
| --- | --- | --- |
| [`cbo.low_threshold`](../configuration/reference.md#cbo) | `3` | À ce seuil ou en dessous, la classe est signalée comme à faible risque. |
| [`cbo.medium_threshold`](../configuration/reference.md#cbo) | `7` | Au-dessus de ce seuil, la classe est à risque élevé. |
| [`cbo.min_cbo`](../configuration/reference.md#cbo) | `0` | Les classes dont le couplage est inférieur à cette valeur sont omises du rapport. |
| [`cbo.include_builtins`](../configuration/reference.md#cbo) | `false` | Compter les types intégrés (`list`, `dict`, `Exception`, …) comme dépendances. |
| [`cbo.include_imports`](../configuration/reference.md#cbo) | `true` | Compter les classes atteintes uniquement via des instructions `import`. |

## Références

- Chidamber, S. R. & Kemerer, C. F. *A Metrics Suite for Object Oriented Design.* IEEE TSE, 1994.
- Implémentation : `internal/analyzer/cbo.go`.
- [Catalogue des règles](index.md) · [low-class-cohesion](low-class-cohesion.md)
