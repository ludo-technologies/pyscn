# mock-email-address

**Catégorie** : Données fictives  
**Sévérité** : Warning  
**Déclenchée par** : `pyscn check --select mockdata`

## Ce qu'elle fait

Signale les adresses e-mail dont la partie domaine est réservée aux tests : `test@example.com`, `admin@test.com`, `foo@localhost`, et similaires.

## Pourquoi est-ce un problème ?

Un e-mail avec un domaine de test dans du code applicatif est presque toujours un reliquat d'une fixture, d'un tutoriel, ou d'un stub « à remplir plus tard ». Contrairement à une adresse mal formée, il passe la validation et la sérialisation sans encombre, et finit donc silencieusement dans des lignes de base de données, des files de notifications et des en-têtes « from ». Le chemin habituel de découverte est un ticket de support demandant pourquoi personne n'a reçu de lien de réinitialisation.

Cette règle complète [mock-domain-in-string](mock-domain-in-string.md) mais traite spécifiquement la forme e-mail afin de garder la liste de domaines restreinte et la correspondance précise.

## Exemple

```python
admin_email = "admin@example.com"
```

## À utiliser à la place

Lisez l'adresse depuis la configuration, ou acceptez-la comme paramètre.

```python
admin_email = settings.admin_email
```

## Options

| Option | Défaut | Description |
| --- | --- | --- |
| [`mock_data.enabled`](../configuration/reference.md#mock_data) | `false` | Opt-in. |
| [`mock_data.domains`](../configuration/reference.md#mock_data) | *(liste RFC 2606)* | Domaines considérés comme placeholder ; partagés avec `mock-domain-in-string`. |
| [`mock_data.min_severity`](../configuration/reference.md#mock_data) | `"info"` | Passer à `"warning"` pour ne conserver que les résultats de ce niveau. |
| [`mock_data.ignore_tests`](../configuration/reference.md#mock_data) | `true` | Ignore les fichiers de test. |
| [`mock_data.ignore_patterns`](../configuration/reference.md#mock_data) | `[]` | Motifs regex confrontés aux chemins de fichiers. |

## Références

- RFC 2606 : *Reserved Top Level DNS Names.*
- Implémentation : `internal/analyzer/mock_data_detector.go`.
- [Catalogue des règles](index.md) · [mock-domain-in-string](mock-domain-in-string.md) · [test-credential-in-code](test-credential-in-code.md)
