# 规则目录

pyscn 提供 7 个类别共 33 条规则。每条规则都有对应页面，说明其检测内容、问题原因、错误示例以及修正方法。

点击规则名称可打开对应页面。

## 不可达代码

永远无法执行的死代码。通过控制流图可达性分析检测。

| 规则 | 严重程度 |
| ---- | -------- |
| [`unreachable-after-return`](unreachable-after-return.md) | Critical |
| [`unreachable-after-raise`](unreachable-after-raise.md) | Critical |
| [`unreachable-after-break`](unreachable-after-break.md) | Critical |
| [`unreachable-after-continue`](unreachable-after-continue.md) | Critical |
| [`unreachable-after-infinite-loop`](unreachable-after-infinite-loop.md) | Warning |
| [`unreachable-branch`](unreachable-branch.md) | Warning |

## 重复代码

项目中复制粘贴或近似复制粘贴的代码片段。

| 规则 | 严重程度 |
| ---- | -------- |
| [`duplicate-code-identical`](duplicate-code-identical.md) | Warning |
| [`duplicate-code-renamed`](duplicate-code-renamed.md) | Warning |
| [`duplicate-code-modified`](duplicate-code-modified.md) | Info (可选启用) |
| [`duplicate-code-semantic`](duplicate-code-semantic.md) | Warning |

## 复杂度

分支过多、难以测试或可靠推理的函数。

| 规则 | 严重程度 |
| ---- | -------- |
| [`high-cyclomatic-complexity`](high-cyclomatic-complexity.md) | 按阈值 |

## 类设计

依赖过多或承担过多无关职责的类。

| 规则 | 严重程度 |
| ---- | -------- |
| [`high-class-coupling`](high-class-coupling.md) | 按阈值 |
| [`low-class-cohesion`](low-class-cohesion.md) | 按阈值 |

## 依赖注入

影响可测试性的构造函数和协作者模式。

| 规则 | 严重程度 |
| ---- | -------- |
| [`too-many-constructor-parameters`](too-many-constructor-parameters.md) | Warning |
| [`global-state-dependency`](global-state-dependency.md) | Error |
| [`module-variable-dependency`](module-variable-dependency.md) | Warning |
| [`singleton-pattern-dependency`](singleton-pattern-dependency.md) | Warning |
| [`concrete-type-hint-dependency`](concrete-type-hint-dependency.md) | Info |
| [`concrete-instantiation-dependency`](concrete-instantiation-dependency.md) | Warning |
| [`service-locator-pattern`](service-locator-pattern.md) | Warning |

## 模块结构

导入图问题：循环依赖、过长导入链、层级违规。

| 规则 | 严重程度 |
| ---- | -------- |
| [`circular-import`](circular-import.md) | 按循环大小 |
| [`deep-import-chain`](deep-import-chain.md) | Info |
| [`layer-violation`](layer-violation.md) | 按架构规则 |
| [`low-package-cohesion`](low-package-cohesion.md) | Warning |
| [`single-responsibility`](single-responsibility.md) | Warning / Error |

## 模拟数据

意外发布到生产环境的占位数据。

| 规则 | 严重程度 |
| ---- | -------- |
| [`mock-keyword-in-code`](mock-keyword-in-code.md) | Info / Warning |
| [`mock-domain-in-string`](mock-domain-in-string.md) | Warning |
| [`mock-email-address`](mock-email-address.md) | Warning |
| [`placeholder-phone-number`](placeholder-phone-number.md) | Warning |
| [`placeholder-uuid`](placeholder-uuid.md) | Warning |
| [`placeholder-comment`](placeholder-comment.md) | Info |
| [`repetitive-string-literal`](repetitive-string-literal.md) | Info |
| [`test-credential-in-code`](test-credential-in-code.md) | Warning |

## 在命令行中选择规则

大多数用户通过 `pyscn analyze` 运行所有规则。在 CI 中，可按分析器类别过滤：

```bash
pyscn check --select deadcode          # 仅不可达代码规则
pyscn check --select clones            # 仅重复代码规则
pyscn check --select complexity        # 仅 high-cyclomatic-complexity
pyscn check --select deps              # circular-import + deep-import-chain + layer-violation
pyscn check --select di                # 所有依赖注入规则（可选启用）
pyscn check --select mockdata          # 所有模拟数据规则（可选启用）
pyscn check --select complexity,deadcode,deps   # 组合使用
```

详见 [`pyscn check`](../cli/check.md) 获取完整标志列表。

## 严重程度含义

| 严重程度 | 含义 |
| -------- | --- |
| **Critical** | 几乎总是 bug。建议在合并前修复。 |
| **Error** | 高风险模式。通常应使 CI 失败。 |
| **Warning** | 值得审查。是 `pyscn check` 的默认失败阈值。 |
| **Info** | 仅供参考。仅在 `min_severity = "info"` 或等效设置时显示。 |
| **按阈值** | 严重程度取决于数值阈值（参见规则的选项部分）。 |
