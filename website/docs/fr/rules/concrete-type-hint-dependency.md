# concrete-type-hint-dependency

**Catégorie** : Injection de dépendances  
**Sévérité** : Info  
**Déclenché par** : `pyscn analyze`, `pyscn check --select di`

## Ce que fait cette règle

Signale un paramètre `__init__` dont l'annotation de type est une classe concrète plutôt qu'un `Protocol`, une classe de base abstraite ou une interface.

## Pourquoi est-ce un problème ?

Une annotation de type concrète indique au lecteur — et au vérificateur de types — que cette classe n'acceptera qu'une implémentation spécifique. Même si l'exécution accepte volontiers des substituts en duck typing, le contrat déclaré dit l'inverse, et les outils qui respectent l'annotation (mypy, l'autocomplétion de l'IDE, les mocks de tests construits à partir du type) refuseront les alternatives.

En pratique, cela rend les tests plus difficiles à écrire : un test qui souhaite substituer un faux en mémoire doit soit hériter de la classe concrète (en récupérant tout son comportement), soit supprimer l'erreur de type. Cela lie aussi le consommateur à tous les imports dont la classe concrète a besoin, donc un petit utilitaire finit par entraîner toute la pile base de données.

Dépendre d'un `Protocol` ou d'une interface abstraite documente ce que la classe utilise réellement — une méthode ou deux — et laisse de la place pour les faux, les adaptateurs et les futures implémentations.

## Exemple

```python
class SqlUserRepository:
    def find(self, user_id): ...

class UserService:
    def __init__(self, repo: SqlUserRepository):
        self.repo = repo
```

## À utiliser à la place

Déclarez un `Protocol` décrivant les méthodes sur lesquelles vous vous appuyez et dépendez-en.

```python
from typing import Protocol

class UserRepository(Protocol):
    def find(self, user_id: str) -> User: ...

class UserService:
    def __init__(self, repo: UserRepository):
        self.repo = repo
```

## Options

| Option | Défaut | Description |
| --- | --- | --- |
| [`di.enabled`](../configuration/reference.md#di) | `false` | Doit être `true` pour qu'`analyze` exécute les règles DI. |
| [`di.min_severity`](../configuration/reference.md#di) | `"warning"` | Cette règle se signale au niveau `info` ; abaissez `min_severity` à `"info"` pour la voir. |

## Références

- Détection des dépendances concrètes (`internal/analyzer/concrete_dependency_detector.go`).
- [Catalogue des règles](index.md) · [Dépendance par instanciation concrète](concrete-instantiation-dependency.md) · [Trop de paramètres de constructeur](too-many-constructor-parameters.md)
