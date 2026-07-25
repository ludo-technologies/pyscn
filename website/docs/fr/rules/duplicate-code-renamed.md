# duplicate-code-renamed

**Catégorie** : Code dupliqué  
**Sévérité** : Avertissement  
**Déclenchée par** : `pyscn analyze`, `pyscn check --select clones`

## Ce que fait cette règle

Signale les blocs de code ayant la même structure mais des identifiants ou littéraux différents (clones de Type 2, similarité ≥ 0,75).

## Pourquoi est-ce un problème ?

Les clones renommés apparaissent quand quelqu'un copie une fonction, lance un rechercher-remplacer sur les noms de variables et passe à autre chose. La structure est identique, seuls les noms ont changé. Le coût de maintenance est le même que pour des clones textuellement identiques — chaque modification doit être appliquée à plusieurs endroits — mais la duplication est plus difficile à repérer à l'œil nu parce que les mots diffèrent.

C'est aussi le signe que le code original n'était pas paramétré alors qu'il aurait dû l'être. Ce qui varie (un type, un nom de champ, une constante) est un argument de fonction naturel.

## Exemple

```python
def total_for_orders(orders):
    total = 0
    for order in orders:
        if order.status == "paid":
            total += order.amount
    return total

def total_for_invoices(invoices):
    total = 0
    for invoice in invoices:
        if invoice.status == "settled":
            total += invoice.amount
    return total
```

## À utiliser à la place

Extrayez une fonction utilitaire générique qui prend le prédicat variable et les accesseurs de champs en paramètres.

```python
def total_where(items, is_active):
    return sum(item.amount for item in items if is_active(item))

def total_for_orders(orders):
    return total_where(orders, lambda o: o.status == "paid")

def total_for_invoices(invoices):
    return total_where(invoices, lambda i: i.status == "settled")
```

## Options

| Option | Valeur par défaut | Description |
| --- | --- | --- |
| [`clones.type2_threshold`](../configuration/reference.md#clones) | `0.75` | Similarité minimale pour qu'une paire soit signalée comme renommée. |
| [`clones.similarity_threshold`](../configuration/reference.md#clones) | `0.65` | Plancher global appliqué avant les seuils par type. |
| [`clones.ignore_identifiers`](../configuration/reference.md#clones) | `true` | Considère les noms de variables différents comme équivalents lors du calcul de la similarité. |
| [`clones.ignore_literals`](../configuration/reference.md#clones) | `true` | Considère les littéraux numériques et chaînes différents comme équivalents. |
| [`clones.min_lines`](../configuration/reference.md#clones) | `5` | Taille minimale du fragment en lignes. |
| [`clones.enabled_clone_types`](../configuration/reference.md#clones) | `["type1","type2","type4"]` | Incluez `"type2"` pour garder cette règle active. |

## Références

- Implémentation de la détection de clones (`internal/analyzer/clone_detector.go`, `internal/analyzer/apted_tree.go`, `internal/analyzer/apted_cost.go`, `polyscan/core/apted`).
- [Catalogue des règles](index.md) · [Clones identiques](duplicate-code-identical.md) · [Clones modifiés](duplicate-code-modified.md) · [Clones sémantiques](duplicate-code-semantic.md)
