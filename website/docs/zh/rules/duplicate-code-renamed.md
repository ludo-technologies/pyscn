# duplicate-code-renamed

**类别**: 重复代码  
**严重程度**: Warning  
**触发方式**: `pyscn analyze`, `pyscn check --select clones`

## 检测内容

标记结构相同但标识符或字面量不同的代码块（Type-2 克隆，相似度 >= 0.75）。

## 为什么这是一个问题

重命名克隆是有人复制一个函数、对变量名执行查找替换然后就完事的产物。结构完全相同；只是名词变了。维护成本与文本完全相同的克隆一样 — 每次修改都必须在多处进行 — 但由于用词不同，这种重复更难通过肉眼发现。

这也是一个信号，说明原始代码在应该参数化时没有被参数化。变化的部分（类型、字段名、常量）天然就是函数参数。

## 示例

```python
def total_for_orders(orders):
    total = 0
    for order in orders:
        if order.status == "paid":
            total += order.amount
    return total

def total_for_invoices(invoices):
    total = 0
    for invoice in invoices:
        if invoice.status == "settled":
            total += invoice.amount
    return total
```

## 修正示例

提取一个通用辅助函数，将变化的谓词和字段访问器作为参数传入。

```python
def total_where(items, is_active):
    return sum(item.amount for item in items if is_active(item))

def total_for_orders(orders):
    return total_where(orders, lambda o: o.status == "paid")

def total_for_invoices(invoices):
    return total_where(invoices, lambda i: i.status == "settled")
```

## 选项

| 选项 | 默认值 | 说明 |
| --- | --- | --- |
| [`clones.type2_threshold`](../configuration/reference.md#clones) | `0.75` | 一对代码被报告为重命名克隆所需的最低相似度。 |
| [`clones.similarity_threshold`](../configuration/reference.md#clones) | `0.65` | 在按类型阈值之前应用的全局下限。 |
| [`clones.ignore_identifiers`](../configuration/reference.md#clones) | `true` | 计算相似度时将不同的变量名视为等价。 |
| [`clones.ignore_literals`](../configuration/reference.md#clones) | `true` | 将不同的数字和字符串字面量视为等价。 |
| [`clones.min_lines`](../configuration/reference.md#clones) | `5` | 最小片段大小（行数）。 |
| [`clones.enabled_clone_types`](../configuration/reference.md#clones) | `["type1","type2","type4"]` | 包含 `"type2"` 以保持此规则生效。 |

## 参考

- 克隆检测实现 (`internal/analyzer/clone_detector.go`, `internal/analyzer/apted.go`)。
- [规则目录](index.md) · [完全相同克隆](duplicate-code-identical.md) · [修改克隆](duplicate-code-modified.md) · [语义克隆](duplicate-code-semantic.md)
