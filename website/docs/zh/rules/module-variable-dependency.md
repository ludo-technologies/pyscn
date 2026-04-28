# module-variable-dependency

**Category**: 依赖注入  
**Severity**: Warning  
**Triggered by**: `pyscn analyze`, `pyscn check --select di`

## 检测内容

标记在没有 `global` 语句的情况下直接读取或写入模块级可变变量的类 — 一种对模块状态的隐式耦合。

## 为什么这是一个问题

与 `global` 赋值不同，读取模块级名称是无声的：类看起来是自包含的，但其行为实际上取决于调用时模块变量中存放的内容。一个孤立实例化该类的测试可能会因为一个无关的导入而通过或失败。

这也破坏了可替换性。你无法在不对模块进行猴子补丁的情况下给类一个不同的协作对象，而猴子补丁是脆弱且依赖顺序的。同一个类的两个实例被迫共享相同的后端对象，无论你是否希望如此。

将协作对象作为构造函数参数可以记录依赖关系、允许按实例配置，并让测试无需触及模块全局变量即可注入假对象。

## 示例

```python
config = load_config()

class UserRepository:
    def find(self, user_id):
        conn = connect(config.database_url)
        return conn.query("SELECT * FROM users WHERE id = ?", user_id)
```

## 修正示例

将协作对象作为构造函数参数接收。

```python
class UserRepository:
    def __init__(self, database_url):
        self.database_url = database_url

    def find(self, user_id):
        conn = connect(self.database_url)
        return conn.query("SELECT * FROM users WHERE id = ?", user_id)
```

## 选项

| 选项 | 默认值 | 说明 |
| --- | --- | --- |
| [`di.enabled`](../configuration/reference.md#di) | `false` | 必须为 `true` 才能在 `analyze` 中运行 DI 规则。 |
| [`di.min_severity`](../configuration/reference.md#di) | `"warning"` | 提高到 `"error"` 可抑制此规则。 |

## 参考

- 隐式依赖检测（`internal/analyzer/hidden_dependency_detector.go`）。
- [规则目录](index.md) · [Global state dependency](global-state-dependency.md) · [Singleton pattern dependency](singleton-pattern-dependency.md)
