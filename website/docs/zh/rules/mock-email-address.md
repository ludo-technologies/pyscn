# mock-email-address

**Category**: 模拟数据  
**Severity**: Warning  
**Triggered by**: `pyscn check --select mockdata`

## 检测内容

标记域名部分为测试保留域名的电子邮件地址：`test@example.com`、`admin@test.com`、`foo@localhost` 等。

## 为什么这是一个问题

应用程序代码中的测试域名电子邮件几乎总是测试夹具、教程或"稍后填写"桩代码的残留。与格式错误的地址不同，它能顺利通过验证和序列化，因此会悄悄地出现在数据库行、通知队列和"发件人"标头中。通常的发现路径是一张询问为什么没有人收到重置链接的工单。

此规则是 [mock-domain-in-string](mock-domain-in-string.md) 的补充，但专门处理电子邮件形式，以便保持域名列表精简且匹配精确。

## 示例

```python
admin_email = "admin@example.com"
```

## 修正示例

从配置中读取地址，或将其作为参数接受。

```python
admin_email = settings.admin_email
```

## 选项

| 选项 | 默认值 | 描述 |
| --- | --- | --- |
| [`mock_data.enabled`](../configuration/reference.md#mock_data) | `false` | 需要显式启用。 |
| [`mock_data.domains`](../configuration/reference.md#mock_data) | *（RFC 2606 列表）* | 被视为占位符的域名；与 `mock-domain-in-string` 共享。 |
| [`mock_data.min_severity`](../configuration/reference.md#mock_data) | `"info"` | 提升至 `"warning"` 仅保留此级别的发现。 |
| [`mock_data.ignore_tests`](../configuration/reference.md#mock_data) | `true` | 跳过测试文件。 |
| [`mock_data.ignore_patterns`](../configuration/reference.md#mock_data) | `[]` | 与文件路径匹配的正则表达式模式。 |

## 参考

- RFC 2606: *Reserved Top Level DNS Names.*
- 实现：`internal/analyzer/mock_data_detector.go`。
- [规则目录](index.md) · [mock-domain-in-string](mock-domain-in-string.md) · [test-credential-in-code](test-credential-in-code.md)
