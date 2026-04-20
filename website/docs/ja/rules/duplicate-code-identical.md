# duplicate-code-identical

**Category**: 重複コード  
**Severity**: Warning  
**Triggered by**: `pyscn analyze`, `pyscn check --select clones`

## 検出内容

空白、レイアウト、コメントを除いてテキストが完全に一致する2つ以上のコードブロックを検出します（Type-1 クローン、類似度 >= 0.85）。

## なぜ問題なのか

コピペされたコードは最も安価な形の重複であり、保守には最もコストがかかります。ロジックを変更する必要がある場合、すべてのコピーを見つけて更新しなければなりません。一箇所が修正されても他は乖離していき、その不整合がバグになります。

同一のブロックは、振る舞いを追加することなくコードベースを膨張させます。読者は2つの領域が本当に同じであることを確認するために時間を費やし、新しいものを読む代わりに確認作業に終始します。

クローンがリテラルであるため、修正はほぼ常に機械的です。ブロックを関数に抽出し、両方の場所からその関数を呼び出してください。

## 例

```python
def send_welcome_email(user):
    subject = "Welcome"
    body = render_template("welcome.html", user=user)
    msg = Message(subject=subject, body=body, to=user.email)
    smtp.send(msg)
    log.info("sent welcome to %s", user.email)

def send_reset_email(user):
    subject = "Reset"
    body = render_template("reset.html", user=user)
    msg = Message(subject=subject, body=body, to=user.email)
    smtp.send(msg)
    log.info("sent reset to %s", user.email)
```

## 修正例

共通ブロックをヘルパー関数に抽出し、異なる部分を引数として渡してください。

```python
def send_email(user, subject, template, tag):
    body = render_template(template, user=user)
    msg = Message(subject=subject, body=body, to=user.email)
    smtp.send(msg)
    log.info("sent %s to %s", tag, user.email)

def send_welcome_email(user):
    send_email(user, "Welcome", "welcome.html", "welcome")

def send_reset_email(user):
    send_email(user, "Reset", "reset.html", "reset")
```

## オプション

| Option | Default | Description |
| --- | --- | --- |
| [`clones.type1_threshold`](../configuration/reference.md#clones) | `0.85` | 同一として報告されるための最小類似度。 |
| [`clones.similarity_threshold`](../configuration/reference.md#clones) | `0.65` | タイプ別の閾値の前に適用されるグローバルな下限値。 |
| [`clones.min_lines`](../configuration/reference.md#clones) | `5` | フラグメントの最小行数。 |
| [`clones.min_nodes`](../configuration/reference.md#clones) | `10` | フラグメントの最小ASTノード数。 |
| [`clones.enabled_clone_types`](../configuration/reference.md#clones) | `["type1","type2","type4"]` | `"type1"` を含めることでこのルールを有効にします。 |

## 参照

- クローン検出の実装 (`internal/analyzer/clone_detector.go`, `internal/analyzer/apted.go`)。
- [ルールカタログ](index.md) · [リネームされたクローン](duplicate-code-renamed.md) · [変更されたクローン](duplicate-code-modified.md) · [意味的クローン](duplicate-code-semantic.md)
