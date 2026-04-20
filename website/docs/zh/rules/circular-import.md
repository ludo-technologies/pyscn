# circular-import

**Category**: 模块结构  
**Severity**: Configurable by cycle size (Low / Medium / High / Critical)  
**Triggered by**: `pyscn analyze`, `pyscn check --select circular`

## 检测内容

标记形成导入循环的模块组——模块 A 导入 B（直接或传递地），同时 B 也导入 A。循环通过在模块依赖图上运行 Tarjan 强连通分量算法来发现。

严重程度根据循环的大小和成员的扇入数来分配：

| 循环成员数 | 严重程度 |
| --- | --- |
| 2 | Low |
| 3 – 5 | Medium |
| 6 – 9 | High |
| 10+，或任何成员的扇入 > 10 | Critical |

## 为什么这是一个问题

循环导入意味着两个或更多模块无法独立地被理解、测试或发布。具体而言：

- **导入时错误。** Python 在循环导入期间会部分初始化模块；对半加载模块的属性访问会根据语句顺序引发 `ImportError` 或 `AttributeError`。
- **紧密耦合。** 循环的成员共享一个分散在多个文件中的单一"逻辑模块"。其中一个的更改往往会迫使其他所有成员都进行更改。
- **阻碍重构。** 你无法移动、重命名或删除循环中的任何成员而不触及其他成员。
- **循环越大越糟糕。** 2 个模块的循环是个麻烦；10 个模块的循环是架构失败——因此严重程度会递增。

## 示例

```python
# myapp/orders.py
from myapp.billing import Invoice

class Order:
    def invoice(self) -> Invoice:
        return Invoice(self)
```

```python
# myapp/billing.py
from myapp.orders import Order

class Invoice:
    def __init__(self, order: "Order"):
        self.order = order
```

`orders` 为了返回类型导入了 `billing`；`billing` 为了构造函数参数导入了 `orders`。在顶层运行任一文件都会触发循环。

## 修正示例

将共享类型提取到第三个模块中，使两者都依赖于它而不是彼此：

```python
# myapp/domain.py
class Order: ...
class Invoice: ...
```

```python
# myapp/orders.py
from myapp.domain import Order, Invoice
```

```python
# myapp/billing.py
from myapp.domain import Order, Invoice
```

如果反向边仅用于类型注解，请使用 `TYPE_CHECKING` 保护它，使其不在运行时被求值：

```python
# myapp/billing.py
from __future__ import annotations
from typing import TYPE_CHECKING

if TYPE_CHECKING:
    from myapp.orders import Order

class Invoice:
    def __init__(self, order: "Order"): ...
```

## 选项

| 选项 | 默认值 | 描述 |
| --- | --- | --- |
| [`dependencies.detect_cycles`](../configuration/reference.md#dependencies) | `true` | 设为 `false` 可禁用此规则。 |
| [`dependencies.cycle_reporting`](../configuration/reference.md#dependencies) | `"summary"` | `all`、`critical` 或 `summary`——控制报告中显示多少循环。 |
| [`dependencies.max_cycles_to_show`](../configuration/reference.md#dependencies) | `10` | 报告循环数量的上限。 |
| `--max-cycles N` (check) | `0` | 当循环数量超过 `N` 时使 `check` 命令失败。 |
| `--allow-circular-deps` (check) | off | 将循环降级为警告而非失败。 |

## 参考

- Tarjan SCC 实现（`internal/analyzer/circular_detector.go`），模块图构建（`internal/analyzer/module_analyzer.go`）。
- [规则目录](index.md) · [deep-import-chain](deep-import-chain.md)
