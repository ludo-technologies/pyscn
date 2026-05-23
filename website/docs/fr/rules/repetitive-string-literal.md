# repetitive-string-literal

**Catégorie** : Données fictives  
**Sévérité** : Info  
**Déclenchée par** : `pyscn check --select mockdata`

## Ce qu'elle fait

Signale les littéraux de chaînes de longueur 4 à 20 avec des motifs de caractères très répétitifs : `aaaa`, `1111`, `xxxxxxxx`, et autres séries similaires à un ou deux caractères.

## Pourquoi est-ce un problème ?

Une chaîne comme `"aaaaaaaaaaaaaaaa"` n'est presque jamais une vraie valeur — c'est la forme que prend quelque chose tapé pour passer un validateur pendant que le développeur câblait autre chose. Laissé en production, elle devient une clé d'API, une entrée de hash, ou un token qui ressemble à de la donnée et passe les vérifications de longueur sans avoir aucun sens.

La règle est bornée en longueur (4–20 caractères) afin d'éviter de signaler des remplissages intentionnels comme des constantes de padding ou des vecteurs de test qui nécessitent légitimement des caractères répétés.

## Exemple

```python
api_key = "aaaaaaaaaaaaaaaa"
```

## À utiliser à la place

Lisez les secrets et les tokens depuis la configuration ou un coffre-fort de secrets. Ne mettez pas de valeurs placeholder en dur dans le code source.

```python
import os

api_key = os.environ["SERVICE_API_KEY"]
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
- [Catalogue des règles](index.md) · [test-credential-in-code](test-credential-in-code.md) · [placeholder-uuid](placeholder-uuid.md)
