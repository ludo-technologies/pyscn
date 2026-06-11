---
title: 如何在 Python 中发现重复代码
description: 如何在 Python 项目中发现重复代码的实用指南。介绍代码克隆的四种类型（Type 1-4）、可用工具，以及如何在 CI 中自动化检测。
---

# 如何在 Python 中发现重复代码

重复代码会带来实际问题。在一处副本中修复的 bug，往往在其他副本中继续存在。每次改动都要在多处同步修改。代码审查者也会浪费时间反复阅读已经看过的逻辑。如今 AI 助手大量参与代码编写，近乎相同的代码块在代码库中出现的速度，远比人工复制粘贴快得多。

本指南将说明什么样的代码算作"重复"，哪些工具可以在 Python 中检测重复代码，以及如何在本地和 CI 中运行检测。

## 快速上手

如需立即扫描项目，运行：

```bash
uvx pyscn@latest analyze --select clones .
```

这会在不安装任何依赖的情况下运行 [pyscn](https://github.com/ludo-technologies/pyscn) 的克隆检测。命令完成后会打开一份 HTML 报告，列出所有重复代码组，并附有相似度评分和文件位置。

## 什么算"重复"？四种克隆类型

大多数开发者对重复代码的印象是字面意义上的复制粘贴。在研究领域，重复代码被称为*代码克隆*，分为四种类型。这一区别很重要，因为大多数工具只能捕获前一两种。

下面用同一个函数的四个变体来说明。

**Type-1：完全相同的代码。** 复制粘贴，仅空白和注释有所不同：

```python
def calculate_order_total(items, discount_rate):
    subtotal = 0.0
    for item in items:
        price = item["price"]
        quantity = item["quantity"]
        if quantity <= 0:
            continue
        subtotal += price * quantity
    if discount_rate > 0:
        subtotal = subtotal * (1 - discount_rate)
    tax = subtotal * 0.1
    total = subtotal + tax
    return round(total, 2)
```

如果这段代码被粘贴到另一个文件（也许加了一条注释），就是 Type-1 克隆。这是最容易检测、也最容易修复的类型：提取为共享函数即可。（[规则：duplicate-code-identical](../rules/duplicate-code-identical.md)）

**Type-2：重命名标识符。** 结构完全相同，只有名称发生了变化：

```python
def compute_cart_amount(products, rebate):
    amount = 0.0
    for product in products:
        cost = product["price"]
        count = product["quantity"]
        if count <= 0:
            continue
        amount += cost * count
    if rebate > 0:
        amount = amount * (1 - rebate)
    levy = amount * 0.1
    result = amount + levy
    return round(result, 2)
```

基于行的工具会忽略这种情况，因为没有任何两行在文本上完全相同。但如果将语法树的名称归一化后再比较，这两个函数的结构完全一致。（[规则：duplicate-code-renamed](../rules/duplicate-code-renamed.md)）

**Type-3：修改后的副本。** 有人复制了该函数，随后增加或删除了几条语句：

```python
def calculate_quote_total(items, discount_rate, shipping=0.0):
    subtotal = 0.0
    for item in items:
        price = item["price"]
        quantity = item["quantity"]
        if quantity <= 0:
            continue
        subtotal += price * quantity
    if discount_rate > 0:
        subtotal = subtotal * (1 - discount_rate)
    subtotal += shipping        # <- 新增
    tax = subtotal * 0.1
    total = subtotal + tax
    return round(total, 2)
```

这是实际代码库中最常见的克隆类型。有人复制了一个函数，为新场景做了调整，然后继续开发。检测这类克隆需要度量两棵语法树之间的距离（树编辑距离），而不只是判断是否匹配。（[规则：duplicate-code-modified](../rules/duplicate-code-modified.md)）

**Type-4：行为相同，实现不同。** 代码从头重写，但计算结果相同：

```python
def total_for_order(items, discount_rate):
    valid_items = []
    for item in items:
        if item["quantity"] > 0:
            valid_items.append(item)
    subtotal = sum(
        item["price"] * item["quantity"]
        for item in valid_items
    )
    if discount_rate > 0:
        subtotal = subtotal * (1 - discount_rate)
    total_with_tax = subtotal * 1.1
    return round(total_with_tax, 2)
```

文本匹配或树匹配都无法将这段代码与原始版本关联起来，但通过比较控制流结构可以发现。（[规则：duplicate-code-semantic](../rules/duplicate-code-semantic.md)）

## 检测 Python 重复代码的工具

以下是主要工具及各自的检测能力：

| 工具 | 检测范围 | 说明 |
| --- | --- | --- |
| [pylint](https://pylint.readthedocs.io/)（`R0801`） | Type-1 | 基于行的相似度检查，随 pylint 一同提供。能捕获复制粘贴，重命名后失效。 |
| [jscpd](https://github.com/kucherenko/jscpd) | Type-1，部分 Type-2 | 基于 token，支持 150 多种语言。适合需要一个检测器覆盖多语言仓库的场景。 |
| [SonarQube](https://www.sonarsource.com/products/sonarqube/) | Type-1，部分 Type-2 | 完整平台，含仪表盘和历史记录。配置和托管成本较高。 |
| [PMD CPD](https://pmd.github.io/pmd/pmd_userdocs_cpd.html) | Type-1，Type-2 | 经典的复制粘贴检测器，需要 JVM。 |
| [pyscn](https://github.com/ludo-technologies/pyscn) | Type-1 至 Type-4 | Python 专用。Type 1-2 使用 AST 哈希，Type-3 使用树编辑距离（APTED），Type-4 使用控制流比较。 |

一个常见的误解：**[ruff](https://docs.astral.sh/ruff/) 不检测重复代码。** Ruff 是 lint 工具和格式化工具，负责检查单行和语句的写法，这与跨文件比较函数是两件完全不同的事。两类工具相辅相成，而非互相竞争。

## 演练：使用 pyscn 检测克隆

将上面四个变体分散到两个文件 `orders.py` 和 `invoices.py` 中，并将原始版本粘贴到两个文件里，然后运行：

```bash
uvx pyscn@latest analyze --select clones .
```

pyscn 会解析所有 Python 文件，提取代码片段，并两两比较。对于大型代码库，它使用 [LSH](https://en.wikipedia.org/wiki/Locality-sensitive_hashing) 加速，分析速度超过 100,000 行/秒。终端会显示摘要：

```text
📊 Analysis Summary:
Health Score: 80/100 (Grade: B)

📈 Detailed Scores:
  Duplication:      0/100 ❌  (10.0% duplication, 1 groups)
```

HTML 报告会将五个片段归入同一个克隆组，并对每对代码进行分类和评分：

| 代码对 | 分类 | 相似度 |
| --- | --- | --- |
| 两个文件中的完全相同副本 | Type-1 | 1.00 |
| 原始版本与修改副本 | Type-2 | 0.85 |
| 原始版本与重写版本 | Type-4 | 0.94 |

注意最后一行。重写后的 `total_for_order` 是任何基于文本的工具都无法与原始版本关联的变体，pyscn 通过控制流结构检测到了 0.94 的相似度。

### 调整阈值

`--clone-threshold` 参数（默认值 `0.65`）设置报告一对代码所需的最低相似度：

```bash
pyscn analyze --select clones --clone-threshold 0.8 .   # 更严格：匹配数量更少、相似度更高
```

如需固化设置，创建 `.pyscn.toml` 文件（或在 `pyproject.toml` 中使用 `[tool.pyscn]`）：

```toml
[clones]
similarity_threshold = 0.8
min_lines = 15        # ignore fragments smaller than this
```

默认会跳过非常短的函数（`min_lines`）。低于某个大小，相似度就失去了意义，因为每个两行的 getter 看起来都差不多。所有选项（包括单独开关各克隆类型）详见[配置参考](../configuration/reference.md#clones)。

## 在 CI 中自动化检测

`pyscn check` 是 `analyze` 的 CI 版本，不生成报告，只输出通过/失败的退出码：

```bash
pyscn check --select clones .
```

作为 GitHub Actions 步骤：

```yaml
- uses: actions/setup-python@v5
  with:
    python-version: "3.12"
- run: pipx run pyscn check --select clones .
```

当新增重复代码超出阈值时，任务失败。这正是关键所在：需要手动触发的检测，迟早会被遗忘而停止执行。完整工作流详见 [CI/CD 集成](../integrations/ci-cd.md)，如需在 Pull Request 上自动发布审查评论，请参阅 [Pyscn Bot](https://github.com/marketplace/pyscn-bot)。

## 如何处理检测结果

并非所有克隆都需要消除。建议按以下优先级处理：

1. **生产代码中的 Type-1 和 Type-2 克隆。** 提取为共享函数。由于副本几乎完全相同，修复方式机械化，风险极低。
2. **Type-3 克隆。** 仔细查看各副本之间的差异。如果差异体现在数据上，提取一个接受参数的函数；如果差异体现在行为上，这些副本可能是有意分化的。有时两处调用确实需要独立演进，强行合并反而会耦合本该独立的模块。
3. **Type-4 克隆。** 将其视为信号，而非必须立即处理的任务。同一逻辑的两套独立实现，往往意味着两位开发者互不知情。选择其中一套保留，或记录说明两者共存的原因。
4. **测试代码中的克隆。** 对这类情况可以宽松一些。测试代码注重显式表达，而非 DRY 原则，适当的重复有助于每个测试用例保持独立可读性。

实践中，设置一个较严格的阈值以保持报告简短，修复排名最靠前的克隆组，再重新运行，效果往往好于一次性清理 40 组克隆的大规模重构。

## 常见问题

**ruff 能检测重复代码吗？**
不能。Ruff 是 lint 工具和格式化工具，没有克隆检测规则。发现重复代码需要跨文件比较代码片段，这超出了 linter 的职责范围。请用 ruff 做风格和正确性检查，用克隆检测工具处理重复问题，两者配合使用效果更佳。

**多少重复是可以接受的？**
没有统一的标准。粗略来说，维护良好的代码库重复行占比通常低于 5%，超过 15% 一般意味着存在系统性的复制粘贴开发习惯。趋势比数字本身更重要。每个版本发布后重复率持续上升，才是真正需要警惕的信号。

**能跨多个仓库检测重复吗？**
可以。将分析器指向包含多个仓库检出目录的父目录即可：`pyscn analyze --select clones repo-a/ repo-b/`。所有在扫描范围内的片段都会两两比较，跨仓库克隆与普通克隆一样会出现在报告中。

**为什么我的重复代码片段没有被报告？**
最可能的原因是片段低于最小大小限制（配置中的 `min_lines` / `min_nodes`）。检测器故意跳过极短的片段。在五行的情况下，代码库中的大半代码都会彼此相似。如需比较短片段，请在 `.pyscn.toml` 中降低限制值。

---

*延伸阅读：查看[重复代码规则目录](../rules/index.md)了解各克隆类型的评分方式，或查看[健康评分文档](../output/health-score.md)了解重复代码如何影响项目评级。*
