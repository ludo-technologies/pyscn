# layer-violation

**Catégorie** : Structure des modules  
**Sévérité** : Configurable via `architecture.rules[].severity`  
**Déclenchée par** : `pyscn analyze`, `pyscn check --select deps`

## Ce qu'elle fait

Signale une instruction `import` lorsque la couche du module source n'est pas autorisée à dépendre de la couche du module cible, selon les `[[architecture.rules]]` que vous avez configurées. Les couches sont assignées aux modules en faisant correspondre les fragments de noms de paquets définis dans `[[architecture.layers]]`.

## Pourquoi est-ce un problème ?

Une architecture en couches n'apporte de bénéfice que tant que les couches tiennent. Un seul raccourci de `presentation` vers `infrastructure` suffit à :

- **Anéantir la testabilité.** La couche de présentation ne peut désormais être exercée qu'avec une vraie base de données / un vrai client HTTP derrière.
- **Créer un couplage caché.** Remplacer l'implémentation de l'infrastructure casse silencieusement du code d'UI qui n'était pas censé en connaître l'existence.
- **Normaliser la violation.** Dès qu'un raccourci existe, le suivant est plus facile à justifier.

La règle est l'application automatisée du schéma d'architecture que vous avez déjà dessiné dans un document de conception.

## Exemple

Configuration :

```toml
[[architecture.layers]]
name = "presentation"
packages = ["api", "handlers"]

[[architecture.layers]]
name = "application"
packages = ["services", "usecases"]

[[architecture.layers]]
name = "infrastructure"
packages = ["repositories", "db"]

[[architecture.rules]]
from = "presentation"
allow = ["application"]
deny = ["infrastructure"]
```

Code en infraction :

```python
# myapp/api/orders.py  (presentation)
from myapp.repositories.orders import OrderRepository   # ← interdit

def list_orders():
    return OrderRepository().all()
```

`presentation` saute par-dessus `application` pour aller directement dans `infrastructure`.

## À utiliser à la place

Faites passer l'appel par la couche application :

```python
# myapp/services/orders.py  (application)
from myapp.repositories.orders import OrderRepository

def list_orders():
    return OrderRepository().all()
```

```python
# myapp/api/orders.py  (presentation)
from myapp.services.orders import list_orders

def get():
    return list_orders()
```

`api` ne dépend désormais que de `services`, et l'infrastructure est remplaçable sans toucher à la couche de présentation.

## Options

| Option | Défaut | Description |
| --- | --- | --- |
| [`[[architecture.layers]]`](../configuration/reference.md#architecture) | — | Définit les couches et les fragments de paquets qui appartiennent à chacune. |
| [`[[architecture.rules]]`](../configuration/reference.md#architecture) | — | `from` / `allow` / `deny` / `severity` optionnel par règle. |
| [`architecture.validate_layers`](../configuration/reference.md#architecture) | `true` | Mettre à `false` pour désactiver cette règle. |
| [`architecture.strict_mode`](../configuration/reference.md#architecture) | `true` | En mode strict, tout ce qui n'est pas explicitement autorisé est refusé. |
| [`architecture.fail_on_violations`](../configuration/reference.md#architecture) | `false` | Code de sortie non nul lorsqu'une violation est détectée. |

Sans couches configurées, l'analyseur fonctionne en mode permissif et cette règle ne produit aucun résultat.

## Références

- Résolution des couches et évaluation des règles (`internal/analyzer/module_analyzer.go`).
- [Catalogue des règles](index.md) · [low-package-cohesion](low-package-cohesion.md)
