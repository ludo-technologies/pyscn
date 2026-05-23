# placeholder-comment

**Catégorie** : Données fictives  
**Sévérité** : Info  
**Déclenchée par** : `pyscn check --select mockdata`

## Ce qu'elle fait

Signale les commentaires contenant des marqueurs de travail inachevé : `TODO`, `FIXME`, `XXX`, `HACK`, `BUG`, `NOTE`.

## Pourquoi est-ce un problème ?

Un `# TODO` dans le code source est une promesse que l'auteur s'est faite à son futur lui-même, sans échéance ni relecteur. La plupart des codebases en accumulent plus vite qu'elles ne les résolvent. Chacun est un petit morceau de périmètre caché — le lecteur doit décider s'il est encore pertinent, s'il bloque une modification, et si quelqu'un en assure réellement le suivi.

Cette règle ne prétend pas que chaque marqueur est un bug. Elle expose la liste pour que vous puissiez décider, projet par projet, de les résoudre, de les convertir en tickets suivis, ou de les accepter avec une politique explicite.

## Exemple

```python
def process_order(order):
    # TODO: handle refunds
    ...
```

## À utiliser à la place

Soit vous implémentez le travail, soit vous convertissez le marqueur en lien vers un ticket suivi pour que l'intention vive quelque part avec un état de clôture.

```python
def process_order(order):
    # Les remboursements sont gérés par le service de facturation : voir ticket #1423.
    ...
```

## Options

| Option | Défaut | Description |
| --- | --- | --- |
| [`mock_data.enabled`](../configuration/reference.md#mock_data) | `false` | Opt-in. |
| [`mock_data.min_severity`](../configuration/reference.md#mock_data) | `"info"` | Passer à `"warning"` pour exclure cette règle. |
| [`mock_data.ignore_tests`](../configuration/reference.md#mock_data) | `true` | Ignore les fichiers de test. |
| [`mock_data.ignore_patterns`](../configuration/reference.md#mock_data) | `[]` | Motifs regex confrontés aux chemins de fichiers. |

## Références

- Implémentation : `internal/analyzer/mock_data_detector.go`.
- [Catalogue des règles](index.md) · [mock-keyword-in-code](mock-keyword-in-code.md)
