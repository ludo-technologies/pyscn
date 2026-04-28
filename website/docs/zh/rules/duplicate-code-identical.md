# duplicate-code-identical

**类别**: 重复代码  
**严重程度**: Warning  
**触发方式**: `pyscn analyze`, `pyscn check --select clones`

## 检测内容

标记除空白、排版或注释外文本完全相同的两个或多个代码块（Type-1 克隆，相似度 >= 0.85）。

## 为什么这是一个问题

复制粘贴的代码是最廉价的重复形式，却是维护成本最高的。当逻辑需要修改时，必须找到并更新每一份副本。修复了一处，其他副本逐渐偏离，不一致就变成了 bug。

完全相同的代码块还会在不增加任何行为的情况下膨胀代码库。读者花费时间确认两个区域确实相同，而不是阅读新的内容。

由于克隆是字面相同的，修复方法几乎总是机械化的：将代码块提取为函数，然后在两处调用它。

## 示例

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

## 修正示例

将共享代码块提取为辅助函数，将变化的部分作为参数传入。

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

## 选项

| 选项 | 默认值 | 说明 |
| --- | --- | --- |
| [`clones.type1_threshold`](../configuration/reference.md#clones) | `0.85` | 一对代码被报告为完全相同所需的最低相似度。 |
| [`clones.similarity_threshold`](../configuration/reference.md#clones) | `0.65` | 在按类型阈值之前应用的全局下限。 |
| [`clones.min_lines`](../configuration/reference.md#clones) | `5` | 最小片段大小（行数）。 |
| [`clones.min_nodes`](../configuration/reference.md#clones) | `10` | 最小片段大小（AST 节点数）。 |
| [`clones.enabled_clone_types`](../configuration/reference.md#clones) | `["type1","type2","type4"]` | 包含 `"type1"` 以保持此规则生效。 |

## 参考

- 克隆检测实现 (`internal/analyzer/clone_detector.go`, `internal/analyzer/apted.go`)。
- [规则目录](index.md) · [重命名克隆](duplicate-code-renamed.md) · [修改克隆](duplicate-code-modified.md) · [语义克隆](duplicate-code-semantic.md)
