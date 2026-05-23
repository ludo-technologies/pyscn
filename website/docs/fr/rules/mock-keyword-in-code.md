# mock-keyword-in-code

**Catégorie** : Données fictives  
**Sévérité** : Info (dans les chaînes) / Warning (dans les identifiants)  
**Déclenchée par** : `pyscn check --select mockdata`

## Ce qu'elle fait

Signale les identifiants et les littéraux de chaînes contenant des mots-clés de placeholder courants : `mock`, `fake`, `dummy`, `test`, `sample`, `example`, `placeholder`, `stub`, `fixture`, `temp`, `foo`, `bar`, `baz`, `lorem`, `ipsum`.

## Pourquoi est-ce un problème ?

Ce sont les mots que vous tapez quand vous êtes encore en train de réfléchir à quelque chose. Ils sont parfaits dans un notebook, un fichier de test ou une expérimentation rapide — mais quand l'un d'eux survit jusqu'en production, cela signifie en général qu'un stub n'a jamais été remplacé. Le nom `foo` dans un module commité n'est presque jamais ce que l'auteur voulait livrer.

Les correspondances dans les **identifiants** sont traitées comme des warnings parce qu'un nom lié (`foo = get_user()`) change le comportement. Les correspondances dans les **littéraux de chaînes** sont au niveau info car un `"fake_user"` oublié est plus souvent cosmétique que cassé — mais il vaut quand même la peine d'être revu avant une mise en production.

## Exemple

```python
def create_user():
    name = "fake_user"    # correspondance sur littéral de chaîne
    foo = get_user()      # correspondance sur l'identifiant `foo`
    return foo
```

## À utiliser à la place

Supprimez le placeholder. Utilisez de vraies données, lisez depuis la configuration, ou déplacez le stub dans une fixture de test où il a sa place.

```python
def create_user(name: str):
    user = get_user(name)
    return user
```

## Options

| Option | Défaut | Description |
| --- | --- | --- |
| [`mock_data.enabled`](../configuration/reference.md#mock_data) | `false` | Opt-in ; toute la catégorie est désactivée par défaut. |
| [`mock_data.keywords`](../configuration/reference.md#mock_data) | *(liste intégrée)* | Remplace la liste de mots-clés pour cette règle. |
| [`mock_data.min_severity`](../configuration/reference.md#mock_data) | `"info"` | `"warning"` ne conserve que les correspondances dans les identifiants. |
| [`mock_data.ignore_tests`](../configuration/reference.md#mock_data) | `true` | Ignore les fichiers qui ressemblent à des tests. |
| [`mock_data.ignore_patterns`](../configuration/reference.md#mock_data) | `[]` | Motifs regex confrontés aux chemins de fichiers ; les correspondances sont supprimées. |

## Références

- Implémentation : `internal/analyzer/mock_data_detector.go`.
- [Catalogue des règles](index.md) · [mock-domain-in-string](mock-domain-in-string.md) · [test-credential-in-code](test-credential-in-code.md) · [placeholder-comment](placeholder-comment.md)
