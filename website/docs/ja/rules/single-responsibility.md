# single-responsibility

**カテゴリ**: モジュール構造  
**重大度**: Warning（ファンイン/ファンアウトが極端に大きいハブモジュールでは Error）  
**トリガー**: `pyscn analyze`, `pyscn check --select deps`

## 検出内容

`architecture.max_responsibilities`（デフォルト `3`）を超える数の異なる依存関心事を抱えているモジュール、またはプロジェクト平均と比較してファンイン/ファンアウトが両方とも高いハブモジュールを検出します。

「関心事」は隣接モジュールの名前から推定されます。当該モジュールがインポートする先・される先のそれぞれについて、現在のモジュールのパスと一致しない最初のセグメントを取り出し、汎用的すぎる接尾辞（`base`, `common`, `helpers`, `node`, `shared`, `util`, `utils`）を除外します。残ったセグメントを重複排除した個数が、このモジュールに割り当てられた関心事の数になります。

次のいずれかに該当するとレポートされます:

- 異なる関心事の数が `max_responsibilities` を超える。
- ファンイン（インポートしてくるモジュール数）とファンアウト（インポート先モジュール数）の両方がプロジェクトの「平均 + 標準偏差」を上回り、かつ関心事を 2 つ以上抱えている。

## なぜ問題なのか

単一責任原則 (SRP) の本質は「変更の軸」です。複数の無関係な依存クラスタにまたがるモジュールは、変更理由を複数持ちます:

- **編集が波及する。** 一つの関心事を直すために、同じモジュール境界を共有する他の関心事まで読み直し・テストし直しが必要になります。
- **インポートが嘘をつく。** `from myapp.core import X` は読み手に何の情報も与えません。`core` は複数の仕事を抱えているからです。
- **ハブはボトルネック化する。** 誰からもインポートされ、誰のことでもインポートしているモジュールは、変更・レビュー・マージの単一争点になります。
- **必要な切れ目を隠している。** 2 つの関心事が同じファイルに集まり続けるなら、両者の関係を名前付ける新しいモジュールを作るのが本来の解です。

## 例

```
myapp/core.py
```

```python
# myapp/core.py
from myapp.routers import user_router, order_router
from myapp.services import billing_service, notification_service
from myapp.repositories import user_repo, order_repo
from myapp.telemetry import metrics, tracing

# ...すべてを束ねるグルーコード...
```

`core` は `routers` / `services` / `repositories` / `telemetry` の 4 関心事を混在させ、さらに routers と services の双方からインポートされています。ファンインもファンアウトも高いため、pyscn は過負荷モジュールとして検出します。

## 修正例

すでに存在している関心事に沿ってモジュールを分割します。各モジュールは「変更の軸」を 1 つだけ名乗るべきです。

```
myapp/wiring/web.py          # ルーター層の配線
myapp/wiring/services.py     # サービス層の配線
myapp/wiring/persistence.py  # リポジトリ層の配線
myapp/wiring/observability.py
```

正当なコンポジションルートとして残すなら、責務を狭めましょう。配線だけを行い、業務ルールの実装・型の定義・テレメトリの所有まで一手に引き受けてはいけません。

## オプション

| オプション | デフォルト | 説明 |
| --- | --- | --- |
| [`architecture.validate_responsibility`](../configuration/reference.md#architecture) | `true` | `false` に設定するとこのルールを無効にします。 |
| [`architecture.max_responsibilities`](../configuration/reference.md#architecture) | `3` | この値を超える関心事を抱えるモジュールが検出されます。 |
| [`architecture.enabled`](../configuration/reference.md#architecture) | `true` | アーキテクチャ解析のマスタースイッチ。 |
| [`architecture.fail_on_violations`](../configuration/reference.md#architecture) | `false` | 違反時にゼロ以外の終了コードを返します。 |

## 参照

- 責務推定と重大度ルール: `service/responsibility_analysis.go`。
- Martin, R. C. *アジャイルソフトウェア開発の奥義*（第 8 章 — SRP）。
- [ルールカタログ](index.md) · [low-package-cohesion](low-package-cohesion.md) · [layer-violation](layer-violation.md)
