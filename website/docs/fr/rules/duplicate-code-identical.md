# duplicate-code-identical

**Catégorie** : Code dupliqué  
**Sévérité** : Avertissement  
**Déclenchée par** : `pyscn analyze`, `pyscn check --select clones`

## Ce que fait cette règle

Signale deux blocs de code ou plus textuellement identiques, à l'exception de l'espacement, de la mise en page ou des commentaires (clones de Type 1, similarité ≥ 0,85).

## Pourquoi est-ce un problème ?

Le code copié-collé est la forme la moins coûteuse de duplication, mais la plus coûteuse à maintenir. Lorsqu'il faut modifier la logique, chaque copie doit être retrouvée et mise à jour. Un endroit est corrigé, les autres divergent, et l'incohérence devient un bug.

Les blocs identiques gonflent aussi la base de code sans ajouter de comportement. Les lecteurs perdent du temps à confirmer que deux zones sont vraiment identiques au lieu de lire quelque chose de nouveau.

Comme les clones sont littéraux, la correction est presque toujours mécanique : extrayez le bloc dans une fonction et appelez-la depuis les deux endroits.

## Exemple

```python
def send_welcome_email(user):
    subject = "Welcome"
    body = render_template("welcome.html", user=user)
    msg = Message(subject=subject, body=body, to=user.email)
    smtp.send(msg)
    log.info("sent welcome to %s", user.email)

def send_reset_email(user):
    subject = "Reset"
    body = render_template("reset.html", user=user)
    msg = Message(subject=subject, body=body, to=user.email)
    smtp.send(msg)
    log.info("sent reset to %s", user.email)
```

## À utiliser à la place

Extrayez le bloc partagé dans une fonction utilitaire et passez les éléments qui varient.

```python
def send_email(user, subject, template, tag):
    body = render_template(template, user=user)
    msg = Message(subject=subject, body=body, to=user.email)
    smtp.send(msg)
    log.info("sent %s to %s", tag, user.email)

def send_welcome_email(user):
    send_email(user, "Welcome", "welcome.html", "welcome")

def send_reset_email(user):
    send_email(user, "Reset", "reset.html", "reset")
```

## Options

| Option | Valeur par défaut | Description |
| --- | --- | --- |
| [`clones.type1_threshold`](../configuration/reference.md#clones) | `0.85` | Similarité minimale pour qu'une paire soit signalée comme identique. |
| [`clones.similarity_threshold`](../configuration/reference.md#clones) | `0.65` | Plancher global appliqué avant les seuils par type. |
| [`clones.min_lines`](../configuration/reference.md#clones) | `5` | Taille minimale du fragment en lignes. |
| [`clones.min_nodes`](../configuration/reference.md#clones) | `10` | Taille minimale du fragment en nœuds AST. |
| [`clones.enabled_clone_types`](../configuration/reference.md#clones) | `["type1","type2","type4"]` | Incluez `"type1"` pour garder cette règle active. |

## Références

- Implémentation de la détection de clones (`internal/analyzer/clone_detector.go`, `internal/analyzer/apted.go`).
- [Catalogue des règles](index.md) · [Clones renommés](duplicate-code-renamed.md) · [Clones modifiés](duplicate-code-modified.md) · [Clones sémantiques](duplicate-code-semantic.md)
