# unreachable-after-break

**Category**: 到達不能コード  
**Severity**: Critical  
**Triggered by**: `pyscn analyze`, `pyscn check`

## 検出内容

ループ内で `break` 文の後に記述された文を検出します。

## なぜ問題なのか

`break` はそれを囲むループを即座に終了します。同じブロック内でその後に続くものは一度も実行されません。

よくある原因：

- **誤配置されたインクリメントやアキュムレータの更新** -- 最後のイテレーションで実行されるはずだった。
- **残されたログやクリーンアップ** -- リファクタリング中に `break` の下に移動された。
- **制御フローの誤解** -- 作者は `break` がイテレーションの一部だけをスキップすると期待していた。

コードは到達不能なため、テストでカバーできず、内部のバグが表面化することはありません。

## 例

```python
for user in users:
    if user.id == target_id:
        break
        user.last_seen = now()   # ← 実行されない
```

## 修正例

`break` の前に処理を行うか、デッドコードの文を削除してください。

```python
for user in users:
    if user.id == target_id:
        user.last_seen = now()
        break
```

## オプション

| Option | Default | Description |
| --- | --- | --- |
| [`dead_code.detect_after_break`](../configuration/reference.md#dead_code) | `true` | `false` に設定するとこのルールを無効にします。 |
| [`dead_code.min_severity`](../configuration/reference.md#dead_code) | `"warning"` | `"critical"` にするとこの種類の検出結果のみ保持します。`"info"` にするとより多く表示します。 |
| [`dead_code.ignore_patterns`](../configuration/reference.md#dead_code) | `[]` | ソース行に対してマッチする正規表現パターン。マッチした場合は抑制されます。 |

## 参照

- 制御フローグラフの到達可能性分析 (`internal/analyzer/dead_code.go`)。
- [ルールカタログ](index.md) · [Unreachable after continue](unreachable-after-continue.md) · [Unreachable after return](unreachable-after-return.md)
