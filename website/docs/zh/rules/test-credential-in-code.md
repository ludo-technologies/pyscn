# test-credential-in-code

**Category**: 模拟数据  
**Severity**: Warning  
**Triggered by**: `pyscn check --select mockdata`

## 检测内容

标记看起来像明显占位符凭据的字符串字面量：`password123`、`secret123`、`testpassword`、`token0`、`api_key_test` 等由凭据相关词汇加上简单后缀构成的模式。

## 为什么这是一个问题

这些不是真正的密钥——这恰恰就是问题所在。它们是在设置客户端、编写第一个测试或填写必填字段时输入的值。一旦签入，它们有两种失败模式：

- **尴尬/正确性问题**：该字面量被用作默认值并发布给用户，因此"默认管理员密码"真的就是 `password123`。
- **遗漏轮换**：一个真正的密钥原本应在发布前替换占位符；没有人注意到它没有被替换。

pyscn 不是安全扫描器，也不尝试检测高熵密钥。此规则捕获的是相反的情况：低熵、明显伪造的凭据，它们本不应该出现在源代码中。

## 示例

```python
DEFAULT_PASSWORD = "password123"
```

## 修正示例

从环境变量或密钥管理器中读取凭据。如果本地开发确实需要默认值，请将其保存在不会在生产环境中加载的单独配置文件中。

```python
import os

DEFAULT_PASSWORD = os.environ["APP_DEFAULT_PASSWORD"]
```

## 选项

| 选项 | 默认值 | 描述 |
| --- | --- | --- |
| [`mock_data.enabled`](../configuration/reference.md#mock_data) | `false` | 需要显式启用。 |
| [`mock_data.min_severity`](../configuration/reference.md#mock_data) | `"info"` | 提升至 `"warning"` 仅保留此级别的发现。 |
| [`mock_data.ignore_tests`](../configuration/reference.md#mock_data) | `true` | 跳过测试文件——测试凭据属于测试中。 |
| [`mock_data.ignore_patterns`](../configuration/reference.md#mock_data) | `[]` | 与文件路径匹配的正则表达式模式。 |

## 参考

- 实现：`internal/analyzer/mock_data_detector.go`。
- [规则目录](index.md) · [repetitive-string-literal](repetitive-string-literal.md) · [mock-keyword-in-code](mock-keyword-in-code.md)
