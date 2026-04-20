# 输出格式

pyscn 将分析结果写入输出目录下的文件。所有格式在补丁版本之间共享稳定的字段语义。

## 输出目录

默认值：当前工作目录下的 `.pyscn/reports/`。

可通过 `.pyscn.toml` 配置：

```toml
[output]
directory = "build/reports"
```

## 文件名模式

```
{command}_YYYYMMDD_HHMMSS.{ext}
```

`{command}` 是 `analyze`（唯一写入报告的 pyscn 命令）。时间戳为本地时间。已有文件不会被覆盖。

## 支持的格式

| 格式 | 扩展名 | 标志          | 规范                             |
| ------ | --------- | ------------- | -------------------------------- |
| text   | —         | （终端）      | 人类可读，不稳定                 |
| json   | `.json`   | `--json`      | [schemas.md](schemas.md)         |
| yaml   | `.yaml`   | `--yaml`      | [schemas.md](schemas.md)         |
| csv    | `.csv`    | `--csv`       | [schemas.md](schemas.md)         |
| html   | `.html`   | `--html`（默认） | [html-report.md](html-report.md) |

`text` 格式用于终端显示，没有稳定性保证；其布局可能在任何版本之间发生变化。

## 稳定性约定

在同一主版本内的补丁和次要版本之间：

- **稳定**：`json`、`yaml` 和 `csv` 中的字段名称、类型和语义。
- **可能变更**：数组元素的顺序、新字段的添加、新顶级部分的添加、`text` 和 `html` 的外观变化。
- **破坏性变更**：仅限于主版本升级（字段的删除或重命名、字段类型的更改）。

第三方集成应忽略未知字段，且不应依赖对象内的字段顺序。
