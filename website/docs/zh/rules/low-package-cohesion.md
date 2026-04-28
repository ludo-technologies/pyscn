# low-package-cohesion

**Category**: 模块结构  
**Severity**: Warning  
**Triggered by**: `pyscn analyze`, `pyscn check --select deps`

## 检测内容

标记内部内聚度分数低于 `architecture.min_cohesion`（默认 `0.5`）的包。内聚度衡量的是子模块之间实际包内导入数与可能导入数的比率——一个子模块之间从不相互导入的包得分为 `0`。

## 为什么这是一个问题

包应该是一个恰好被拆分到多个文件中的单一概念。当文件之间互不引用时，该包只是一个共享命名空间的无关代码文件夹：

- **误导性的导入路径。** `from myapp.utils import X` 暗示 `X` 与 `utils` 中的其他内容有关联；低内聚度意味着这个承诺是空的。
- **没有自然的负责人。** 没有人对 "`utils`" 整体负责，因为根本没有整体可言。
- **无限增长。** 杂项包会积累不相关的辅助函数，直到变成一个垃圾场。
- **隐藏了缺失的抽象。** 正确的做法通常不是"继续添加"，而是找到两个子模块真正共享的概念并提取出来。

## 示例

```
myapp/utils/
    __init__.py
    string_utils.py     # slugify, truncate
    math_utils.py       # clamp, lerp
    io_utils.py         # atomic_write, read_json
```

这三个模块没有任何一个导入其他模块。`utils` 包的内聚度为零。

## 修正示例

将包拆分为以其实际功能命名的专注包：

```
myapp/text/          # slugify, truncate, and the helpers they share
myapp/geometry/      # clamp, lerp
myapp/fs/            # atomic_write, read_json
```

或者——如果内容确实是不相关的一次性辅助函数——承认这一点，不再假装有关联。将包命名为 `misc`，或将每个辅助函数移到实际使用它的模块中，并将垃圾场排除在内聚度检查之外。

## 选项

| 选项 | 默认值 | 描述 |
| --- | --- | --- |
| [`architecture.validate_cohesion`](../configuration/reference.md#architecture) | `true` | 设为 `false` 可禁用此规则。 |
| [`architecture.min_cohesion`](../configuration/reference.md#architecture) | `0.5` | 低于此分数的包将被标记。 |
| [`architecture.enabled`](../configuration/reference.md#architecture) | `true` | 架构分析的主开关。 |
| [`architecture.fail_on_violations`](../configuration/reference.md#architecture) | `false` | 发现违规时以非零退出码退出。 |

## 参考

- 包内聚度计算（`internal/analyzer/coupling_metrics.go`、`internal/analyzer/module_analyzer.go`）。
- [规则目录](index.md) · [layer-violation](layer-violation.md)
