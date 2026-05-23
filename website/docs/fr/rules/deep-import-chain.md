# deep-import-chain

**Catégorie** : Structure des modules  
**Sévérité** : Info  
**Déclenchée par** : `pyscn analyze`, `pyscn check --select deps`

## Ce qu'elle fait

Rapporte la plus longue chaîne d'imports acyclique du projet lorsque sa profondeur dépasse la profondeur attendue pour un projet de cette taille. pyscn utilise `log₂(module_count) + 1` comme référence — un projet de 64 modules ne devrait pas avoir de chaînes plus longues que 7.

Une chaîne est un chemin dans le graphe de dépendances des modules : `a → b → c → …`, où chaque flèche représente un `import`.

## Pourquoi est-ce un problème ?

Des chaînes profondes indiquent un mauvais découpage en couches. Chaque maillon supplémentaire est un module qui doit être chargé, analysé et initialisé avant que le bas de la chaîne ne devienne utilisable, et chaque maillon est un endroit où un changement sans rapport peut se propager vers le bas.

Symptômes d'une chaîne trop profonde :

- **Démarrage lent.** L'import du module feuille déclenche une cascade d'effets de bord au niveau supérieur.
- **Tests fragiles.** Un test unitaire pour la feuille embarque la chaîne entière et casse dès qu'un élément en amont change.
- **Couplage caché.** Les modules au milieu de la chaîne n'existent souvent que comme des relais, masquant la véritable dépendance.
- **Difficile à appréhender.** Il n'y a pas un seul « niveau » auquel le code vit.

## Exemple

```
myapp.cli
  → myapp.commands
    → myapp.services
      → myapp.orchestrator
        → myapp.workers
          → myapp.adapters
            → myapp.drivers
```

Sept niveaux pour atteindre le driver. En pratique, la couche CLI n'a pas besoin de savoir que les workers existent, et les workers n'ont pas besoin de connaître la CLI — mais une modification de `drivers` peut forcer à retester toutes les couches au-dessus.

## À utiliser à la place

Introduisez une façade à la frontière pour que les couches supérieures parlent à un seul module, pas à une chaîne :

```
myapp.cli
  → myapp.commands
    → myapp.services        # point d'entrée unique
        (câble en interne orchestrator / workers / adapters / drivers)
```

Ou aplatissez : si `services`, `orchestrator` et `workers` font tous de la coordination, fusionnez-les en une seule couche et laissez-la dépendre directement de `adapters`.

## Options

| Option | Défaut | Description |
| --- | --- | --- |
| [`dependencies.find_long_chains`](../configuration/reference.md#dependencies) | `true` | Mettre à `false` pour désactiver cette règle. |
| [`dependencies.enabled`](../configuration/reference.md#dependencies) | `false` | Opt-in pour `pyscn check` ; toujours actif pour `pyscn analyze`. |

Il n'existe pas de seuil de profondeur explicite — pyscn compare la chaîne la plus longue à `log₂(module_count) + 1` et signale lorsqu'il est dépassé.

## Références

- Recherche du plus long chemin dans le DAG des modules (`internal/analyzer/module_analyzer.go`, `internal/analyzer/coupling_metrics.go`).
- [Catalogue des règles](index.md) · [circular-import](circular-import.md)
