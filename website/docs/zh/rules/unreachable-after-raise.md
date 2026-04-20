# unreachable-after-raise

**类别**: 不可达代码  
**严重程度**: Critical  
**触发方式**: `pyscn analyze`, `pyscn check`

## 检测内容

标记同一代码块中 `raise` 语句之后出现的语句。

## 为什么这是一个问题

`raise` 会无条件地展开调用栈。同一代码块中其后的任何语句都不会执行。

这通常指向以下问题：

- **过时的清理代码** — 本应在异常抛出之前执行的代码。
- **重构产物** — `raise` 替换了先前的分支，周围的代码被遗留了下来。
- **逻辑错误** — 开发者假设执行会在 `raise` 之后继续。

`raise` 之后的死代码永远不会被测试覆盖，也不会在生产环境中暴露，因此其中隐藏的任何 bug 都是无声的。

## 示例

```python
def withdraw(account, amount):
    if amount > account.balance:
        raise InsufficientFunds(account.id)
        account.balance -= amount   # ← 永远不会执行
    account.balance -= amount
```

## 修正示例

将语句移到 `raise` 之前，或者删除它。

```python
def withdraw(account, amount):
    if amount > account.balance:
        raise InsufficientFunds(account.id)
    account.balance -= amount
```

## 选项

| 选项 | 默认值 | 说明 |
| --- | --- | --- |
| [`dead_code.detect_after_raise`](../configuration/reference.md#dead_code) | `true` | 设为 `false` 可禁用此规则。 |
| [`dead_code.min_severity`](../configuration/reference.md#dead_code) | `"warning"` | 设为 `"critical"` 仅保留此类发现；设为 `"info"` 可显示更多。 |
| [`dead_code.ignore_patterns`](../configuration/reference.md#dead_code) | `[]` | 用于匹配源代码行的正则表达式模式；匹配项将被抑制。 |

## 参考

- 控制流图可达性分析 (`internal/analyzer/dead_code.go`)。
- [规则目录](index.md) · [return 后不可达](unreachable-after-return.md) · [不可达分支](unreachable-branch.md)
