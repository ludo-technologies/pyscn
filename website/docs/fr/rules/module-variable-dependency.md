# module-variable-dependency

**Catégorie** : Injection de dépendances  
**Sévérité** : Avertissement  
**Déclenché par** : `pyscn analyze`, `pyscn check --select di`

## Ce que fait cette règle

Signale une classe qui lit ou écrit directement une variable mutable au niveau du module sans instruction `global` — un couplage implicite à l'état du module.

## Pourquoi est-ce un problème ?

Contrairement à une affectation `global`, la lecture d'un nom au niveau du module est silencieuse : la classe semble autonome, mais son comportement dépend en réalité de ce qui se trouve dans cette variable de module au moment de l'appel. Un test qui instancie la classe de manière isolée peut quand même réussir ou échouer en fonction d'un import sans rapport.

Cela casse également la substituabilité. Vous ne pouvez pas donner à la classe un collaborateur différent sans monkey-patcher le module, ce qui est fragile et dépend de l'ordre. Deux instances de la classe sont forcées de partager le même objet sous-jacent, que vous le vouliez ou non.

Faire du collaborateur un paramètre du constructeur documente la dépendance, permet une configuration par instance et autorise les tests à injecter un faux sans toucher aux variables globales du module.

## Exemple

```python
config = load_config()

class UserRepository:
    def find(self, user_id):
        conn = connect(config.database_url)
        return conn.query("SELECT * FROM users WHERE id = ?", user_id)
```

## À utiliser à la place

Acceptez le collaborateur comme paramètre du constructeur.

```python
class UserRepository:
    def __init__(self, database_url):
        self.database_url = database_url

    def find(self, user_id):
        conn = connect(self.database_url)
        return conn.query("SELECT * FROM users WHERE id = ?", user_id)
```

## Options

| Option | Défaut | Description |
| --- | --- | --- |
| [`di.enabled`](../configuration/reference.md#di) | `false` | Doit être `true` pour qu'`analyze` exécute les règles DI. |
| [`di.min_severity`](../configuration/reference.md#di) | `"warning"` | Augmentez à `"error"` pour supprimer cette règle. |

## Références

- Détection des dépendances cachées (`internal/analyzer/hidden_dependency_detector.go`).
- [Catalogue des règles](index.md) · [Dépendance à l'état global](global-state-dependency.md) · [Dépendance au patron singleton](singleton-pattern-dependency.md)
