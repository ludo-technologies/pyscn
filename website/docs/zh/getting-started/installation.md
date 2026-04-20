# 安装

## 系统要求

- Python 3.8-3.13（仅用于启动器；pyscn 无 Python 运行时依赖）
- Linux / macOS / Windows，x86_64 或 arm64

## 安装方式

| 方式 | 命令 | 说明 |
| --- | --- | --- |
| **uvx**（推荐） | `uvx pyscn@latest <command>` | 免安装直接运行；首次调用后会缓存。 |
| uv tool | `uv tool install pyscn` | 持久安装，与项目依赖隔离。 |
| pipx | `pipx install pyscn` | 持久安装，与项目依赖隔离。 |
| pip | `pip install pyscn` | 安装到当前环境。 |
| Go | `go install github.com/ludo-technologies/pyscn/cmd/pyscn@latest` | 不会安装 `pyscn-mcp`。 |

`uvx` 是一次性使用的最快方式，也适用于 CI 环境。如需在本地反复使用且不影响项目依赖，推荐 `uv tool install` 或 `pipx`。

每个 [GitHub release](https://github.com/ludo-technologies/pyscn/releases) 都附带预编译二进制文件。

## 验证安装

```bash
pyscn version
pyscn version --short    # 仅输出版本号
```

## 升级

```bash
uv tool upgrade pyscn        # 通过 uv tool 安装的情况
pipx upgrade pyscn           # 通过 pipx 安装的情况
pip install --upgrade pyscn  # 通过 pip 安装的情况
```

`uvx pyscn@latest` 始终解析为最新版本，因此无需升级操作。

## 卸载

```bash
uv tool uninstall pyscn
pipx uninstall pyscn
pip uninstall pyscn
```
