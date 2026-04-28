# mock-keyword-in-code

**Category**: 模拟数据  
**Severity**: Info (in strings) / Warning (in identifiers)  
**Triggered by**: `pyscn check --select mockdata`

## 检测内容

标记包含常见占位符关键词的标识符和字符串字面量：`mock`、`fake`、`dummy`、`test`、`sample`、`example`、`placeholder`、`stub`、`fixture`、`temp`、`foo`、`bar`、`baz`、`lorem`、`ipsum`。

## 为什么这是一个问题

这些词是你在还没想清楚时随手输入的。它们在笔记本、测试文件或五分钟原型中没有问题——但一旦残留到生产环境中，通常意味着一个桩代码从未被替换。签入模块中的 `foo` 几乎从来不是作者打算发布的内容。

**标识符**中的匹配被视为警告，因为绑定名称（`foo = get_user()`）会改变行为。**字符串字面量**中的匹配为信息级别，因为残留的 `"fake_user"` 更多是外观问题而非功能故障——但在发布前仍值得审查。

## 示例

```python
def create_user():
    name = "fake_user"    # string literal matches
    foo = get_user()      # identifier `foo` matches
    return foo
```

## 修正示例

移除占位符。使用真实数据，从配置中读取，或将桩代码移到它应该存在的测试夹具中。

```python
def create_user(name: str):
    user = get_user(name)
    return user
```

## 选项

| 选项 | 默认值 | 描述 |
| --- | --- | --- |
| [`mock_data.enabled`](../configuration/reference.md#mock_data) | `false` | 需要显式启用；整个类别默认关闭。 |
| [`mock_data.keywords`](../configuration/reference.md#mock_data) | *（内置列表）* | 覆盖此规则的关键词列表。 |
| [`mock_data.min_severity`](../configuration/reference.md#mock_data) | `"info"` | 设为 `"warning"` 仅保留标识符匹配。 |
| [`mock_data.ignore_tests`](../configuration/reference.md#mock_data) | `true` | 跳过看起来像测试的文件。 |
| [`mock_data.ignore_patterns`](../configuration/reference.md#mock_data) | `[]` | 与文件路径匹配的正则表达式模式；匹配的结果将被抑制。 |

## 参考

- 实现：`internal/analyzer/mock_data_detector.go`。
- [规则目录](index.md) · [mock-domain-in-string](mock-domain-in-string.md) · [test-credential-in-code](test-credential-in-code.md) · [placeholder-comment](placeholder-comment.md)
