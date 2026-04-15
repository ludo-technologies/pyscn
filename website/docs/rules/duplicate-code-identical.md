# duplicate-code-identical

**Category**: Duplicate Code  
**Severity**: Warning  
**Triggered by**: `pyscn analyze`, `pyscn check --select clones`

## What it does

Flags two or more code blocks that are textually identical except for whitespace, layout, or comments (Type-1 clones, similarity ≥ 0.85).

## Why is this a problem?

Copy-pasted code is the cheapest form of duplication and the most expensive to maintain. When the logic needs to change, every copy has to be found and updated. One site gets fixed, the others drift, and the inconsistency becomes a bug.

Identical blocks also inflate the codebase without adding behaviour. Readers spend time confirming that two regions really are the same instead of reading something new.

Because the clones are literal, the fix is almost always mechanical: lift the block into a function and call it from both places.

## Example

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

## Use instead

Extract the shared block into a helper and pass the parts that vary.

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

| Option | Default | Description |
| --- | --- | --- |
| [`clones.type1_threshold`](../configuration/reference.md#clones) | `0.85` | Minimum similarity for a pair to be reported as identical. |
| [`clones.similarity_threshold`](../configuration/reference.md#clones) | `0.65` | Global floor applied before per-type thresholds. |
| [`clones.min_lines`](../configuration/reference.md#clones) | `5` | Minimum fragment size in lines. |
| [`clones.min_nodes`](../configuration/reference.md#clones) | `10` | Minimum fragment size in AST nodes. |
| [`clones.enabled_clone_types`](../configuration/reference.md#clones) | `["type1","type2","type4"]` | Include `"type1"` to keep this rule active. |

## References

- Clone detection implementation (`internal/analyzer/clone_detector.go`, `internal/analyzer/apted.go`).
- [Rule catalog](index.md) · [Renamed clones](duplicate-code-renamed.md) · [Modified clones](duplicate-code-modified.md) · [Semantic clones](duplicate-code-semantic.md)
