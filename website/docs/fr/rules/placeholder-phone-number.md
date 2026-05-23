# placeholder-phone-number

**Catégorie** : Données fictives  
**Sévérité** : Warning  
**Déclenchée par** : `pyscn check --select mockdata`

## Ce qu'elle fait

Signale les numéros de téléphone dans les littéraux de chaînes qui suivent des motifs manifestement factices : que des zéros (`000-0000-0000`), des chiffres séquentiels (`123-456-7890`, `012-345-6789`), ou de longues séries de chiffres répétés.

## Pourquoi est-ce un problème ?

Un numéro de téléphone placeholder est le genre de valeur qui survit depuis la première version d'un formulaire et n'est jamais revisitée. Il valide, il se formate et il transite correctement par la base de données — donc rien ne casse jusqu'à ce qu'un véritable utilisateur le voie sur un écran de confirmation ou qu'un agent de support tente de l'appeler.

## Exemple

```python
default_phone = "000-0000-0000"
```

## À utiliser à la place

Laissez le champ vide, exigez-le de l'appelant, ou tirez-le de la configuration. Un numéro de téléphone inconnu doit être absent, pas falsifié.

```python
default_phone: str | None = None
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
- [Catalogue des règles](index.md) · [placeholder-uuid](placeholder-uuid.md) · [repetitive-string-literal](repetitive-string-literal.md)
