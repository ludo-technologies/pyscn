# placeholder-phone-number

**Category**: 模拟数据  
**Severity**: Warning  
**Triggered by**: `pyscn check --select mockdata`

## 检测内容

标记字符串字面量中明显伪造模式的电话号码：全零（`000-0000-0000`）、连续数字（`123-456-7890`、`012-345-6789`）或长段重复数字。

## 为什么这是一个问题

占位符电话号码是那种从表单的初稿就存在、之后再也没有人回顾过的值。它能通过验证、格式化，也能在数据库中往返存取——因此在真实用户在确认页面上看到它或客服人员尝试拨打之前，什么都不会出错。

## 示例

```python
default_phone = "000-0000-0000"
```

## 修正示例

将字段留空，要求调用者提供，或从配置中获取。不知道的电话号码应该缺失，而不是伪造。

```python
default_phone: str | None = None
```

## 选项

| 选项 | 默认值 | 描述 |
| --- | --- | --- |
| [`mock_data.enabled`](../configuration/reference.md#mock_data) | `false` | 需要显式启用。 |
| [`mock_data.min_severity`](../configuration/reference.md#mock_data) | `"info"` | 提升至 `"warning"` 仅保留此级别的发现。 |
| [`mock_data.ignore_tests`](../configuration/reference.md#mock_data) | `true` | 跳过测试文件。 |
| [`mock_data.ignore_patterns`](../configuration/reference.md#mock_data) | `[]` | 与文件路径匹配的正则表达式模式。 |

## 参考

- 实现：`internal/analyzer/mock_data_detector.go`。
- [规则目录](index.md) · [placeholder-uuid](placeholder-uuid.md) · [repetitive-string-literal](repetitive-string-literal.md)
