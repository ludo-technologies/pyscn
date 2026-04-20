# low-class-cohesion

**Category**: 类设计  
**Severity**: Configurable by threshold  
**Triggered by**: `pyscn analyze`, `pyscn check`

## 检测内容

标记方法之间不共享实例状态的类（LCOM4 指标 — 方法内聚度缺失，第 4 版）。pyscn 构建一个图，如果两个方法访问了相同的 `self.` 属性，则将它们连接起来，然后计算连通分量的数量。`LCOM4 = 1` 表示每个方法都与其他方法相关；`LCOM4 = N` 表示该类实际上是 `N` 个无关的子类拼凑在一起。

使用 `@staticmethod` 或 `@classmethod` 装饰的方法不引用 `self`，因此被排除在图之外。

简单来说：*这个类在做不相关的工作 — 应该拆分它，或者把它变成一个函数模块。*

## 为什么这是一个问题

类的目的是将状态与操作该状态的操作捆绑在一起。当方法不触及相同的状态时：

- **类名具有误导性** — 它声称自己是一件事，但实际行为像两三件事。
- **变更分散** — 一个职责中的缺陷只能通过阅读与之无关的代码才能发现。
- **复用受阻** — 你无法只提取需要的部分而不连带其余部分。
- **它通常是真正抽象的前身** — "Utilities"或"Manager"类是典型的症状。

## 示例

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

`LCOM4 = 3`：`{load, save}` 共享 `db` 和 `cache`，`{send_welcome, send_reset}` 共享 `smtp`，`{format_joined_at}` 独立存在。三个分量，一个类。

## 修正示例

拆分为内聚的类，并将无状态的部分提取为自由函数。

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

每个类现在的 `LCOM4 = 1`，格式化函数也成为了它应有的一行函数。

## 选项

| 选项 | 默认值 | 说明 |
| --- | --- | --- |
| [`lcom.low_threshold`](../configuration/reference.md#lcom) | `2` | 等于或低于此值，类报告为低风险。 |
| [`lcom.medium_threshold`](../configuration/reference.md#lcom) | `5` | 高于此值，类为高风险。 |

## 参考

- Hitz, M. & Montazeri, B. *Chidamber and Kemerer's Metrics Suite: A Measurement Theory Perspective.* IEEE TSE, 1996（LCOM4 定义）。
- 实现：`internal/analyzer/lcom.go`。
- [规则目录](index.md) · [high-class-coupling](high-class-coupling.md)
