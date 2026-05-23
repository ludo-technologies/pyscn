# test-credential-in-code

**Catégorie** : Données fictives  
**Sévérité** : Warning  
**Déclenchée par** : `pyscn check --select mockdata`

## Ce qu'elle fait

Signale les littéraux de chaînes qui ressemblent à des credentials de test évidents : `password123`, `secret123`, `testpassword`, `token0`, `api_key_test`, et autres motifs construits à partir de mots évoquant des identifiants plus des suffixes triviaux.

## Pourquoi est-ce un problème ?

Ce ne sont pas de vrais secrets — c'est précisément le problème. Ce sont les valeurs tapées en mettant en place un client, en écrivant un premier test, ou en remplissant un champ obligatoire. Une fois commités, ils présentent deux modes de défaillance :

- **Embarras / correction** : le littéral est utilisé comme valeur par défaut et est livré aux utilisateurs, si bien que le « mot de passe administrateur par défaut » est vraiment `password123`.
- **Rotation oubliée** : un vrai secret était censé remplacer le placeholder avant la sortie ; personne n'a remarqué que ce n'était pas fait.

pyscn n'est pas un scanner de sécurité et ne tente pas de détecter les secrets à forte entropie. Cette règle attrape le cas inverse : des credentials à faible entropie, manifestement factices, qui n'auraient jamais dû figurer dans le code source.

## Exemple

```python
DEFAULT_PASSWORD = "password123"
```

## À utiliser à la place

Lisez les credentials depuis l'environnement ou un gestionnaire de secrets. Si une valeur par défaut est réellement nécessaire pour le développement local, conservez-la dans un fichier de configuration séparé qui n'est pas chargé en production.

```python
import os

DEFAULT_PASSWORD = os.environ["APP_DEFAULT_PASSWORD"]
```

## Options

| Option | Défaut | Description |
| --- | --- | --- |
| [`mock_data.enabled`](../configuration/reference.md#mock_data) | `false` | Opt-in. |
| [`mock_data.min_severity`](../configuration/reference.md#mock_data) | `"info"` | Passer à `"warning"` pour ne conserver que les résultats de ce niveau. |
| [`mock_data.ignore_tests`](../configuration/reference.md#mock_data) | `true` | Ignore les fichiers de test — les credentials de test ont leur place dans les tests. |
| [`mock_data.ignore_patterns`](../configuration/reference.md#mock_data) | `[]` | Motifs regex confrontés aux chemins de fichiers. |

## Références

- Implémentation : `internal/analyzer/mock_data_detector.go`.
- [Catalogue des règles](index.md) · [repetitive-string-literal](repetitive-string-literal.md) · [mock-keyword-in-code](mock-keyword-in-code.md)
