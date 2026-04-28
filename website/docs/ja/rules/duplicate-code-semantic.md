# duplicate-code-semantic

**Category**: 重複コード  
**Severity**: Warning  
**Triggered by**: `pyscn analyze`, `pyscn check --select clones`

## 検出内容

構文的には異なるが同じ結果を計算するコードブロックを検出します（Type-4 クローン、類似度 >= 0.65）。構造ではなく振る舞いを比較するためにデータフロー分析を使用します。

## なぜ問題なのか

意味的クローンは、レビュー時に気づかない重複です。一方の関数はループを使い、もう一方は内包表記を使います。一方は `update` で辞書を構築し、もう一方はマージ構文を使います。コードは異なるように見えるため目視チェックを通過しますが、両方の実装は同じ仕事をしています。

リスクはあらゆる重複と同じで、変更を複数の場所で行わなければなりません。しかし追加のコストがあります。読者は2つの実装がエッジケースで一致するかどうかを一目で判断できません。ループ版も `None` をスキップするのか？内包表記版は空の入力で例外を発生させるのか？この精神的な検証を毎回行わなければなりません。

単一の実装に統合すれば、検証作業と乖離を排除できます。

## 例

```python
def unique_emails(users):
    seen = set()
    result = []
    for u in users:
        if u.email not in seen:
            seen.add(u.email)
            result.append(u.email)
    return result

def distinct_emails(users):
    return list({u.email: None for u in users}.keys())
```

## 修正例

1つの実装を選び、すべての場所で使用してください。両方に利点がある場合は、1つを残してその理由をドキュメントに記載してください。

```python
def unique_emails(users):
    """Return user emails in first-seen order, without duplicates."""
    return list(dict.fromkeys(u.email for u in users))
```

## オプション

| Option | Default | Description |
| --- | --- | --- |
| [`clones.type4_threshold`](../configuration/reference.md#clones) | `0.65` | 意味的クローンとして報告されるための最小類似度。 |
| [`clones.enable_dfa`](../configuration/reference.md#clones) | `true` | Type-4 検出を支えるデータフロー分析を有効にします。 |
| [`clones.similarity_threshold`](../configuration/reference.md#clones) | `0.65` | タイプ別の閾値の前に適用されるグローバルな下限値。 |
| [`clones.enabled_clone_types`](../configuration/reference.md#clones) | `["type1","type2","type4"]` | `"type4"` を含めることでこのルールを有効にします。 |
| [`clones.min_lines`](../configuration/reference.md#clones) | `5` | フラグメントの最小行数。 |

## 参照

- クローン検出の実装 (`internal/analyzer/clone_detector.go`, `internal/analyzer/apted.go`)。
- [ルールカタログ](index.md) · [同一クローン](duplicate-code-identical.md) · [リネームされたクローン](duplicate-code-renamed.md) · [変更されたクローン](duplicate-code-modified.md)
