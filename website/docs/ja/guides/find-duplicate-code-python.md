---
title: Python で重複コードを見つける方法
description: Python で重複コードを見つけるための実践ガイドです。コードクローンの種類（Type 1〜4）、対応ツール、CI での自動検出方法を解説します。
---

# Python で重複コードを見つける方法

重複コードは実際の問題を引き起こします。一方のコピーで修正したバグが、もう一方に残り続けます。変更のたびに複数の箇所を同時に直さなければなりません。レビュアーは既に読んだロジックを何度も読み返す羽目になります。しかも今やAIアシスタントが大量のコードを生成するため、ほぼ同一のブロックが人間のコピー&ペーストをはるかに上回るペースでコードベースに増殖します。

このガイドでは、「重複」コードの定義、Python でそれを検出するツールの種類、そしてローカルおよび CI での検出手順を説明します。

## 手っ取り早く試す

今すぐプロジェクトをスキャンするには、次のコマンドを実行してください:

```bash
uvx pyscn@latest analyze --select clones .
```

これは何もインストールせずに [pyscn](https://github.com/ludo-technologies/pyscn) のクローン検出を実行します。重複コードのグループを類似度スコアとファイル位置付きで一覧表示する HTML レポートが開きます。

## 「重複」の定義: 4 種類のクローン

多くの開発者は重複コードをコピー&ペーストそのものとして捉えています。研究分野では重複コードを*コードクローン*と呼び、4 つのタイプに分類されます。この違いが重要なのは、ほとんどのツールが最初の 1〜2 種類しか検出できないからです。

以下は同じ関数の 4 つのバリアントです。

**Type-1: 完全一致。** 空白とコメントだけが異なるコピー&ペーストです:

```python
def calculate_order_total(items, discount_rate):
    subtotal = 0.0
    for item in items:
        price = item["price"]
        quantity = item["quantity"]
        if quantity <= 0:
            continue
        subtotal += price * quantity
    if discount_rate > 0:
        subtotal = subtotal * (1 - discount_rate)
    tax = subtotal * 0.1
    total = subtotal + tax
    return round(total, 2)
```

このブロックを別のファイルに貼り付け、コメントを追加しただけのものが Type-1 クローンです。最も検出しやすく、修正も簡単です。共通の関数に切り出せばよいだけです。（[ルール: duplicate-code-identical](../rules/duplicate-code-identical.md)）

**Type-2: 識別子のリネーム。** 構造はそのままで、名前だけが変わっています:

```python
def compute_cart_amount(products, rebate):
    amount = 0.0
    for product in products:
        cost = product["price"]
        count = product["quantity"]
        if count <= 0:
            continue
        amount += cost * count
    if rebate > 0:
        amount = amount * (1 - rebate)
    levy = amount * 0.1
    result = amount + levy
    return round(result, 2)
```

行ベースのツールはこれを見逃します。どの行もテキストとして一致しないからです。しかし名前を正規化した上で構文木を比較すると、2 つの関数はまったく同じ形をしています。（[ルール: duplicate-code-renamed](../rules/duplicate-code-renamed.md)）

**Type-3: 変更されたコピー。** 関数をコピーした後、いくつかの文を追加・削除しています:

```python
def calculate_quote_total(items, discount_rate, shipping=0.0):
    subtotal = 0.0
    for item in items:
        price = item["price"]
        quantity = item["quantity"]
        if quantity <= 0:
            continue
        subtotal += price * quantity
    if discount_rate > 0:
        subtotal = subtotal * (1 - discount_rate)
    subtotal += shipping        # <- 追加
    tax = subtotal * 0.1
    total = subtotal + tax
    return round(total, 2)
```

これは実際のコードベースで最もよく見られるクローンです。関数をコピーして新しいケース向けに少し手を加え、そのまま放置するパターンです。検出には 2 つの木の距離を測る（木編集距離）必要があり、単純なマッチングでは対応できません。（[ルール: duplicate-code-modified](../rules/duplicate-code-modified.md)）

**Type-4: 同じ動作、異なる実装。** ゼロから書き直したにもかかわらず、同じ処理を行っています:

```python
def total_for_order(items, discount_rate):
    valid_items = []
    for item in items:
        if item["quantity"] > 0:
            valid_items.append(item)
    subtotal = sum(
        item["price"] * item["quantity"]
        for item in valid_items
    )
    if discount_rate > 0:
        subtotal = subtotal * (1 - discount_rate)
    total_with_tax = subtotal * 1.1
    return round(total_with_tax, 2)
```

テキスト照合でも木の照合でも、このコードを元の関数と結びつけることはできません。制御フローの構造を比較することで初めて関連性がわかります。（[ルール: duplicate-code-semantic](../rules/duplicate-code-semantic.md)）

## Python で重複コードを検出するツール

主なツールとそれぞれが検出できる範囲をまとめます:

| ツール | 検出対象 | 備考 |
| --- | --- | --- |
| [pylint](https://pylint.readthedocs.io/)（`R0801`） | Type-1 | 行ベースの類似チェック。pylint に同梱。コピー&ペーストを検出。リネームには対応不可。 |
| [jscpd](https://github.com/kucherenko/jscpd) | Type-1、Type-2 の一部 | トークンベース。150 以上の言語をサポート。複数言語の混在リポジトリに向く。 |
| [SonarQube](https://www.sonarsource.com/products/sonarqube/) | Type-1、Type-2 の一部 | ダッシュボードと履歴を持つフルプラットフォーム。セットアップとホスティングに手間がかかる。 |
| [PMD CPD](https://pmd.github.io/pmd/pmd_userdocs_cpd.html) | Type-1、Type-2 | 定番のコピー&ペースト検出器。JVM が必要。 |
| [pyscn](https://github.com/ludo-technologies/pyscn) | Type-1〜Type-4 | Python 専用。Type 1〜2 に AST ハッシュ、Type-3 に木編集距離（APTED）、Type-4 に制御フロー比較を使用。 |

よくある誤解を一点補足します。**[ruff](https://docs.astral.sh/ruff/) は重複コードを検出しません。** ruff はリンター兼フォーマッターで、個々の行や文の書き方をチェックするツールです。ファイルをまたいでコードフラグメントを比較する処理とは役割が異なります。2 種類のツールは競合するのではなく、互いを補い合う関係です。

## ウォークスルー: pyscn でクローンを検出する

上記の 4 つのバリアントを `orders.py` と `invoices.py` の 2 ファイルに分散させ、元の関数を両方に貼り付けておきます。次のコマンドを実行してください:

```bash
uvx pyscn@latest analyze --select clones .
```

pyscn はすべての Python ファイルをパースし、コードフラグメントを抽出してペアごとに比較します。大規模なコードベースでは [LSH](https://en.wikipedia.org/wiki/Locality-sensitive_hashing) を活用して 100,000 行/秒以上の速度を維持します。ターミナルには次のようなサマリーが表示されます:

```text
📊 Analysis Summary:
Health Score: 80/100 (Grade: B)

📈 Detailed Scores:
  Duplication:      0/100 ❌  (10.0% duplication, 1 groups)
```

HTML レポートでは、5 つのフラグメントがすべて 1 つのクローングループにまとめられ、各ペアの分類とスコアが確認できます:

| ペア | 分類 | 類似度 |
| --- | --- | --- |
| 2 ファイル間の完全一致コピー | Type-1 | 1.00 |
| 元の関数と変更されたコピー | Type-2 | 0.85 |
| 元の関数と書き直したバージョン | Type-4 | 0.94 |

最後の行に注目してください。書き直した `total_for_order` は、テキストベースのツールでは元の関数と結びつけられないバリアントです。pyscn は制御フローの構造から 0.94 の類似度として検出します。

### 閾値のチューニング

`--clone-threshold` フラグ（デフォルト `0.65`）で、ペアを報告するための最小類似度を設定できます:

```bash
pyscn analyze --select clones --clone-threshold 0.8 .   # stricter: fewer, closer matches
```

設定を永続化するには `.pyscn.toml` ファイルを作成するか（または `pyproject.toml` の `[tool.pyscn]` セクションを使用）、以下のように記述してください:

```toml
[clones]
similarity_threshold = 0.8
min_lines = 15        # ignore fragments smaller than this
```

非常に短い関数はデフォルトでスキップされます（`min_lines`）。一定サイズ以下では類似度があまり意味を持たず、2 行の getter はどれも似たようなものになるためです。個別のクローンタイプをオン・オフにする方法を含む全オプションは [設定リファレンス](../configuration/reference.md#clones) を参照してください。

## CI での検出を自動化する

`pyscn check` は `analyze` の CI 向けバージョンです。レポートは生成せず、パス/フェイルの終了コードだけを返します:

```bash
pyscn check --select clones .
```

GitHub Actions のステップとして記述する場合:

```yaml
- uses: actions/setup-python@v5
  with:
    python-version: "3.12"
- run: pipx run pyscn check --select clones .
```

新たな重複が設定した閾値を超えるとジョブが失敗します。それがこの仕組みのポイントです。手動で実行することを覚えていなければならない検出は、やがて実行されなくなります。完全なワークフローは [CI/CD 連携](../integrations/ci-cd.md) を、プルリクエストに自動でレビューを投稿したい場合は [Pyscn Bot](https://github.com/marketplace/pyscn-bot) を参照してください。

## 検出結果の対処法

すべてのクローンを除去する必要はありません。推奨する対処の優先順位は次のとおりです:

1. **本番コードの Type-1・Type-2 クローン。** 共通の関数に抽出してください。コピーがほぼ同一のため、修正は機械的かつリスクが低いです。
2. **Type-3 クローン。** コピー間で何が異なるかを確認してください。差分がデータなら、パラメータを受け取る関数に切り出します。差分が振る舞いなら、コピーが意図的に分岐している可能性があります。2 つの呼び出し元が本当に独立して進化すべき場合、無理に統合するとかえって不要な結合を生みます。
3. **Type-4 クローン。** アクションアイテムというよりシグナルとして扱ってください。同じロジックの独立した実装が 2 つ存在する場合、多くは互いの作業を知らなかったことを意味します。どちらか一方を採用するか、両方が存在する理由をドキュメントに残してください。
4. **テストコードのクローン。** ここではより寛容に構えてください。テストは DRY 性よりも明示性を重視します。各テストを単独で読みやすくするための適度な繰り返しは、たいていの場合で許容できます。

実際には、厳しめの閾値でレポートを短く保ち、上位グループを修正してから再実行するサイクルが効果的です。40 グループものレポートを一度の大規模リファクタリングで解消しようとしても、うまくいくことはほとんどありません。

## FAQ

**ruff は重複コードを検出しますか?**
検出しません。ruff はリンター兼フォーマッターであり、クローン検出のルールを持っていません。重複の発見にはファイルをまたいでコードフラグメントを比較する処理が必要で、これはリンターの役割の外です。スタイルや正確性のチェックには ruff を、重複の検出にはクローン検出ツールを使ってください。2 つはうまく組み合わせて使えます。

**どの程度の重複は許容範囲ですか?**
普遍的な基準はありません。おおよその目安として、重複行が 5% 未満なら適切に管理されたコードベースとして典型的な水準です。15% を超えると、組織的なコピー&ペースト開発が行われていることを示すことが多いです。絶対値よりもトレンドのほうが重要です。リリースごとに重複が増え続けることが本当の警戒サインです。

**複数のリポジトリをまたいで重複を検出できますか?**
できます。両方のチェックアウトを含むディレクトリをアナライザーに指定してください: `pyscn analyze --select clones repo-a/ repo-b/`。スコープ内のすべてのコードが比較対象になるため、リポジトリをまたいだクローンも通常のペアと同様に検出されます。

**重複している箇所が報告されないのはなぜですか?**
ほとんどの場合、フラグメントの最小サイズ（設定の `min_lines` / `min_nodes`）を下回っているためです。検出器は意図的に短いフラグメントをスキップします。5 行程度だとコードベースの半分が互いに似通ってしまうからです。短いフラグメントも比較したい場合は、`.pyscn.toml` で制限値を下げてください。

---

*次のステップ: 各クローンタイプのスコアリングを確認するには [重複コードルールカタログ](../rules/index.md) を、重複がプロジェクトグレードに与える影響を確認するには [ヘルススコアのドキュメント](../output/health-score.md) をご覧ください。*
