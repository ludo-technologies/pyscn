# unreachable-branch

**类别**: 不可达代码  
**严重程度**: Warning  
**触发方式**: `pyscn analyze`, `pyscn check`

## 检测内容

标记因前面的每个分支都以 `return`、`raise`、`break` 或 `continue` 终止而无法到达的 `if`、`elif` 或 `else` 分支。

## 为什么这是一个问题

当前面的每个分支都已退出函数或循环时，剩余的分支在逻辑上就是死代码。守卫条件看起来仍然有意义，这会掩盖实际的控制流。

这通常表明：

- **冗余条件** — `else` 之所以存在，只是因为 `if` 曾经是贯穿执行的。
- **隐晦的 bug** — 开发者期望后面的分支在某些情况下执行，但前面的退出使其不可能。
- **过时的防御性代码** — 已经无法到达的回退逻辑。

测试无法覆盖该分支，审查者浪费时间推理一条永远不会执行的路径。

## 示例

```python
def classify(payment):
    if payment.amount < 0:
        raise ValueError("negative amount")
    elif payment.amount == 0:
        return "empty"
    else:
        return "normal"
    return "unknown"   # ← 不可达分支
```

## 修正示例

移除死分支，或重构前面的分支使回退逻辑确实可达。

```python
def classify(payment):
    if payment.amount < 0:
        raise ValueError("negative amount")
    if payment.amount == 0:
        return "empty"
    return "normal"
```

## 选项

| 选项 | 默认值 | 说明 |
| --- | --- | --- |
| [`dead_code.detect_unreachable_branches`](../configuration/reference.md#dead_code) | `true` | 设为 `false` 可禁用此规则。 |
| [`dead_code.min_severity`](../configuration/reference.md#dead_code) | `"warning"` | 提高到 `"critical"` 可隐藏此类发现；降低到 `"info"` 可显示更多。 |
| [`dead_code.ignore_patterns`](../configuration/reference.md#dead_code) | `[]` | 用于匹配源代码行的正则表达式模式；匹配项将被抑制。 |

## 参考

- 控制流图可达性分析 (`internal/analyzer/dead_code.go`)。
- [规则目录](index.md) · [return 后不可达](unreachable-after-return.md) · [raise 后不可达](unreachable-after-raise.md)
