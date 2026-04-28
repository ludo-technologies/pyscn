# duplicate-code-modified

**类别**: 重复代码  
**严重程度**: Info  
**触发方式**: `pyscn analyze`, `pyscn check --select clones`

## 检测内容

标记共享大部分结构但有添加、删除或修改语句的代码块（Type-3 克隆，相似度 >= 0.70）。默认禁用；将 `"type3"` 添加到 `clones.enabled_clone_types` 可启用。

## 为什么这是一个问题

近似重复是真实代码库中最常见的克隆类型，通常也是最难清理的。两个函数做的事情几乎相同，但其中一个多了一步验证或错误处理路径略有不同。共享部分应存在于一处；不同之处应该是调用点之间唯一的差异。

如果不加处理，修改克隆会逐渐偏离：一处副本修复了 bug，另一处没有；一处新增了功能，另一处被遗忘。审查者不再信任外观相似的代码行为确实相同。

此规则为 Info 级别，因为正确的重构不像 Type-1 或 Type-2 克隆那样机械化 — 有时差异本身就是重点，合并反而会损害清晰度。将发现视为值得查看的候选项，而非自动缺陷。

## 示例

```python
def export_users_csv(users, path):
    with open(path, "w") as f:
        writer = csv.writer(f)
        writer.writerow(["id", "name", "email"])
        for u in users:
            writer.writerow([u.id, u.name, u.email])

def export_orders_csv(orders, path):
    with open(path, "w") as f:
        writer = csv.writer(f)
        writer.writerow(["id", "total", "status"])
        for o in orders:
            if o.status != "draft":
                writer.writerow([o.id, o.total, o.status])
```

## 修正示例

提取通用脚手架，将表头、行格式和过滤条件参数化。

```python
def export_csv(rows, path, header, to_row, keep=lambda _: True):
    with open(path, "w") as f:
        writer = csv.writer(f)
        writer.writerow(header)
        for row in rows:
            if keep(row):
                writer.writerow(to_row(row))

def export_users_csv(users, path):
    export_csv(users, path, ["id", "name", "email"],
               lambda u: [u.id, u.name, u.email])

def export_orders_csv(orders, path):
    export_csv(orders, path, ["id", "total", "status"],
               lambda o: [o.id, o.total, o.status],
               keep=lambda o: o.status != "draft")
```

## 选项

| 选项 | 默认值 | 说明 |
| --- | --- | --- |
| [`clones.type3_threshold`](../configuration/reference.md#clones) | `0.70` | 一对代码被报告为修改克隆所需的最低相似度。 |
| [`clones.enabled_clone_types`](../configuration/reference.md#clones) | `["type1","type2","type4"]` | 添加 `"type3"` 以启用此规则。 |
| [`clones.similarity_threshold`](../configuration/reference.md#clones) | `0.65` | 在按类型阈值之前应用的全局下限。 |
| [`clones.min_lines`](../configuration/reference.md#clones) | `5` | 最小片段大小（行数）。 |
| [`clones.min_nodes`](../configuration/reference.md#clones) | `10` | 最小片段大小（AST 节点数）。 |

## 参考

- 克隆检测实现 (`internal/analyzer/clone_detector.go`, `internal/analyzer/apted.go`)。
- [规则目录](index.md) · [完全相同克隆](duplicate-code-identical.md) · [重命名克隆](duplicate-code-renamed.md) · [语义克隆](duplicate-code-semantic.md)
