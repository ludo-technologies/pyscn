# unreachable-after-infinite-loop

**Category**: 到達不能コード  
**Severity**: Warning  
**Triggered by**: `pyscn analyze`, `pyscn check`

## 検出内容

`break` や `return` のない `while True:` のように、到達可能な出口を持たないループの後に記述された文を検出します。

## なぜ問題なのか

ループに抜け出すパスがない場合、実行はループの先に進みません。ループの後に書かれたものはすべてデッドコードです。

これは通常、以下のいずれかです：

- **忘れられた終了条件** -- ループは終了するはずだったが、リファクタリングで `break` が失われた。
- **誤配置されたクリーンアップ** -- 決して return しないワーカーループの後に置かれたシャットダウンやティアダウンコード。
- **コピペのエラー** -- 関数の以前のバージョンから残されたループ後のロジック。

読者は後続のコードがいずれ実行されると期待しますが、実行されません。

## 例

```python
def run_worker(queue):
    while True:
        job = queue.get()
        job.run()
    queue.close()   # ← 実行されない
```

## 修正例

ループに到達可能な出口を設けるか、到達不能な末尾のコードを削除してください。

```python
def run_worker(queue):
    while not queue.closed:
        job = queue.get()
        job.run()
    queue.close()
```

## オプション

| Option | Default | Description |
| --- | --- | --- |
| [`dead_code.enabled`](../configuration/reference.md#dead_code) | `true` | このルールには専用のトグルはありません。`dead_code.enabled` と CFG 分析によって制御されます。 |
| [`dead_code.min_severity`](../configuration/reference.md#dead_code) | `"warning"` | `"critical"` に上げるとこれらの検出結果を非表示にします。`"info"` に下げるとより多く表示します。 |
| [`dead_code.ignore_patterns`](../configuration/reference.md#dead_code) | `[]` | ソース行に対してマッチする正規表現パターン。マッチした場合は抑制されます。 |

## 参照

- 制御フローグラフの到達可能性分析 (`internal/analyzer/dead_code.go`)。
- [ルールカタログ](index.md) · [Unreachable after return](unreachable-after-return.md) · [Unreachable branch](unreachable-branch.md)
