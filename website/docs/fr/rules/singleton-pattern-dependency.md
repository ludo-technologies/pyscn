# singleton-pattern-dependency

**Catégorie** : Injection de dépendances  
**Sévérité** : Avertissement  
**Déclenché par** : `pyscn analyze`, `pyscn check --select di`

## Ce que fait cette règle

Signale une classe qui implémente le patron singleton en se mettant elle-même en cache sur un attribut `_instance` au niveau de la classe.

## Pourquoi est-ce un problème ?

Un singleton est un état global déguisé en classe. Chaque appelant qui écrit `PaymentGateway.instance()` dépend de l'unique objet que la classe choisit de retourner, et cet objet survit entre les tests à moins que chaque test ne pense à le réinitialiser. Un test distrait suffit pour que le suivant hérite d'un état périmé.

Comme le singleton décide lui-même de sa durée de vie, les appelants ne peuvent pas lui fournir différents collaborateurs dans différents contextes — une seconde configuration, un faux pour les tests, une instance par locataire. La substitution exige d'accéder à la classe et de réinitialiser `_instance`, ce qui est exactement le couplage que le singleton était censé cacher.

Le patron occulte également les vraies dépendances : lire le code d'une méthode qui appelle `X.instance()` ne vous dit rien sur ce dont `X` a besoin ou sur l'endroit où il a été configuré.

## Exemple

```python
class PaymentGateway:
    _instance = None

    @classmethod
    def instance(cls):
        if cls._instance is None:
            cls._instance = cls()
        return cls._instance

    def charge(self, order):
        ...
```

## À utiliser à la place

Construisez l'objet une seule fois en bordure de l'application et passez-le à ce qui en a besoin.

```python
class PaymentGateway:
    def charge(self, order):
        ...

# wiring, done once at startup
gateway = PaymentGateway()
order_service = OrderService(gateway)
```

## Options

| Option | Défaut | Description |
| --- | --- | --- |
| [`di.enabled`](../configuration/reference.md#di) | `false` | Doit être `true` pour qu'`analyze` exécute les règles DI. |
| [`di.min_severity`](../configuration/reference.md#di) | `"warning"` | Augmentez à `"error"` pour supprimer cette règle. |

## Références

- Détection des dépendances cachées (`internal/analyzer/hidden_dependency_detector.go`).
- [Catalogue des règles](index.md) · [Dépendance à l'état global](global-state-dependency.md) · [Dépendance à une variable de module](module-variable-dependency.md)
