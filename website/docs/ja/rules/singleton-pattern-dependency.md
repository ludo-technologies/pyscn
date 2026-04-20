# singleton-pattern-dependency

**カテゴリ**: 依存性注入  
**重大度**: Warning  
**検出コマンド**: `pyscn analyze`, `pyscn check --select di`

## 検出内容

クラスレベルの `_instance` 属性にキャッシュすることでシングルトンパターンを実装しているクラスを検出します。

## なぜ問題なのか

シングルトンは、クラスの形をしたグローバル状態です。`PaymentGateway.instance()` と書くすべての呼び出し元は、クラスが返すことを選んだ1つのオブジェクトに依存し、そのオブジェクトは各テストがリセットを忘れない限りテスト間で生き残ります。1つのテストがリセットを忘れると、次のテストが古い状態を引き継ぎます。

シングルトンが自身のライフタイムを決定するため、呼び出し元は異なるコンテキストで異なる協調オブジェクトを与えることができません — 別の設定、テスト用のフェイク、テナントごとのインスタンスなど。代替するにはクラスの内部に手を伸ばして `_instance` をリセットする必要がありますが、これはまさにシングルトンが隠そうとしていた結合そのものです。

このパターンは実際の依存も隠します。`X.instance()` を呼び出すメソッドのコードを読んでも、`X` が何を必要とし、どこで設定されたかについて何もわかりません。

## 例

```python
class PaymentGateway:
    _instance = None

    @classmethod
    def instance(cls):
        if cls._instance is None:
            cls._instance = cls()
        return cls._instance

    def charge(self, order):
        ...
```

## 修正例

アプリケーションの境界でオブジェクトを一度構築し、必要な箇所に渡します。

```python
class PaymentGateway:
    def charge(self, order):
        ...

# ワイヤリング、起動時に一度だけ実行
gateway = PaymentGateway()
order_service = OrderService(gateway)
```

## オプション

| オプション | デフォルト | 説明 |
| --- | --- | --- |
| [`di.enabled`](../configuration/reference.md#di) | `false` | `analyze` で DI ルールを実行するには `true` にする必要があります。 |
| [`di.min_severity`](../configuration/reference.md#di) | `"warning"` | `"error"` に上げるとこのルールは非表示になります。 |

## 参照

- 隠れた依存の検出 (`internal/analyzer/hidden_dependency_detector.go`)。
- [ルールカタログ](index.md) · [Global state dependency](global-state-dependency.md) · [Module variable dependency](module-variable-dependency.md)
