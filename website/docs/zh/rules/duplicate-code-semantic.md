# duplicate-code-semantic

**类别**: 重复代码  
**严重程度**: Warning  
**触发方式**: `pyscn analyze`, `pyscn check --select clones`

## 检测内容

标记语法不同但计算结果相同的代码块（Type-4 克隆，相似度 >= 0.65）。使用数据流分析比较行为而非结构。

## 为什么这是一个问题

语义克隆是在代码审查中不会被注意到的重复。一个函数使用循环，另一个使用推导式；一个通过 `update` 构建字典，另一个使用合并语法。代码看起来不同，因此通过了视觉检查，但两种实现做的是同一件事。

风险与任何重复一样 — 修改必须在多处进行 — 但还有额外的成本。读者无法一眼判断两种实现在边界情况上是否一致。循环版本是否也会跳过 `None`？推导式版本在输入为空时是否会抛出异常？每次都必须进行这样的心理审计。

统一为单一实现可以消除审计和偏离的问题。

## 示例

```python
def unique_emails(users):
    seen = set()
    result = []
    for u in users:
        if u.email not in seen:
            seen.add(u.email)
            result.append(u.email)
    return result

def distinct_emails(users):
    return list({u.email: None for u in users}.keys())
```

## 修正示例

选择一种实现并在所有地方使用。优先选择更清晰的版本；如果两者各有优点，保留一个并记录原因。

```python
def unique_emails(users):
    """按首次出现顺序返回用户邮箱，不含重复项。"""
    return list(dict.fromkeys(u.email for u in users))
```

## 选项

| 选项 | 默认值 | 说明 |
| --- | --- | --- |
| [`clones.type4_threshold`](../configuration/reference.md#clones) | `0.65` | 一对代码被报告为语义克隆所需的最低相似度。 |
| [`clones.enable_dfa`](../configuration/reference.md#clones) | `true` | 启用驱动 Type-4 检测的数据流分析。 |
| [`clones.similarity_threshold`](../configuration/reference.md#clones) | `0.65` | 在按类型阈值之前应用的全局下限。 |
| [`clones.enabled_clone_types`](../configuration/reference.md#clones) | `["type1","type2","type4"]` | 包含 `"type4"` 以保持此规则生效。 |
| [`clones.min_lines`](../configuration/reference.md#clones) | `5` | 最小片段大小（行数）。 |

## 参考

- 克隆检测实现 (`internal/analyzer/clone_detector.go`, `internal/analyzer/apted.go`)。
- [规则目录](index.md) · [完全相同克隆](duplicate-code-identical.md) · [重命名克隆](duplicate-code-renamed.md) · [修改克隆](duplicate-code-modified.md)
