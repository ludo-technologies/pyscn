# global-state-dependency

**Catégorie** : Injection de dépendances  
**Sévérité** : Erreur  
**Déclenché par** : `pyscn analyze`, `pyscn check --select di`

## Ce que fait cette règle

Signale une méthode de classe qui utilise une instruction `global` pour lire ou modifier un état au niveau du module.

## Pourquoi est-ce un problème ?

Un `global` à l'intérieur d'une méthode lie la classe à une variable de module spécifique qui n'est pas visible depuis l'interface de la classe. Rien dans `OrderService(...)` n'indique à un lecteur que la construire ne suffit pas — une valeur au niveau du module doit aussi être amorcée, sinon la méthode se comportera de manière inattendue.

Les tests en souffrent le plus. Pour exercer une méthode qui touche un état global, chaque test doit accéder au module, sauvegarder l'ancienne valeur, en installer une nouvelle et la restaurer au teardown — et tout test qui oublie de nettoyer laisse fuiter l'état vers le suivant. Exécuter les tests en parallèle devient dangereux.

La dépendance est réelle ; elle est simplement cachée. En faire un paramètre explicite du constructeur supprime la surprise.

## Exemple

```python
_current_user = None

class AuditLog:
    def record(self, action):
        global _current_user
        entry = {"user": _current_user, "action": action}
        db.insert("audit", entry)
```

## À utiliser à la place

Passez la valeur via `__init__` pour que la dépendance soit visible et substituable.

```python
class AuditLog:
    def __init__(self, user):
        self.user = user

    def record(self, action):
        db.insert("audit", {"user": self.user, "action": action})
```

## Options

| Option | Défaut | Description |
| --- | --- | --- |
| [`di.enabled`](../configuration/reference.md#di) | `false` | Doit être `true` pour qu'`analyze` exécute les règles DI. |
| [`di.min_severity`](../configuration/reference.md#di) | `"warning"` | Cette règle se signale au niveau `error` ; elle est affichée sauf si `min_severity` est élevée au-dessus de `error`. |

## Références

- Détection des dépendances cachées (`internal/analyzer/hidden_dependency_detector.go`).
- [Catalogue des règles](index.md) · [Dépendance à une variable de module](module-variable-dependency.md) · [Dépendance au patron singleton](singleton-pattern-dependency.md)
