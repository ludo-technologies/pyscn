# low-class-cohesion

**Catégorie** : Conception de classes  
**Sévérité** : Configurable par seuil  
**Déclenché par** : `pyscn analyze`, `pyscn check`

## Ce que fait cette règle

Signale les classes dont les méthodes ne partagent pas d'état d'instance (la métrique LCOM4 — Lack of Cohesion of Methods, version 4). pyscn construit un graphe dans lequel deux méthodes sont connectées si elles touchent un attribut `self.` commun, puis compte les composantes connexes. `LCOM4 = 1` signifie que chaque méthode est liée à toutes les autres ; `LCOM4 = N` signifie que la classe est en réalité `N` sous-classes indépendantes collées ensemble.

Les méthodes décorées avec `@staticmethod` ou `@classmethod` ne référencent pas `self` et sont exclues du graphe.

En clair : *cette classe accomplit des tâches sans rapport — scindez-la, ou faites-en un module de fonctions.*

## Pourquoi est-ce un problème ?

Une classe est censée regrouper un état avec les opérations qui agissent dessus. Lorsque les méthodes ne touchent pas le même état :

- **Le nom de la classe ment** — elle prétend être une chose mais se comporte comme deux ou trois.
- **Les changements se dispersent** — un bogue dans une responsabilité ne peut être trouvé qu'en lisant du code qui n'a rien à voir avec lui.
- **La réutilisation est bloquée** — vous ne pouvez pas extraire la partie dont vous avez besoin sans entraîner tout le reste.
- **Cela précède souvent une vraie abstraction** — la classe « Utilities » ou « Manager » est un symptôme classique.

## Exemple

```python
class UserUtility:
    def __init__(self, db, smtp, clock):
        self.db = db
        self.smtp = smtp
        self.clock = clock
        self.cache = {}

    # --- persistence ---
    def load(self, user_id):
        if user_id in self.cache:
            return self.cache[user_id]
        row = self.db.fetch("users", user_id)
        self.cache[user_id] = row
        return row

    def save(self, user):
        self.db.upsert("users", user)
        self.cache[user.id] = user

    # --- email ---
    def send_welcome(self, address):
        self.smtp.send(address, "Welcome")

    def send_reset(self, address, token):
        self.smtp.send(address, f"Reset: {token}")

    # --- formatting ---
    def format_joined_at(self, user):
        return self.clock.format(user.joined_at)
```

`LCOM4 = 3` : `{load, save}` partagent `db` et `cache`, `{send_welcome, send_reset}` partagent `smtp`, `{format_joined_at}` est isolé. Trois composantes, une seule classe.

## À utiliser à la place

Scindez en classes cohésives et déplacez la partie sans état vers des fonctions libres.

```python
class UserRepository:
    def __init__(self, db):
        self._db = db
        self._cache = {}

    def load(self, user_id):
        if user_id in self._cache:
            return self._cache[user_id]
        row = self._db.fetch("users", user_id)
        self._cache[user_id] = row
        return row

    def save(self, user):
        self._db.upsert("users", user)
        self._cache[user.id] = user


class UserMailer:
    def __init__(self, smtp):
        self._smtp = smtp

    def send_welcome(self, address):
        self._smtp.send(address, "Welcome")

    def send_reset(self, address, token):
        self._smtp.send(address, f"Reset: {token}")


# user_formatting.py — no class, no state
def format_joined_at(user, clock):
    return clock.format(user.joined_at)
```

Chaque classe a maintenant `LCOM4 = 1`, et le formateur est une fonction d'une ligne, là où il doit être.

## Options

| Option | Défaut | Description |
| --- | --- | --- |
| [`lcom.low_threshold`](../configuration/reference.md#lcom) | `2` | À ce seuil ou en dessous, la classe est signalée comme à faible risque. |
| [`lcom.medium_threshold`](../configuration/reference.md#lcom) | `5` | Au-dessus de ce seuil, la classe est à risque élevé. |

## Références

- Hitz, M. & Montazeri, B. *Chidamber and Kemerer's Metrics Suite: A Measurement Theory Perspective.* IEEE TSE, 1996 (définition de LCOM4).
- Implémentation : `internal/analyzer/lcom.go`.
- [Catalogue des règles](index.md) · [high-class-coupling](high-class-coupling.md)
