# deep-import-chain

**Category**: 模块结构  
**Severity**: Info  
**Triggered by**: `pyscn analyze`, `pyscn check --select deps`

## 检测内容

当项目中最长的无环导入链深度超过该项目规模的预期深度时进行报告。pyscn 使用 `log₂(module_count) + 1` 作为参考——一个 64 个模块的项目预期链长度不超过 7。

链是模块依赖图中的一条路径：`a → b → c → …`，其中每个箭头代表一个 `import`。

## 为什么这是一个问题

过深的链表明分层不良。每增加一个环节就多一个必须被加载、解析和初始化后链底部才能使用的模块，而且每个环节都是无关更改可能向下传播的地方。

链过深的症状：

- **启动缓慢。** 导入叶模块会触发一连串的顶层副作用。
- **测试脆弱。** 叶模块的单元测试会拉入整条链，当上游任何内容发生变化时就会中断。
- **隐藏的耦合。** 链中间的模块通常只是作为传递层存在，掩盖了真正的依赖关系。
- **难以推理。** 代码没有单一的"层级"可以归属。

## 示例

```
myapp.cli
  → myapp.commands
    → myapp.services
      → myapp.orchestrator
        → myapp.workers
          → myapp.adapters
            → myapp.drivers
```

到达驱动层需要七个层级。实际上 CLI 层不需要知道 workers 的存在，workers 也不需要知道 CLI 的存在——但 `drivers` 的更改可能迫使其上方的每一层都重新测试。

## 修正示例

在边界处引入门面，使上层与一个模块交互而非一条链：

```
myapp.cli
  → myapp.commands
    → myapp.services        # single entry point
        (internally wires orchestrator / workers / adapters / drivers)
```

或者扁平化：如果 `services`、`orchestrator` 和 `workers` 都在做协调工作，将它们合并为一层，让其直接依赖 `adapters`。

## 选项

| 选项 | 默认值 | 描述 |
| --- | --- | --- |
| [`dependencies.find_long_chains`](../configuration/reference.md#dependencies) | `true` | 设为 `false` 可禁用此规则。 |
| [`dependencies.enabled`](../configuration/reference.md#dependencies) | `false` | `pyscn check` 需要显式启用；`pyscn analyze` 始终开启。 |

没有显式的深度阈值——pyscn 将最长链与 `log₂(module_count) + 1` 进行比较，超出时进行报告。

## 参考

- 模块 DAG 上的最长路径搜索（`internal/analyzer/module_analyzer.go`、`internal/analyzer/coupling_metrics.go`）。
- [规则目录](index.md) · [circular-import](circular-import.md)
