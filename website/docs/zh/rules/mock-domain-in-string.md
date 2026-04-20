# mock-domain-in-string

**Category**: 模拟数据  
**Severity**: Warning  
**Triggered by**: `pyscn check --select mockdata`

## 检测内容

标记包含为文档和测试保留的域名的字符串字面量：`example.com`、`example.org`、`example.net`、`test.com`、`localhost`、`invalid`、`foo.com`、`bar.com`，以及类似的 RFC 2606 / RFC 6761 名称。

## 为什么这是一个问题

这些域名的存在正是为了让示例和测试不会与真实流量冲突。在编写文档时这很有用——但一旦字面量被发布就成了问题。生产环境中硬编码的 `example.com` URL 通常是以下情况之一：

- 一个原本应在发布前被替换的占位符，或
- 一个本不应该被硬编码的配置值。

在这两种情况下，失败模式都是静默的：请求成功（域名解析到文档页面或什么都没有），不会引发异常，只有当有人问为什么注册邮件没有到达时才会发现这个 bug。

## 示例

```python
SIGNUP_URL = "https://example.com/signup"
```

## 修正示例

将值移到配置中，或内联真实 URL。

```python
SIGNUP_URL = settings.signup_url
```

## 选项

| 选项 | 默认值 | 描述 |
| --- | --- | --- |
| [`mock_data.enabled`](../configuration/reference.md#mock_data) | `false` | 需要显式启用。 |
| [`mock_data.domains`](../configuration/reference.md#mock_data) | *（RFC 2606 列表）* | 覆盖或扩展保留域名列表。 |
| [`mock_data.min_severity`](../configuration/reference.md#mock_data) | `"info"` | 提升至 `"warning"` 仅保留此级别的发现。 |
| [`mock_data.ignore_tests`](../configuration/reference.md#mock_data) | `true` | 跳过测试文件。 |
| [`mock_data.ignore_patterns`](../configuration/reference.md#mock_data) | `[]` | 与文件路径匹配的正则表达式模式。 |

## 参考

- RFC 2606: *Reserved Top Level DNS Names.*
- RFC 6761: *Special-Use Domain Names.*
- 实现：`internal/analyzer/mock_data_detector.go`。
- [规则目录](index.md) · [mock-email-address](mock-email-address.md) · [mock-keyword-in-code](mock-keyword-in-code.md)
