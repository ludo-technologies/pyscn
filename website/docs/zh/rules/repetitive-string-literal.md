# repetitive-string-literal

**Category**: 模拟数据  
**Severity**: Info  
**Triggered by**: `pyscn check --select mockdata`

## 检测内容

标记长度为 4 到 20 且字符模式高度重复的字符串字面量：`aaaa`、`1111`、`xxxxxxxx` 等单字符或双字符重复序列。

## 为什么这是一个问题

像 `"aaaaaaaaaaaaaaaa"` 这样的字符串几乎从来不是真实值——它是开发者在连接其他东西时为了通过验证器而随手输入的。留在生产环境中，它会成为 API 密钥、哈希输入或令牌，看起来像数据并通过长度检查，但没有任何意义。

此规则限定长度范围（4-20 个字符），以避免标记填充常量或确实需要重复字符的测试向量等有意为之的内容。

## 示例

```python
api_key = "aaaaaaaaaaaaaaaa"
```

## 修正示例

从配置或密钥存储中读取密钥和令牌。不要在源代码中嵌入占位符值。

```python
import os

api_key = os.environ["SERVICE_API_KEY"]
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
- [规则目录](index.md) · [test-credential-in-code](test-credential-in-code.md) · [placeholder-uuid](placeholder-uuid.md)
