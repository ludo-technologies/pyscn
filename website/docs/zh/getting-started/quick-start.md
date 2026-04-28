# 快速开始

## 运行分析

```bash
uvx pyscn@latest analyze .
```

如果已安装 `pyscn`（通过 `uv tool install pyscn`、`pipx install pyscn` 或 `pip install pyscn`），可省略 `uvx pyscn@latest` 前缀：

```bash
pyscn analyze .
```

分析完成后会将 HTML 报告写入 `.pyscn/reports/analyze_YYYYMMDD_HHMMSS.html` 并在默认浏览器中打开。

## 选择输出格式

```bash
pyscn analyze --json .
pyscn analyze --yaml .
pyscn analyze --csv .
pyscn analyze --no-open .       # 不自动打开浏览器
```

## 运行特定分析器

```bash
pyscn analyze --select complexity .
pyscn analyze --select complexity,deadcode .
pyscn analyze --skip-clones .
```

所有选项详见 [`analyze`](../cli/analyze.md)。

## CI 质量门禁

```bash
pyscn check .                              # 退出码 0 通过，1 失败
pyscn check --max-complexity 15 src/
pyscn check --select complexity,deadcode,deps src/
```

详见 [`check`](../cli/check.md) 和 [CI/CD 集成](../integrations/ci-cd.md)。

## 生成配置文件

```bash
pyscn init
```

生成 `.pyscn.toml`，其中包含所有选项的注释说明。详见[配置参考](../configuration/reference.md)。
