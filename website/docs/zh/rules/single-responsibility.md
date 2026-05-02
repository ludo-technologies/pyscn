# single-responsibility

**Category**: 模块结构  
**Severity**: Warning（当模块同时是高扇入/扇出的枢纽时为 Error）  
**Triggered by**: `pyscn analyze`, `pyscn check --select deps`

## 检测内容

标记承担超过 `architecture.max_responsibilities`（默认 `3`）个不同依赖关注点的模块，或扇入扇出均高于项目平均水平的枢纽模块。

"关注点"由邻居模块的名称推断而来：对于该模块导入的每个模块以及导入它的每个模块，分析器取邻居路径中与当前模块不重叠的第一个段，并跳过通用兜底名称（`base`、`common`、`helpers`、`node`、`shared`、`util`、`utils`）。去重后的段数即为 pyscn 归属于该模块的职责数量。

满足以下任一条件即会报告:

- 拥有的不同关注点数量超过 `max_responsibilities`。
- 扇入（导入该模块的模块数）与扇出（被该模块导入的模块数）均高于项目的"均值 + 标准差"，并且承担多于一个的关注点。

## 为什么这是一个问题

单一职责原则 (SRP) 关心的是"变化的轴"。横跨多个不相关依赖簇的模块拥有多个变化原因:

- **修改会扩散。** 改动一个关注点，迫使你重新阅读和测试共享同一模块边界的其他关注点。
- **导入会撒谎。** `from myapp.core import X` 没有告诉读者任何信息——`core` 在做好几件事。
- **枢纽会成为瓶颈。** 一个被所有人导入、又导入所有人的模块，是变更、评审和合并的单点争用。
- **隐藏了缺失的接缝。** 当两个关注点不断聚集在同一个文件中，正确的解决方案通常是新建一个模块来命名它们之间的关系。

## 示例

```
myapp/core.py
```

```python
# myapp/core.py
from myapp.routers import user_router, order_router
from myapp.services import billing_service, notification_service
from myapp.repositories import user_repo, order_repo
from myapp.telemetry import metrics, tracing

# ...将所有部分粘合到一起的代码...
```

`core` 混合了 `routers` / `services` / `repositories` / `telemetry` 四个关注点，并且 routers 和 services 都会导入它——扇入和扇出都很高。pyscn 将其标记为过载模块。

## 修正示例

按已经存在的关注点拆分模块，每个新模块只命名*一个*变化轴。

```
myapp/wiring/web.py          # 路由层装配
myapp/wiring/services.py     # 服务层装配
myapp/wiring/persistence.py  # 仓储层装配
myapp/wiring/observability.py
```

或者，如果该模块确实是合理的组合根，请缩小其职责：它应当只负责*装配*各部分，而不是同时实现业务规则、定义类型或拥有遥测。

## 选项

| 选项 | 默认值 | 描述 |
| --- | --- | --- |
| [`architecture.validate_responsibility`](../configuration/reference.md#architecture) | `true` | 设为 `false` 可禁用此规则。 |
| [`architecture.max_responsibilities`](../configuration/reference.md#architecture) | `3` | 拥有更多关注点的模块将被标记。 |
| [`architecture.enabled`](../configuration/reference.md#architecture) | `true` | 架构分析的主开关。 |
| [`architecture.fail_on_violations`](../configuration/reference.md#architecture) | `false` | 发现违规时以非零退出码退出。 |

## 参考

- 职责推断与严重程度规则：`service/responsibility_analysis.go`。
- Martin, R. C. *敏捷软件开发：原则、模式与实践*，2002 年（第 8 章 — SRP）。
- [规则目录](index.md) · [low-package-cohesion](low-package-cohesion.md) · [layer-violation](layer-violation.md)
