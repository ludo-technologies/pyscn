# global-state-dependency

**Category**: 依赖注入  
**Severity**: Error  
**Triggered by**: `pyscn analyze`, `pyscn check --select di`

## 检测内容

标记使用 `global` 语句读取或修改模块级状态的类方法。

## 为什么这是一个问题

方法中的 `global` 将类绑定到一个特定的模块变量，而该变量在类的接口中不可见。`OrderService(...)` 的签名不会告诉阅读者仅构造它是不够的 — 还必须先准备好某个模块级的值，否则方法会出现意外行为。

测试受到的影响最大。要测试一个触及全局状态的方法，每个测试都必须深入模块、保存旧值、设置新值，并在清理阶段恢复 — 任何忘记清理的测试都会将状态泄漏到下一个测试。并行运行测试也变得不安全。

依赖是真实存在的；只是被隐藏了。将它作为显式的构造函数参数可以消除这种意外。

## 示例

```python
_current_user = None

class AuditLog:
    def record(self, action):
        global _current_user
        entry = {"user": _current_user, "action": action}
        db.insert("audit", entry)
```

## 修正示例

通过 `__init__` 传入值，使依赖可见且可替换。

```python
class AuditLog:
    def __init__(self, user):
        self.user = user

    def record(self, action):
        db.insert("audit", {"user": self.user, "action": action})
```

## 选项

| 选项 | 默认值 | 说明 |
| --- | --- | --- |
| [`di.enabled`](../configuration/reference.md#di) | `false` | 必须为 `true` 才能在 `analyze` 中运行 DI 规则。 |
| [`di.min_severity`](../configuration/reference.md#di) | `"warning"` | 此规则以 `error` 级别报告；除非将 `min_severity` 提高到 `error` 以上，否则会显示此规则。 |

## 参考

- 隐式依赖检测（`internal/analyzer/hidden_dependency_detector.go`）。
- [规则目录](index.md) · [Module variable dependency](module-variable-dependency.md) · [Singleton pattern dependency](singleton-pattern-dependency.md)
