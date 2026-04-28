# Python パッケージング

pyscn はネイティブ Go バイナリを含む wheel として PyPI で配布されています。Python レイヤーは stdlib ランチャーであり、Python のランタイム依存関係はありません。

## どのインストーラーを使うべきか

| ツール | 適したケース | 備考 |
| --- | --- | --- |
| `uvx`（推奨） | ワンショット実行、CI | インストール不要で実行。初回呼び出し後にキャッシュされます。 |
| `uv tool install` | 永続的なツール管理 | 高速、分離環境。 |
| `pipx` | 永続的な CLI インストール | プロジェクトの依存関係から分離。 |
| `pip` | venv へのインストール | 分離なし。 |

CI: `uvx pyscn@latest check .`。ローカル開発: `uv tool install pyscn` または `pipx install pyscn`。

## プラットフォームサポート

| OS | アーキテクチャ |
| --- | --- |
| Linux | x86_64, arm64 |
| macOS | x86_64, arm64 |
| Windows | x86_64, arm64 |

Python 3.8〜3.13。

## パッケージ

| パッケージ | 内容 | インストールするタイミング |
| --- | --- | --- |
| `pyscn` | CLI + MCP サーバー | CLI を使いたい場合。 |
| `pyscn-mcp` | MCP サーバーのみ | MCP サーバーだけが必要な場合。 |

## バージョニング

[PEP 440](https://peps.python.org/pep-0440/) に準拠し、Git タグと一致します:

- `0.1.0` — 安定版
- `0.2.0.dev1` — 開発版
- `0.2.0b1` — ベータ版

再現性のためにバージョンを固定:

```bash
pip install pyscn==0.2.0
```

## コンテナ

```dockerfile
FROM python:3.12-slim
RUN pip install --no-cache-dir pyscn
ENTRYPOINT ["pyscn"]
```

## wheel の内容

```
pyscn-0.2.0-py3-none-manylinux_2_17_x86_64.whl
├── pyscn/
│   ├── __init__.py
│   ├── __main__.py        # CLI launcher
│   ├── mcp_main.py        # MCP launcher
│   └── bin/
│       └── pyscn          # Go binary
```

ランチャーは OS とアーキテクチャを検出し、一致するバイナリを `exec` します。

## リリース

`v` で始まる Git タグでバージョンが公開されます:

```bash
git tag -a v0.2.0 -m "Release v0.2.0"
git push origin v0.2.0
```

GitHub Actions がクロスコンパイルし、プラットフォーム固有の wheel をパッケージ化し、`twine check` を実行し、OS x Python のマトリクスでスモークテストを行い、PyPI に公開し、GitHub リリースを作成します。[リリースページ](https://github.com/ludo-technologies/pyscn/releases)を参照してください。

## PyPI 以外の代替手段

- `go install github.com/ludo-technologies/pyscn/cmd/pyscn@latest`（Go 1.22 以上。`pyscn-mcp` はインストールされません）。
- GitHub Releases からのバイナリダウンロード。

## 関連項目

- [インストール](../getting-started/installation.md)
- [CI/CD 連携](ci-cd.md)
