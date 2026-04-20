# placeholder-comment

**Category**: 模拟数据  
**Severity**: Info  
**Triggered by**: `pyscn check --select mockdata`

## 检测内容

标记包含未完成工作标记的注释：`TODO`、`FIXME`、`XXX`、`HACK`、`BUG`、`NOTE`。

## 为什么这是一个问题

源代码中的 `# TODO` 是作者对未来自己的承诺，没有截止日期也没有审阅者。大多数代码库积累它们的速度远快于清除速度。每一个都是一小块隐藏的范围——读者必须判断它是否仍然相关、是否阻碍变更，以及是否真的有人在跟踪它。

此规则不是说每个标记都是 bug。它将列表呈现出来，以便你可以根据项目决定是清除它们、将它们转化为已跟踪的 issue，还是通过明确的策略接受它们。

## 示例

```python
def process_order(order):
    # TODO: handle refunds
    ...
```

## 修正示例

要么实现该工作，要么将标记转换为已跟踪的 issue 链接，使意图存在于一个有关闭状态的地方。

```python
def process_order(order):
    # Refunds are handled by the billing service: see issue #1423.
    ...
```

## 选项

| 选项 | 默认值 | 描述 |
| --- | --- | --- |
| [`mock_data.enabled`](../configuration/reference.md#mock_data) | `false` | 需要显式启用。 |
| [`mock_data.min_severity`](../configuration/reference.md#mock_data) | `"info"` | 提升至 `"warning"` 可排除此规则。 |
| [`mock_data.ignore_tests`](../configuration/reference.md#mock_data) | `true` | 跳过测试文件。 |
| [`mock_data.ignore_patterns`](../configuration/reference.md#mock_data) | `[]` | 与文件路径匹配的正则表达式模式。 |

## 参考

- 实现：`internal/analyzer/mock_data_detector.go`。
- [规则目录](index.md) · [mock-keyword-in-code](mock-keyword-in-code.md)
