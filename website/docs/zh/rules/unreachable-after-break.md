# unreachable-after-break

**类别**: 不可达代码  
**严重程度**: Critical  
**触发方式**: `pyscn analyze`, `pyscn check`

## 检测内容

标记循环内 `break` 语句之后出现的语句。

## 为什么这是一个问题

`break` 会立即退出所在循环。同一代码块中其后的任何语句都不会执行。

常见原因：

- **错位的递增或累加器更新** — 开发者期望它在最后一次迭代中执行。
- **遗留的日志或清理代码** — 在重构过程中被移到了 `break` 之后。
- **控制流混淆** — 开发者期望 `break` 只跳过迭代的一部分。

该代码不可达，因此测试无法覆盖它，其中的 bug 也永远不会暴露。

## 示例

```python
for user in users:
    if user.id == target_id:
        break
        user.last_seen = now()   # ← 永远不会执行
```

## 修正示例

在 `break` 之前执行操作，或者移除死代码语句。

```python
for user in users:
    if user.id == target_id:
        user.last_seen = now()
        break
```

## 选项

| 选项 | 默认值 | 说明 |
| --- | --- | --- |
| [`dead_code.detect_after_break`](../configuration/reference.md#dead_code) | `true` | 设为 `false` 可禁用此规则。 |
| [`dead_code.min_severity`](../configuration/reference.md#dead_code) | `"warning"` | 设为 `"critical"` 仅保留此类发现；设为 `"info"` 可显示更多。 |
| [`dead_code.ignore_patterns`](../configuration/reference.md#dead_code) | `[]` | 用于匹配源代码行的正则表达式模式；匹配项将被抑制。 |

## 参考

- 控制流图可达性分析 (`internal/analyzer/dead_code.go`)。
- [规则目录](index.md) · [continue 后不可达](unreachable-after-continue.md) · [return 后不可达](unreachable-after-return.md)
