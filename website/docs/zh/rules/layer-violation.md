# layer-violation

**Category**: 模块结构  
**Severity**: Configurable via `architecture.rules[].severity`  
**Triggered by**: `pyscn analyze`, `pyscn check --select deps`

## 检测内容

当源模块所在层不允许依赖目标模块所在层时，标记该 `import` 语句。层的允许关系根据你配置的 `[[architecture.rules]]` 来判定。层通过匹配 `[[architecture.layers]]` 中定义的包名片段来分配给模块。

## 为什么这是一个问题

分层架构只有在层级保持完整时才有价值。从 `presentation` 到 `infrastructure` 的一次捷径就足以：

- **破坏可测试性。** 表示层现在只能在背后有真实数据库/HTTP 客户端的情况下才能进行测试。
- **产生隐藏的耦合。** 替换基础设施实现会悄悄破坏本不应知道其存在的 UI 代码。
- **使违规正常化。** 一旦存在一个捷径，下一个就更容易被合理化。

此规则是你在设计文档中已经画好的架构图的自动化执行。

## 示例

配置：

```toml
[[architecture.layers]]
name = "presentation"
packages = ["api", "handlers"]

[[architecture.layers]]
name = "application"
packages = ["services", "usecases"]

[[architecture.layers]]
name = "infrastructure"
packages = ["repositories", "db"]

[[architecture.rules]]
from = "presentation"
allow = ["application"]
deny = ["infrastructure"]
```

违规代码：

```python
# myapp/api/orders.py  (presentation)
from myapp.repositories.orders import OrderRepository   # ← forbidden

def list_orders():
    return OrderRepository().all()
```

`presentation` 越过 `application` 直接访问了 `infrastructure`。

## 修正示例

通过应用层路由调用：

```python
# myapp/services/orders.py  (application)
from myapp.repositories.orders import OrderRepository

def list_orders():
    return OrderRepository().all()
```

```python
# myapp/api/orders.py  (presentation)
from myapp.services.orders import list_orders

def get():
    return list_orders()
```

`api` 现在只依赖 `services`，基础设施可以在不触及表示层的情况下进行替换。

## 选项

| 选项 | 默认值 | 描述 |
| --- | --- | --- |
| [`[[architecture.layers]]`](../configuration/reference.md#architecture) | — | 定义层及属于每层的包名片段。 |
| [`[[architecture.rules]]`](../configuration/reference.md#architecture) | — | `from` / `allow` / `deny` / 可选的每条规则 `severity`。 |
| [`architecture.validate_layers`](../configuration/reference.md#architecture) | `true` | 设为 `false` 可禁用此规则。 |
| [`architecture.strict_mode`](../configuration/reference.md#architecture) | `true` | 严格模式下，未显式允许的依赖一律被拒绝。 |
| [`architecture.fail_on_violations`](../configuration/reference.md#architecture) | `false` | 发现违规时以非零退出码退出。 |

未配置层时，分析器以宽松模式运行，此规则不会产生任何发现。

## 参考

- 层解析和规则评估（`internal/analyzer/module_analyzer.go`）。
- [规则目录](index.md) · [low-package-cohesion](low-package-cohesion.md)
