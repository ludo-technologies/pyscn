# unreachable-after-return

**类别**: 不可达代码  
**严重程度**: Critical  
**触发方式**: `pyscn analyze`, `pyscn check`

## 检测内容

标记同一代码块中 `return` 语句之后出现的语句。

## 为什么这是一个问题

`return` 之后的代码永远不会执行。通常属于以下情况之一：

- **重构遗留** — `return` 被上移，但下方的代码被遗忘了。
- **Bug** — 开发者期望代码会执行，但控制流变更使其变得不可达。
- **错位的清理代码** — 本应在返回之前执行的操作。

无论哪种情况，代码都是死代码：它占用阅读时间，无法被测试覆盖（因为根本执行不到），如果其中隐藏了 bug，该 bug 也永远不会通过用户行为暴露出来。

## 示例

```python
def charge(order):
    if order.total <= 0:
        return None
        log.debug("zero-value charge")   # ← 永远不会执行
    ...
```

## 修正示例

将语句移到 `return` 之前，或者如果不再需要则删除它。

```python
def charge(order):
    if order.total <= 0:
        log.debug("zero-value charge")
        return None
    ...
```

## 选项

| 选项 | 默认值 | 说明 |
| --- | --- | --- |
| [`dead_code.detect_after_return`](../configuration/reference.md#dead_code) | `true` | 设为 `false` 可禁用此规则。 |
| [`dead_code.min_severity`](../configuration/reference.md#dead_code) | `"warning"` | 设为 `"critical"` 仅保留此类发现；设为 `"info"` 可显示更多。 |
| [`dead_code.ignore_patterns`](../configuration/reference.md#dead_code) | `[]` | 用于匹配源代码行的正则表达式模式；匹配项将被抑制。 |

## 参考

- 控制流图可达性分析 (`internal/analyzer/dead_code.go`)。
- [规则目录](index.md) · [不可达分支](unreachable-branch.md) · [raise 后不可达](unreachable-after-raise.md)
