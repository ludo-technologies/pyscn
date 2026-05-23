# single-responsibility

**Catégorie** : Structure des modules  
**Sévérité** : Warning (Error lorsque le module est aussi un hub avec un fan-in/out très élevé)  
**Déclenchée par** : `pyscn analyze`, `pyscn check --select deps`

## Ce qu'elle fait

Signale un module qui mélange plus de `architecture.max_responsibilities` (par défaut `3`) préoccupations de dépendance distinctes, ou qui agit comme un hub fan-in/fan-out pour plus de préoccupations que la moyenne du reste du projet.

Une « préoccupation » est déduite des noms des voisins du module : pour chaque module que celui-ci importe ou qui l'importe, l'analyseur prend le premier segment du chemin du voisin qui ne fait pas partie du chemin du module courant et qui n'est pas un terme générique fourre-tout (`base`, `common`, `helpers`, `node`, `shared`, `util`, `utils`). Ces segments sont dédupliqués ; le compte donne le nombre de responsabilités que pyscn attribue au module.

Un module est signalé lorsqu'au moins une des conditions suivantes est remplie :

- Il possède plus de `max_responsibilities` préoccupations distinctes.
- Son fan-in (nombre d'importateurs) et son fan-out (nombre d'imports) sont tous deux au-dessus de la `moyenne + écart-type` du projet, et il possède plus d'une préoccupation.

## Pourquoi est-ce un problème ?

Le principe de responsabilité unique (SRP) concerne les *axes de changement*. Un module qui participe à plusieurs clusters de dépendances sans rapport a plusieurs raisons de changer :

- **Les modifications se propagent.** Toucher une préoccupation force à relire et retester les autres, parce qu'elles partagent toutes la même frontière de module.
- **Les imports mentent.** `from myapp.core import X` ne dit rien au lecteur — `core` fait plusieurs choses.
- **Les hubs deviennent des goulots d'étranglement.** Un module que tout le monde importe *et* qui importe tout est un point unique de contention pour les modifications, les revues et les fusions.
- **Cache une couture manquante.** Quand deux préoccupations finissent constamment dans le même fichier, la bonne correction est en général un nouveau module qui nomme la relation entre elles.

## Exemple

```
myapp/core.py
```

```python
# myapp/core.py
from myapp.routers import user_router, order_router
from myapp.services import billing_service, notification_service
from myapp.repositories import user_repo, order_repo
from myapp.telemetry import metrics, tracing

# ...code de glue qui rassemble tout...
```

`core` mélange quatre préoccupations (`routers`, `services`, `repositories`, `telemetry`), et il est importé à la fois par des routers et des services ailleurs — fan-in et fan-out sont tous deux élevés. pyscn le signale comme surchargé.

## À utiliser à la place

Découpez le module selon les préoccupations qui existent déjà. Chaque nouveau module doit nommer *un seul* axe de changement.

```
myapp/wiring/web.py          # câblage au niveau des routers
myapp/wiring/services.py     # câblage au niveau des services
myapp/wiring/persistence.py  # câblage des repositories
myapp/wiring/observability.py
```

Ou, si le module est légitimement une racine de composition, restreignez sa portée : il ne devrait que *câbler* les parties entre elles, sans en plus implémenter des règles métier, définir des types ou posséder la télémétrie.

## Options

| Option | Défaut | Description |
| --- | --- | --- |
| [`architecture.validate_responsibility`](../configuration/reference.md#architecture) | `true` | Mettre à `false` pour désactiver cette règle. |
| [`architecture.max_responsibilities`](../configuration/reference.md#architecture) | `3` | Les modules possédant plus de préoccupations que cela sont signalés. |
| [`architecture.enabled`](../configuration/reference.md#architecture) | `true` | Interrupteur principal de l'analyse d'architecture. |
| [`architecture.fail_on_violations`](../configuration/reference.md#architecture) | `false` | Code de sortie non nul en cas de violation. |

## Références

- Inférence des responsabilités et règles de sévérité : `service/responsibility_analysis.go`.
- Martin, R. C. *Agile Software Development: Principles, Patterns, and Practices*, 2002 (Chapitre 8 — SRP).
- [Catalogue des règles](index.md) · [low-package-cohesion](low-package-cohesion.md) · [layer-violation](layer-violation.md)
