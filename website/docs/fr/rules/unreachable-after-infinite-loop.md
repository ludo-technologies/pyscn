# unreachable-after-infinite-loop

**Catégorie** : Code inatteignable  
**Sévérité** : Avertissement  
**Déclenchée par** : `pyscn analyze`, `pyscn check`

## Ce que fait cette règle

Signale les instructions qui suivent une boucle sans sortie atteignable, comme un `while True:` sans `break` ni `return`.

## Pourquoi est-ce un problème ?

Si une boucle n'a aucun chemin de sortie, l'exécution ne dépasse jamais ce point. Tout ce qui est écrit après la boucle est mort.

Il s'agit généralement de l'un des cas suivants :

- **Une condition de sortie oubliée** — la boucle était censée se terminer mais le `break` a été perdu lors d'un refactoring.
- **Un nettoyage mal placé** — du code d'arrêt ou de libération situé après une boucle de worker qui ne retourne jamais.
- **Une erreur de copier-coller** — de la logique post-boucle laissée par une version antérieure de la fonction.

Le lecteur s'attend à ce que le code qui suit finisse par s'exécuter. Ce n'est pas le cas.

## Exemple

```python
def run_worker(queue):
    while True:
        job = queue.get()
        job.run()
    queue.close()   # ← ne s'exécute jamais
```

## À utiliser à la place

Donnez à la boucle une sortie atteignable, ou supprimez la suite inatteignable.

```python
def run_worker(queue):
    while not queue.closed:
        job = queue.get()
        job.run()
    queue.close()
```

## Options

| Option | Valeur par défaut | Description |
| --- | --- | --- |
| [`dead_code.enabled`](../configuration/reference.md#dead_code) | `true` | Cette règle n'a pas d'interrupteur dédié ; elle est contrôlée par `dead_code.enabled` et l'analyse CFG. |
| [`dead_code.min_severity`](../configuration/reference.md#dead_code) | `"warning"` | Passez à `"critical"` pour masquer ces constats ; abaissez à `"info"` pour en remonter davantage. |
| [`dead_code.ignore_patterns`](../configuration/reference.md#dead_code) | `[]` | Expressions régulières comparées à la ligne source ; les correspondances sont supprimées. |

## Références

- Analyse d'accessibilité sur le graphe de flot de contrôle (`internal/analyzer/dead_code.go`).
- [Catalogue des règles](index.md) · [Code inatteignable après return](unreachable-after-return.md) · [Branche inatteignable](unreachable-branch.md)
