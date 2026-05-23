# service-locator-pattern

**Catégorie** : Injection de dépendances  
**Sévérité** : Avertissement  
**Déclenché par** : `pyscn analyze`, `pyscn check --select di`

## Ce que fait cette règle

Signale une classe qui reçoit un localisateur, un registre ou un conteneur, et qui en extrait des dépendances nommées au moment de l'appel de méthode, par exemple `self.locator.get("payment_service")`.

## Pourquoi est-ce un problème ?

Un localisateur de services échange une dépendance claire contre de nombreuses dépendances cachées. La signature du constructeur suggère que la classe a besoin d'un seul objet, mais le vrai contrat est « tout ce que cette classe se trouve à rechercher à l'exécution ». Un lecteur doit parcourir toutes les méthodes à coups de grep pour découvrir qu'`OrderService` a en réalité besoin d'une passerelle de paiement, d'un notificateur et d'une horloge.

Les tests doivent fournir un faux localisateur qui connaît chaque clé que la classe peut demander, et une clé manquante apparaît généralement sous forme d'`AttributeError` ou de `KeyError` au fond d'une méthode plutôt que comme un échec de construction clair. Renommer un service exige de retrouver chaque recherche par chaîne ; l'analyse statique et les outils de refactoring d'IDE ne peuvent pas aider.

Passer directement les services réels via `__init__` donne à la classe un contrat vérifiable et élimine l'indirection par chaînes de caractères.

## Exemple

```python
class OrderService:
    def __init__(self, locator):
        self.locator = locator

    def place(self, order):
        self.locator.get("order_repo").save(order)
        self.locator.get("payment_service").charge(order)
        self.locator.get("notifier").send(order.user, "placed")
```

## À utiliser à la place

Recevez chaque service directement, afin que les dépendances soient visibles et vérifiables par typage.

```python
class OrderService:
    def __init__(self, repo, payments, notifier):
        self.repo = repo
        self.payments = payments
        self.notifier = notifier

    def place(self, order):
        self.repo.save(order)
        self.payments.charge(order)
        self.notifier.send(order.user, "placed")
```

## Options

| Option | Défaut | Description |
| --- | --- | --- |
| [`di.enabled`](../configuration/reference.md#di) | `false` | Doit être `true` pour qu'`analyze` exécute les règles DI. |
| [`di.min_severity`](../configuration/reference.md#di) | `"warning"` | Augmentez à `"error"` pour supprimer cette règle. |

## Références

- Détection de localisateur de services (`internal/analyzer/service_locator_detector.go`).
- [Catalogue des règles](index.md) · [Dépendance par instanciation concrète](concrete-instantiation-dependency.md) · [Trop de paramètres de constructeur](too-many-constructor-parameters.md)
