# クイックスタート

## 解析を実行する

```bash
uvx pyscn@latest analyze .
```

`pyscn` がすでにインストールされている場合（`uv tool install pyscn`、`pipx install pyscn`、`pip install pyscn` のいずれか）、`uvx pyscn@latest` プレフィックスは不要です:

```bash
pyscn analyze .
```

HTML レポートが `.pyscn/reports/analyze_YYYYMMDD_HHMMSS.html` に出力され、デフォルトブラウザで自動的に開かれます。

## 出力フォーマットの選択

```bash
pyscn analyze --json .
pyscn analyze --yaml .
pyscn analyze --csv .
pyscn analyze --no-open .       # suppress browser open
```

## 特定の解析だけを実行する

```bash
pyscn analyze --select complexity .
pyscn analyze --select complexity,deadcode .
pyscn analyze --skip-clones .
```

すべてのフラグについては [`analyze`](../cli/analyze.md) を参照してください。

## CI 品質ゲート

```bash
pyscn check .                              # exit 0 pass, 1 fail
pyscn check --max-complexity 15 src/
pyscn check --select complexity,deadcode,deps src/
```

詳しくは [`check`](../cli/check.md) と [CI/CD Integration](../integrations/ci-cd.md) をご覧ください。

## 設定ファイルの生成

```bash
pyscn init
```

すべてのオプションがコメント付きで記載された `.pyscn.toml` を生成します。詳しくは [Configuration Reference](../configuration/reference.md) をご覧ください。
