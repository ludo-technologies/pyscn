# low-package-cohesion

**Catégorie** : Structure des modules  
**Sévérité** : Warning  
**Déclenchée par** : `pyscn analyze`, `pyscn check --select deps`

## Ce qu'elle fait

Signale un paquet dont le score de cohésion interne tombe en dessous de `architecture.min_cohesion` (par défaut `0.5`). La cohésion est mesurée comme le rapport entre les imports intra-paquet réels et le nombre possible entre ses sous-modules — un paquet dont les modules ne s'importent jamais entre eux obtient `0`.

## Pourquoi est-ce un problème ?

Un paquet est censé être un concept unique qui se trouve réparti entre plusieurs fichiers. Quand les fichiers ne se référencent pas entre eux, le paquet n'est qu'un dossier de code sans rapport partageant un espace de noms :

- **Chemins d'import trompeurs.** `from myapp.utils import X` suggère une relation entre `X` et tout le reste de `utils` ; une faible cohésion signifie que cette promesse est vide.
- **Aucun propriétaire naturel.** Personne n'est responsable d'« `utils` » dans son ensemble, parce qu'il n'y a pas d'ensemble.
- **Croissance sans limite.** Les paquets fourre-tout accumulent des utilitaires sans rapport jusqu'à devenir une décharge.
- **Cache une abstraction manquante.** Souvent, la bonne action n'est pas « continuer à ajouter », mais trouver le vrai concept que deux des sous-modules partagent et l'extraire.

## Exemple

```
myapp/utils/
    __init__.py
    string_utils.py     # slugify, truncate
    math_utils.py       # clamp, lerp
    io_utils.py         # atomic_write, read_json
```

Aucun de ces trois modules n'importe les autres. Le paquet `utils` a une cohésion nulle.

## À utiliser à la place

Découpez le paquet en paquets ciblés nommés d'après ce qu'ils font réellement :

```
myapp/text/          # slugify, truncate, et les utilitaires qu'ils partagent
myapp/geometry/      # clamp, lerp
myapp/fs/            # atomic_write, read_json
```

Ou — si le contenu est réellement constitué d'utilitaires ponctuels sans rapport — reconnaissez-le et arrêtez de prétendre le contraire. Nommez le paquet `misc` ou déplacez chaque utilitaire vers le module qui l'utilise réellement, et excluez le fourre-tout des contrôles de cohésion.

## Options

| Option | Défaut | Description |
| --- | --- | --- |
| [`architecture.validate_cohesion`](../configuration/reference.md#architecture) | `true` | Mettre à `false` pour désactiver cette règle. |
| [`architecture.min_cohesion`](../configuration/reference.md#architecture) | `0.5` | Les paquets en dessous de ce score sont signalés. |
| [`architecture.enabled`](../configuration/reference.md#architecture) | `true` | Interrupteur principal de l'analyse d'architecture. |
| [`architecture.fail_on_violations`](../configuration/reference.md#architecture) | `false` | Code de sortie non nul en cas de violation. |

## Références

- Calcul de la cohésion de paquet (`internal/analyzer/coupling_metrics.go`, `internal/analyzer/module_analyzer.go`).
- [Catalogue des règles](index.md) · [layer-violation](layer-violation.md)
