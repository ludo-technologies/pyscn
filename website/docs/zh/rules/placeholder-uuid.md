# placeholder-uuid

**Category**: 模拟数据  
**Severity**: Warning  
**Triggered by**: `pyscn check --select mockdata`

## 检测内容

标记熵值极低的 UUID 格式字符串字面量：全零 UUID（`00000000-0000-0000-0000-000000000000`）、全一、全 `f`，或虽然能解析为 UUID 但包含长段重复字符的值。

## 为什么这是一个问题

全零 UUID 在少数场景中是合法值，但在大多数应用程序代码中，它是桩代码 `DEFAULT_USER_ID = "00..."` 的残留，本应被替换。因为它像其他 UUID 一样能被解析和序列化，外键查找、日志行和审计记录都会接受它——于是行记录开始汇聚到同一个"用户"上，而不会有任何错误被抛出。

## 示例

```python
DEFAULT_USER_ID = "00000000-0000-0000-0000-000000000000"
```

## 修正示例

在使用时生成真实的 UUID，或要求调用者提供。不要携带一个看起来像数据的哨兵值。

```python
import uuid

def new_user_id() -> uuid.UUID:
    return uuid.uuid4()
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
- [规则目录](index.md) · [placeholder-phone-number](placeholder-phone-number.md) · [repetitive-string-literal](repetitive-string-literal.md)
