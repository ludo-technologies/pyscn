# HTML 报告

`pyscn analyze` 生成的 HTML 输出规范（未指定 `--json`/`--yaml`/`--csv` 标志时的默认格式）。

## 文件特性

| 属性 | 值 |
| --- | --- |
| 路径 | `.pyscn/reports/analyze_YYYYMMDD_HHMMSS.html` |
| 编码 | UTF-8 |
| 外部资源 | 无（CSS 和 JS 内联） |
| 依赖 | 无（无 CDN、无远程加载字体） |
| 大小 | 通常 50–500 KB |

该文件是自包含的，可以安全地存档、通过电子邮件发送或从任何静态托管服务器提供。

## 文档结构

| 元素 | 内容 |
| --- | --- |
| 顶栏 | pyscn 版本、项目名称和根目录、生成时间戳、文件数、耗时。 |
| 标签栏 | Overview 以及每个已运行领域一个标签页，各带发现计数徽章。 |
| Overview | 健康评分环和等级、一段总评、规模数字、各维度评分卡、优先建议、热点文件、复杂度直方图，以及重复、类、结构的摘要卡。 |
| 详情标签页 | Functions、Duplication、Classes、Architecture。 |
| 页脚 | pyscn 仓库链接和版本字符串。 |

评分卡、摘要卡和标签页仅在对应分析器已运行时显示。Architecture 在依赖、架构或社区分析任一运行时显示；分层规则需要配置 `[architecture]` 层。

## Overview

| 区块 | 内容 |
| --- | --- |
| 总评 | 以圆环绘制的健康评分（0–100）、等级徽章（A–F）、对应等级的标题，以及一句话列出干净的维度和最弱的两三个维度及其关键数字。 |
| Score breakdown | 每个已启用维度（Complexity、Dead code、Duplication、Coupling、Cohesion、Dependencies、Architecture、Communities）一张卡片，包含 0–100 评分、按区间着色的条形和两个辅助数字。卡片链接到对应详情标签页。 |
| Fix first | 优先级最高的五条建议，含严重性、工作量、位置和理由。 |
| Hotspot files | 最多八个模块，按高风险函数数、最大复杂度、死代码、克隆片段数排序。 |
| Complexity distribution | 按配置的风险阈值分箱的函数复杂度直方图，附中位数、最深嵌套和最长函数。 |
| Duplication / Classes / Structure | 链接到详情标签页的紧凑摘要卡。 |

## 详情标签页

| 标签页 | 内容 |
| --- | --- |
| Functions | 复杂度指标条、最复杂函数（前 20）、最长函数、死代码发现（前 20），以及两个折叠的可排序表格：所有模块和目录复杂度汇总。 |
| Duplication | 克隆统计条、克隆组（前 10）及片段与可选代码预览；未形成组时显示克隆对。 |
| Classes | 耦合度（CBO）和内聚度（LCOM4）指标条，以及耦合最高和内聚最低的类（各前 15）。 |
| Architecture | 模块依赖指标、主序列区域、循环依赖、最长链、分层规则违规，以及社区检测和宏观架构图。 |

## JavaScript

内联脚本负责切换标签页（当前标签页会写入 URL 哈希，链接可直接打开指定标签页），并通过一个共享排序器处理模块和目录表格。没有网络请求。

模块汇总使用 `min_complexity`、`report_unchanged`、`min_severity` 展示过滤之前的完整分析总体。目录复杂度使用复杂度过滤之后的报告函数总体，因此其计数和平均值与 Functions 标签页一致。

## CSS

样式从 `service/templates/analyze/report.css` 内联，并使用 CSS 自定义属性。主要令牌：

| 变量 | 语义角色 |
| --- | --- |
| `--good` / `--warn` / `--bad` | 评分区间（75 以上、60–74、低于 60）、风险级别、严重性。 |
| `--accent` | 导航、链接、中性图表条形。 |
| `--ink` / `--ink-2` / `--muted` | 正文、次要和说明文本。 |
| `--surface` / `--page` / `--line` | 卡片、页面背景和细线。 |

暗色模式遵循 `prefers-color-scheme` 媒体查询，根元素上的 `data-theme="light"` 或 `data-theme="dark"` 属性可覆盖它。不渲染切换开关。

## 自动打开行为

当以下条件**全部**满足时，报告会在默认浏览器中打开：

- 格式为 HTML。
- 标准输入为 TTY。
- `SSH_TTY` 和 `SSH_CONNECTION` 环境变量未设置。
- `CI` 环境变量未设置。
- 未传递 `--no-open`。

打开机制：macOS 上使用 `open`，Linux 上使用 `xdg-open`（或 `gnome-open` / `kde-open`），Windows 上使用 `cmd /c start`。

文件 URL 格式：`file:///{absolute-path-to-report}`。

无论是否自动打开，报告路径始终会打印到 stderr。

## 禁止自动打开

```bash
pyscn analyze --no-open .
```

或者在环境中导出 `CI=true`。

## 等级徽章映射

| 等级 | 分数 | 徽章颜色 |
| ----- | ----- | --- |
| A | 90–100 | 绿色（`--good`） |
| B | 75–89  | 绿色（`--good`） |
| C | 60–74  | 琥珀色（`--warn`） |
| D | 45–59  | 红色（`--bad`） |
| F | 0–44   | 红色（`--bad`） |

## 交叉引用

- [健康评分](health-score.md) — 总体评分的计算公式。
- [输出模式](schemas.md) — 机器可读的替代格式。
- [输出格式](index.md) — 所有输出格式和稳定性约定。
