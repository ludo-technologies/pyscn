# 配置

pyscn 从 `.pyscn.toml` 或 `pyproject.toml` 中的 `[tool.pyscn]` 读取配置。CLI 标志优先于配置文件；配置文件优先于内置默认值。

- **[配置文件格式](format.md)** — 发现规则、文件位置、优先级顺序。
- **[参考手册](reference.md)** — 所有配置键、类型、默认值及作用。
- **[示例](examples.md)** — 严格 CI、大型代码库、最小覆盖配置。

使用以下命令生成带注释的初始配置文件：

```bash
pyscn init
```
