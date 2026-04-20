# Python 打包

pyscn 在 PyPI 上以包含原生 Go 二进制文件的 wheel 形式分发。Python 层是一个标准库启动器；没有 Python 运行时依赖。

## 选择安装工具

| 工具 | 适用场景 | 说明 |
| --- | --- | --- |
| `uvx`（推荐） | 一次性运行、CI | 无需安装即可运行；首次调用后自动缓存。 |
| `uv tool install` | 持久化工具管理 | 快速、隔离。 |
| `pipx` | 持久化 CLI 安装 | 与项目依赖隔离。 |
| `pip` | 安装到 venv 中 | 无隔离。 |

CI：`uvx pyscn@latest check .`。本地开发：`uv tool install pyscn` 或 `pipx install pyscn`。

## 平台支持

| 操作系统 | 架构 |
| --- | --- |
| Linux | x86_64, arm64 |
| macOS | x86_64, arm64 |
| Windows | x86_64, arm64 |

Python 3.8–3.13。

## 包

| 包 | 包含 | 安装场景 |
| --- | --- | --- |
| `pyscn` | CLI + MCP 服务器 | 需要 CLI 时。 |
| `pyscn-mcp` | 仅 MCP 服务器 | 仅需要 MCP 服务器时。 |

## 版本管理

遵循 [PEP 440](https://peps.python.org/pep-0440/)，与 Git 标签匹配：

- `0.1.0` — 稳定版
- `0.2.0.dev1` — 开发版
- `0.2.0b1` — 测试版

为确保可重现性，请固定版本：

```bash
pip install pyscn==0.2.0
```

## 容器

```dockerfile
FROM python:3.12-slim
RUN pip install --no-cache-dir pyscn
ENTRYPOINT ["pyscn"]
```

## Wheel 内容

```
pyscn-0.2.0-py3-none-manylinux_2_17_x86_64.whl
├── pyscn/
│   ├── __init__.py
│   ├── __main__.py        # CLI launcher
│   ├── mcp_main.py        # MCP launcher
│   └── bin/
│       └── pyscn          # Go binary
```

启动器检测操作系统和架构，然后 `exec` 匹配的二进制文件。

## 发布

版本通过以 `v` 开头的 Git 标签发布：

```bash
git tag -a v0.2.0 -m "Release v0.2.0"
git push origin v0.2.0
```

GitHub Actions 进行交叉编译、打包平台特定 wheel、运行 `twine check`、在多个操作系统和 Python 版本组合上进行冒烟测试、发布到 PyPI 并创建 GitHub Release。详见 [Releases 页面](https://github.com/ludo-technologies/pyscn/releases)。

## PyPI 之外的替代方案

- `go install github.com/ludo-technologies/pyscn/cmd/pyscn@latest`（需要 Go 1.22+；不安装 `pyscn-mcp`）。
- 从 GitHub Releases 下载二进制文件。

## 另请参阅

- [安装](../getting-started/installation.md)
- [CI/CD 集成](ci-cd.md)
