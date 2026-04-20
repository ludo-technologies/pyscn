# low-package-cohesion

**カテゴリ**: モジュール構造  
**重大度**: Warning  
**トリガー**: `pyscn analyze`, `pyscn check --select deps`

## 検出内容

パッケージの内部凝集度スコアが `architecture.min_cohesion`（デフォルト `0.5`）を下回る場合に検出します。凝集度は、サブモジュール間で可能なパッケージ内インポート数に対する実際のインポート数の比率として測定されます。モジュール同士が一切インポートし合わないパッケージのスコアは `0` です。

## なぜ問題なのか

パッケージとは、ファイルに分割された単一の概念であるべきです。ファイル同士が参照し合わない場合、そのパッケージは名前空間を共有しているだけの無関係なコードの寄せ集めにすぎません:

- **誤解を招くインポートパス。** `from myapp.utils import X` は、`X` と `utils` 内の他のすべてに関係があることを示唆しますが、凝集度が低いとその約束は空虚です。
- **明確なオーナーの不在。** 「全体」が存在しないため、"`utils`" 全体の責任者がいません。
- **際限のない肥大化。** 雑多なパッケージは、ゴミ捨て場になるまで無関係なヘルパーを蓄積し続けます。
- **欠落した抽象の隠蔽。** 正しい対処は「追加し続ける」ことではなく、2つのサブモジュールが共有している本当の概念を見つけて抽出することである場合が多いです。

## 例

```
myapp/utils/
    __init__.py
    string_utils.py     # slugify, truncate
    math_utils.py       # clamp, lerp
    io_utils.py         # atomic_write, read_json
```

これら3つのモジュールは互いにインポートしていません。`utils` パッケージの凝集度はゼロです。

## 修正例

パッケージを、実際に行っていることに基づいた名前の焦点を絞ったパッケージに分割します:

```
myapp/text/          # slugify, truncate、およびそれらが共有するヘルパー
myapp/geometry/      # clamp, lerp
myapp/fs/            # atomic_write, read_json
```

あるいは、内容が本当に無関係な単発のヘルパーであるなら、そのことを認めて取り繕うのをやめましょう。パッケージを `misc` と名付けるか、各ヘルパーを実際に使用しているモジュールに移動し、そのゴミ捨て場を凝集度チェックから除外します。

## オプション

| オプション | デフォルト | 説明 |
| --- | --- | --- |
| [`architecture.validate_cohesion`](../configuration/reference.md#architecture) | `true` | `false` に設定するとこのルールを無効にします。 |
| [`architecture.min_cohesion`](../configuration/reference.md#architecture) | `0.5` | このスコアを下回るパッケージが検出されます。 |
| [`architecture.enabled`](../configuration/reference.md#architecture) | `true` | アーキテクチャ解析のマスタースイッチ。 |
| [`architecture.fail_on_violations`](../configuration/reference.md#architecture) | `false` | 違反時にゼロ以外の終了コードを返します。 |

## 参照

- パッケージ凝集度の計算 (`internal/analyzer/coupling_metrics.go`, `internal/analyzer/module_analyzer.go`)。
- [ルールカタログ](index.md) · [layer-violation](layer-violation.md)
