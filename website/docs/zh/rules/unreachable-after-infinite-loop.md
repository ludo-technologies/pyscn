# unreachable-after-infinite-loop

**类别**: 不可达代码  
**严重程度**: Warning  
**触发方式**: `pyscn analyze`, `pyscn check`

## 检测内容

标记没有可达出口的循环（例如不含 `break` 或 `return` 的 `while True:`）之后出现的语句。

## 为什么这是一个问题

如果循环没有任何退出路径，执行将永远不会越过该循环。循环之后编写的任何代码都是死代码。

这通常属于以下情况之一：

- **遗忘的退出条件** — 循环本应终止，但 `break` 在重构中丢失了。
- **错位的清理代码** — 关闭或拆卸代码放在了一个永不返回的工作循环之后。
- **复制粘贴错误** — 从函数的先前版本中遗留的循环后逻辑。

读者期望尾部代码最终会执行。但实际上不会。

## 示例

```python
def run_worker(queue):
    while True:
        job = queue.get()
        job.run()
    queue.close()   # ← 永远不会执行
```

## 修正示例

为循环添加可达的退出条件，或者移除不可达的尾部代码。

```python
def run_worker(queue):
    while not queue.closed:
        job = queue.get()
        job.run()
    queue.close()
```

## 选项

| 选项 | 默认值 | 说明 |
| --- | --- | --- |
| [`dead_code.enabled`](../configuration/reference.md#dead_code) | `true` | 此规则没有独立开关；由 `dead_code.enabled` 和 CFG 分析控制。 |
| [`dead_code.min_severity`](../configuration/reference.md#dead_code) | `"warning"` | 提高到 `"critical"` 可隐藏此类发现；降低到 `"info"` 可显示更多。 |
| [`dead_code.ignore_patterns`](../configuration/reference.md#dead_code) | `[]` | 用于匹配源代码行的正则表达式模式；匹配项将被抑制。 |

## 参考

- 控制流图可达性分析 (`internal/analyzer/dead_code.go`)。
- [规则目录](index.md) · [return 后不可达](unreachable-after-return.md) · [不可达分支](unreachable-branch.md)
