# ルールカタログ

pyscn は7つのカテゴリにわたる33のルールを提供しています。各ルールには、検出内容、なぜ問題なのか、悪い例、修正方法を説明するページがあります。

ルール名をクリックすると、そのページが開きます。

## 到達不能コード

制御フローグラフの到達可能性分析により検出される、実行されることのないデッドコードです。

| ルール | 重大度 |
| ---- | -------- |
| [`unreachable-after-return`](unreachable-after-return.md) | Critical |
| [`unreachable-after-raise`](unreachable-after-raise.md) | Critical |
| [`unreachable-after-break`](unreachable-after-break.md) | Critical |
| [`unreachable-after-continue`](unreachable-after-continue.md) | Critical |
| [`unreachable-after-infinite-loop`](unreachable-after-infinite-loop.md) | Warning |
| [`unreachable-branch`](unreachable-branch.md) | Warning |

## 重複コード

プロジェクト内のコピペまたはほぼコピペのコード断片です。

| ルール | 重大度 |
| ---- | -------- |
| [`duplicate-code-identical`](duplicate-code-identical.md) | Warning |
| [`duplicate-code-renamed`](duplicate-code-renamed.md) | Warning |
| [`duplicate-code-modified`](duplicate-code-modified.md) | Info (オプトイン) |
| [`duplicate-code-semantic`](duplicate-code-semantic.md) | Warning |

## 複雑度

テストや推論が困難なほど分岐の多い関数です。

| ルール | 重大度 |
| ---- | -------- |
| [`high-cyclomatic-complexity`](high-cyclomatic-complexity.md) | By threshold |

## クラス設計

依存先が多すぎる、または無関係な仕事を多く抱えすぎているクラスです。

| ルール | 重大度 |
| ---- | -------- |
| [`high-class-coupling`](high-class-coupling.md) | By threshold |
| [`low-class-cohesion`](low-class-cohesion.md) | By threshold |

## 依存性注入

テスト容易性を損なうコンストラクタやコラボレータのパターンです。

| ルール | 重大度 |
| ---- | -------- |
| [`too-many-constructor-parameters`](too-many-constructor-parameters.md) | Warning |
| [`global-state-dependency`](global-state-dependency.md) | Error |
| [`module-variable-dependency`](module-variable-dependency.md) | Warning |
| [`singleton-pattern-dependency`](singleton-pattern-dependency.md) | Warning |
| [`concrete-type-hint-dependency`](concrete-type-hint-dependency.md) | Info |
| [`concrete-instantiation-dependency`](concrete-instantiation-dependency.md) | Warning |
| [`service-locator-pattern`](service-locator-pattern.md) | Warning |

## モジュール構造

インポートグラフの問題：循環、長いチェーン、レイヤー違反です。

| ルール | 重大度 |
| ---- | -------- |
| [`circular-import`](circular-import.md) | By cycle size |
| [`deep-import-chain`](deep-import-chain.md) | Info |
| [`layer-violation`](layer-violation.md) | By architecture rule |
| [`low-package-cohesion`](low-package-cohesion.md) | Warning |
| [`single-responsibility`](single-responsibility.md) | Warning / Error |

## モックデータ

本番環境に誤って含まれたプレースホルダーデータです。

| ルール | 重大度 |
| ---- | -------- |
| [`mock-keyword-in-code`](mock-keyword-in-code.md) | Info / Warning |
| [`mock-domain-in-string`](mock-domain-in-string.md) | Warning |
| [`mock-email-address`](mock-email-address.md) | Warning |
| [`placeholder-phone-number`](placeholder-phone-number.md) | Warning |
| [`placeholder-uuid`](placeholder-uuid.md) | Warning |
| [`placeholder-comment`](placeholder-comment.md) | Info |
| [`repetitive-string-literal`](repetitive-string-literal.md) | Info |
| [`test-credential-in-code`](test-credential-in-code.md) | Warning |

## コマンドラインでのルール選択

ほとんどのユーザーは `pyscn analyze` ですべてのルールを実行します。CI では、アナライザカテゴリでフィルタリングできます：

```bash
pyscn check --select deadcode          # only unreachable-code rules
pyscn check --select clones            # only duplicate-code rules
pyscn check --select complexity        # only high-cyclomatic-complexity
pyscn check --select deps              # circular-import + deep-import-chain + layer-violation
pyscn check --select di                # all dependency-injection rules (opt-in)
pyscn check --select mockdata          # all mock-data rules (opt-in)
pyscn check --select complexity,deadcode,deps   # combine
```

詳細は [`pyscn check`](../cli/check.md) のフラグ一覧をご覧ください。

## 重大度の意味

| 重大度 | 意図 |
| -------- | --- |
| **Critical** | ほぼ確実にバグです。マージ前に修正することを推奨します。 |
| **Error** | 高リスクなパターンです。通常、CI を失敗させるべきです。 |
| **Warning** | レビューする価値があります。`pyscn check` のデフォルトの失敗閾値です。 |
| **Info** | 情報提供のみです。`min_severity = "info"` または同等の設定時にのみ表示されます。 |
| **By threshold** | 重大度は数値の閾値に依存します（各ルールのオプションを参照）。 |
