# deep-import-chain

**カテゴリ**: モジュール構造  
**重大度**: Info  
**トリガー**: `pyscn analyze`, `pyscn check --select deps`

## 検出内容

プロジェクト内の最長の非循環インポートチェーンが、プロジェクト規模に対して期待される深さを超えた場合に報告します。pyscn は `log₂(module_count) + 1` を基準値として使用します。64モジュールのプロジェクトでは、チェーンの長さは7以下が期待されます。

チェーンとはモジュール依存グラフ上のパスです: `a → b → c → …`、各矢印は `import` を表します。

## なぜ問題なのか

深いチェーンはレイヤリングの不備を示しています。リンクが追加されるたびに、チェーンの末端が使用可能になる前にロード・パース・初期化しなければならないモジュールが増え、また無関係な変更が下流に波及しうるポイントが増えます。

チェーンが深すぎる場合の症状:

- **起動の遅延。** 末端のモジュールをインポートすると、トップレベルの副作用が連鎖的に発生します。
- **脆いテスト。** 末端のユニットテストがチェーン全体を引き込み、上流の何かが変更されると壊れます。
- **隠れた結合。** チェーンの中間にあるモジュールは、しばしば単なるパススルーとして存在し、本当の依存関係を隠しています。
- **推論の困難さ。** コードが存在する単一の「レベル」がありません。

## 例

```
myapp.cli
  → myapp.commands
    → myapp.services
      → myapp.orchestrator
        → myapp.workers
          → myapp.adapters
            → myapp.drivers
```

ドライバーに到達するまで7階層。実際には CLI レイヤーは workers の存在を知る必要がなく、workers も CLI について知る必要はありません。しかし `drivers` への変更は、その上のすべてのレイヤーの再テストを強いる可能性があります。

## 修正例

境界にファサードを導入し、上位レイヤーがチェーンではなく1つのモジュールと通信するようにします:

```
myapp.cli
  → myapp.commands
    → myapp.services        # 単一のエントリーポイント
        (内部で orchestrator / workers / adapters / drivers を接続)
```

または平坦化します: `services`、`orchestrator`、`workers` がすべてコーディネーションを行っているなら、1つのレイヤーに統合し、`adapters` に直接依存させます。

## オプション

| オプション | デフォルト | 説明 |
| --- | --- | --- |
| [`dependencies.find_long_chains`](../configuration/reference.md#dependencies) | `true` | `false` に設定するとこのルールを無効にします。 |
| [`dependencies.enabled`](../configuration/reference.md#dependencies) | `false` | `pyscn check` ではオプトイン。`pyscn analyze` では常に有効です。 |

明示的な深さの閾値はありません。pyscn は最長チェーンを `log₂(module_count) + 1` と比較し、超過した場合に報告します。

## 参照

- モジュール DAG 上の最長パス探索 (`internal/analyzer/module_analyzer.go`, `internal/analyzer/coupling_metrics.go`)。
- [ルールカタログ](index.md) · [circular-import](circular-import.md)
