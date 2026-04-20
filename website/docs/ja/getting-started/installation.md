# インストール

## 動作要件

- Python 3.8〜3.13（ランチャーのみ使用。pyscn 自体は Python ランタイムに依存しません）
- Linux / macOS / Windows、x86_64 または arm64

## インストール方法

| 方法 | コマンド | 備考 |
| --- | --- | --- |
| **uvx**（推奨） | `uvx pyscn@latest <command>` | インストール不要で実行。初回呼び出し後にキャッシュされます。 |
| uv tool | `uv tool install pyscn` | 永続インストール。プロジェクトの依存関係から隔離されます。 |
| pipx | `pipx install pyscn` | 永続インストール。プロジェクトの依存関係から隔離されます。 |
| pip | `pip install pyscn` | 現在の環境にインストールされます。 |
| Go | `go install github.com/ludo-technologies/pyscn/cmd/pyscn@latest` | `pyscn-mcp` はインストールされません。 |

`uvx` は単発の利用に最も手軽で、CI でもうまく動作します。繰り返しローカルで使う場合は `uv tool install` や `pipx` を使えば、プロジェクトの依存関係を汚しません。

ビルド済みバイナリはすべての [GitHub release](https://github.com/ludo-technologies/pyscn/releases) に添付されています。

## 確認

```bash
pyscn version
pyscn version --short    # just the version number
```

## アップグレード

```bash
uv tool upgrade pyscn        # if installed with uv tool
pipx upgrade pyscn           # if installed with pipx
pip install --upgrade pyscn  # if installed with pip
```

`uvx pyscn@latest` は常に最新バージョンに解決されるため、アップグレード操作は不要です。

## アンインストール

```bash
uv tool uninstall pyscn
pipx uninstall pyscn
pip uninstall pyscn
```
