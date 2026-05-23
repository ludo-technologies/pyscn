# duplicate-code-semantic

**Catégorie** : Code dupliqué  
**Sévérité** : Avertissement  
**Déclenchée par** : `pyscn analyze`, `pyscn check --select clones`

## Ce que fait cette règle

Signale les blocs de code syntaxiquement différents qui calculent le même résultat (clones de Type 4, similarité ≥ 0,65). Utilise l'analyse de flot de données pour comparer le comportement plutôt que la structure.

## Pourquoi est-ce un problème ?

Les clones sémantiques sont les duplications que vous ne remarquez pas pendant la revue. Une fonction utilise une boucle, l'autre une compréhension ; l'une construit un dictionnaire via `update`, l'autre via la syntaxe de fusion. Le code a l'air différent et passe donc l'inspection visuelle, mais les deux implémentations font le même travail.

Les risques sont les mêmes que pour toute duplication — les modifications doivent être faites à plusieurs endroits — mais il y a un coût supplémentaire. Les lecteurs ne peuvent pas dire d'un coup d'œil si les deux implémentations s'accordent sur les cas limites. La version en boucle saute-t-elle aussi les `None` ? La compréhension lève-t-elle une exception sur une entrée vide ? L'audit mental doit être refait à chaque fois.

Réduire à une implémentation unique supprime l'audit et la divergence.

## Exemple

```python
def unique_emails(users):
    seen = set()
    result = []
    for u in users:
        if u.email not in seen:
            seen.add(u.email)
            result.append(u.email)
    return result

def distinct_emails(users):
    return list({u.email: None for u in users}.keys())
```

## À utiliser à la place

Choisissez une implémentation et utilisez-la partout. Préférez la version la plus claire ; si les deux ont du mérite, conservez l'une et documentez pourquoi.

```python
def unique_emails(users):
    """Return user emails in first-seen order, without duplicates."""
    return list(dict.fromkeys(u.email for u in users))
```

## Options

| Option | Valeur par défaut | Description |
| --- | --- | --- |
| [`clones.type4_threshold`](../configuration/reference.md#clones) | `0.65` | Similarité minimale pour qu'une paire soit signalée comme sémantique. |
| [`clones.enable_dfa`](../configuration/reference.md#clones) | `true` | Active l'analyse de flot de données qui alimente la détection de Type 4. |
| [`clones.similarity_threshold`](../configuration/reference.md#clones) | `0.65` | Plancher global appliqué avant les seuils par type. |
| [`clones.enabled_clone_types`](../configuration/reference.md#clones) | `["type1","type2","type4"]` | Incluez `"type4"` pour garder cette règle active. |
| [`clones.min_lines`](../configuration/reference.md#clones) | `5` | Taille minimale du fragment en lignes. |

## Références

- Implémentation de la détection de clones (`internal/analyzer/clone_detector.go`, `internal/analyzer/apted.go`).
- [Catalogue des règles](index.md) · [Clones identiques](duplicate-code-identical.md) · [Clones renommés](duplicate-code-renamed.md) · [Clones modifiés](duplicate-code-modified.md)
