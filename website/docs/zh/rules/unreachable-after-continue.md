# unreachable-after-continue

**类别**: 不可达代码  
**严重程度**: Critical  
**触发方式**: `pyscn analyze`, `pyscn check`

## 检测内容

标记循环内 `continue` 语句之后出现的语句。

## 为什么这是一个问题

`continue` 会直接跳转到下一次迭代。同一代码块中其后的任何语句在每次迭代时都会被跳过，因此执行次数为零。

典型原因：

- **逻辑重排** — 守卫条件改为 `continue`，但尾部的操作被遗留了。
- **错位的副作用** — 本应在跳过之前执行的计数器更新或日志记录。
- **语义误解** — 开发者期望 `continue` 的行为类似于 `pass`。

由于该语句不可达，测试无法覆盖它，预期的行为默默地不会发生。

## 示例

```python
for order in orders:
    if order.status == "cancelled":
        continue
        metrics.record_skip(order.id)   # ← 永远不会执行
    process(order)
```

## 修正示例

在 `continue` 之前执行语句，或者将其删除。

```python
for order in orders:
    if order.status == "cancelled":
        metrics.record_skip(order.id)
        continue
    process(order)
```

## 选项

| 选项 | 默认值 | 说明 |
| --- | --- | --- |
| [`dead_code.detect_after_continue`](../configuration/reference.md#dead_code) | `true` | 设为 `false` 可禁用此规则。 |
| [`dead_code.min_severity`](../configuration/reference.md#dead_code) | `"warning"` | 设为 `"critical"` 仅保留此类发现；设为 `"info"` 可显示更多。 |
| [`dead_code.ignore_patterns`](../configuration/reference.md#dead_code) | `[]` | 用于匹配源代码行的正则表达式模式；匹配项将被抑制。 |

## 参考

- 控制流图可达性分析 (`internal/analyzer/dead_code.go`)。
- [规则目录](index.md) · [break 后不可达](unreachable-after-break.md) · [return 后不可达](unreachable-after-return.md)
