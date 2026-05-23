# mock-domain-in-string

**Catégorie** : Données fictives  
**Sévérité** : Warning  
**Déclenchée par** : `pyscn check --select mockdata`

## Ce qu'elle fait

Signale les littéraux de chaînes contenant des domaines réservés à la documentation et aux tests : `example.com`, `example.org`, `example.net`, `test.com`, `localhost`, `invalid`, `foo.com`, `bar.com`, et noms similaires définis par les RFC 2606 / RFC 6761.

## Pourquoi est-ce un problème ?

Ces domaines existent précisément pour que les exemples et les tests n'entrent pas en collision avec du vrai trafic. C'est utile pendant la rédaction de la documentation — et problématique une fois qu'un littéral est livré. Une URL `example.com` codée en dur en production est en général soit :

- Un placeholder qui devait être remplacé avant la sortie, soit
- Une valeur de configuration qui n'aurait jamais dû être codée en dur en premier lieu.

Dans les deux cas, le mode de défaillance est silencieux : les requêtes réussissent (le domaine résout vers une page de documentation ou rien du tout), aucune exception n'est levée, et le bug n'est remarqué que lorsque quelqu'un demande pourquoi les inscriptions n'arrivent pas.

## Exemple

```python
SIGNUP_URL = "https://example.com/signup"
```

## À utiliser à la place

Déplacez la valeur dans la configuration, ou mettez la vraie URL en clair.

```python
SIGNUP_URL = settings.signup_url
```

## Options

| Option | Défaut | Description |
| --- | --- | --- |
| [`mock_data.enabled`](../configuration/reference.md#mock_data) | `false` | Opt-in. |
| [`mock_data.domains`](../configuration/reference.md#mock_data) | *(liste RFC 2606)* | Remplace ou étend la liste de domaines réservés. |
| [`mock_data.min_severity`](../configuration/reference.md#mock_data) | `"info"` | Passer à `"warning"` pour ne conserver que les résultats de ce niveau. |
| [`mock_data.ignore_tests`](../configuration/reference.md#mock_data) | `true` | Ignore les fichiers de test. |
| [`mock_data.ignore_patterns`](../configuration/reference.md#mock_data) | `[]` | Motifs regex confrontés aux chemins de fichiers. |

## Références

- RFC 2606 : *Reserved Top Level DNS Names.*
- RFC 6761 : *Special-Use Domain Names.*
- Implémentation : `internal/analyzer/mock_data_detector.go`.
- [Catalogue des règles](index.md) · [mock-email-address](mock-email-address.md) · [mock-keyword-in-code](mock-keyword-in-code.md)
