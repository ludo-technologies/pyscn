# 設定例

一般的なシナリオ向けのコピー&ペースト用テンプレートです。

## 最小限のオーバーライド

いくつかの厳格な閾値のみを指定し、それ以外はデフォルトのままにします。

```toml
# .pyscn.toml
[complexity]
max_complexity = 15

[dead_code]
min_severity = "critical"
```

## 厳格な CI ゲート

品質低下があればビルドを失敗させます。`pyscn check` と組み合わせて使用します。

```toml
[complexity]
max_complexity = 10

[dead_code]
min_severity = "warning"
detect_after_return = true
detect_after_raise = true
detect_unreachable_branches = true

[clones]
# ほぼ同一のコードのみを検出
similarity_threshold = 0.90
min_lines = 15

[cbo]
medium_threshold = 7

[dependencies]
enabled = true
detect_cycles = true
```

実行:

```bash
pyscn check --select complexity,deadcode,deps --max-cycles 0 src/
```

## レガシーコードベースの猶予期間

既存プロジェクトに pyscn を導入する場合、大量の失敗を避けつつシグナルを得たいときに使用します。

```toml
[complexity]
max_complexity = 25    # 既存の複雑度を許容

[dead_code]
min_severity = "critical"   # 最も深刻なもののみ

[clones]
min_lines = 20              # 長い重複のみ
similarity_threshold = 0.90

[analysis]
exclude_patterns = [
  "legacy/**",     # 古いコードを隔離
  "**/_archive/*",
  "generated/**",
]
```

時間をかけて閾値を段階的に厳しくしていきます。

## 大規模コードベース（10k+ ファイル）

スループットを最適化します。LSH は自動で有効になりますが、並列度を上げます。

```toml
[clones]
lsh_enabled = true
max_goroutines = 16
max_memory_mb = 2048
batch_size = 500
timeout_seconds = 600
min_lines = 15           # より少なく、より意味のある候補

[analysis]
exclude_patterns = [
  "**/test_*.py", "**/*_test.py",
  "**/migrations/**",
  "**/__generated__/**",
  "**/node_modules/**",
  ".venv/**", "venv/**",
]
```

## クリーンアーキテクチャのバリデーション

レイヤードアーキテクチャを強制します: presentation → application → domain、infrastructure はエッジに配置。

```toml
[architecture]
enabled = true
strict_mode = true
fail_on_violations = true

[[architecture.layers]]
name = "presentation"
packages = ["api", "routers", "handlers", "views"]

[[architecture.layers]]
name = "application"
packages = ["services", "usecases"]

[[architecture.layers]]
name = "domain"
packages = ["models", "entities", "core"]

[[architecture.layers]]
name = "infrastructure"
packages = ["repositories", "db", "adapters", "clients"]

[[architecture.rules]]
from = "presentation"
allow = ["application", "domain"]
deny = ["infrastructure"]

[[architecture.rules]]
from = "application"
allow = ["domain", "infrastructure"]
deny = ["presentation"]

[[architecture.rules]]
from = "domain"
deny = ["presentation", "application", "infrastructure"]
```

## データ中心の ML / リサーチコードベース

ノートブックから変換されたモジュールでは高い複雑度が想定されます。重複とデッドコードに焦点を当てます。

```toml
[complexity]
max_complexity = 30    # データパイプラインは分岐が多い

[dead_code]
min_severity = "critical"

[clones]
# リサーチコードではほぼ同一の実験バリアントが多い
# 大量の検出結果を避けるため閾値を上げる
min_lines = 20
similarity_threshold = 0.85

[analysis]
exclude_patterns = [
  "notebooks/**",
  "experiments/**/*.ipynb",
]
```

## `pyproject.toml` との共存

既に `pyproject.toml` がある場合、新しいファイルを作成する代わりにそこに pyscn の設定を記述できます:

```toml
# pyproject.toml
[project]
name = "my-package"
# ... その他のプロジェクトメタデータ

[tool.pyscn.complexity]
max_complexity = 15

[tool.pyscn.dead_code]
min_severity = "critical"

[tool.pyscn.clones]
similarity_threshold = 0.85
```

!!! note
    両方のファイルが存在する場合、`.pyscn.toml` が `pyproject.toml` より優先されます。混乱を避けるためどちらか一方を選択してください。

## 関連項目

- [設定ファイル形式](format.md)
- [設定リファレンス](reference.md)
