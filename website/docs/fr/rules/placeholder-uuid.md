# placeholder-uuid

**Catégorie** : Données fictives  
**Sévérité** : Warning  
**Déclenchée par** : `pyscn check --select mockdata`

## Ce qu'elle fait

Signale les littéraux de chaînes en forme de UUID avec une très faible entropie : le UUID nul (`00000000-0000-0000-0000-000000000000`), tout à un, tout en `f`, ou de longues séries de caractères répétés qui parsent néanmoins comme un UUID.

## Pourquoi est-ce un problème ?

Le UUID nul est une valeur légitime dans quelques contextes, mais dans la plupart du code applicatif c'est un reliquat d'un stub `DEFAULT_USER_ID = "00..."` qui devait être remplacé. Comme il parse et se sérialise comme n'importe quel autre UUID, les lookups de clés étrangères, les lignes de logs et les pistes d'audit l'acceptent tous — et des lignes commencent à fusionner sur le même « utilisateur » sans qu'aucune erreur ne soit levée.

## Exemple

```python
DEFAULT_USER_ID = "00000000-0000-0000-0000-000000000000"
```

## À utiliser à la place

Générez un vrai UUID au point d'utilisation, ou exigez de l'appelant qu'il en fournisse un. Ne transportez pas une valeur sentinelle qui ressemble à de la donnée.

```python
import uuid

def new_user_id() -> uuid.UUID:
    return uuid.uuid4()
```

## Options

| Option | Défaut | Description |
| --- | --- | --- |
| [`mock_data.enabled`](../configuration/reference.md#mock_data) | `false` | Opt-in. |
| [`mock_data.min_severity`](../configuration/reference.md#mock_data) | `"info"` | Passer à `"warning"` pour ne conserver que les résultats de ce niveau. |
| [`mock_data.ignore_tests`](../configuration/reference.md#mock_data) | `true` | Ignore les fichiers de test. |
| [`mock_data.ignore_patterns`](../configuration/reference.md#mock_data) | `[]` | Motifs regex confrontés aux chemins de fichiers. |

## Références

- Implémentation : `internal/analyzer/mock_data_detector.go`.
- [Catalogue des règles](index.md) · [placeholder-phone-number](placeholder-phone-number.md) · [repetitive-string-literal](repetitive-string-literal.md)
