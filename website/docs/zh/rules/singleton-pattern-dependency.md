# singleton-pattern-dependency

**Category**: 依赖注入  
**Severity**: Warning  
**Triggered by**: `pyscn analyze`, `pyscn check --select di`

## 检测内容

标记通过在类级 `_instance` 属性上缓存自身来实现单例模式的类。

## 为什么这是一个问题

单例是披着类外衣的全局状态。每个调用 `PaymentGateway.instance()` 的地方都依赖于该类选择返回的那一个对象，而且该对象会在测试之间存活，除非每个测试都记得重置它。一个粗心的测试就会让下一个测试继承陈旧的状态。

因为单例自行决定其生命周期，调用者无法在不同上下文中为它提供不同的协作对象 — 第二套配置、用于测试的假对象、按租户隔离的实例。替换需要深入类内部重置 `_instance`，而这恰恰是单例本应隐藏的耦合。

这种模式还遮蔽了真实的依赖关系：阅读调用 `X.instance()` 的方法代码，无法知道 `X` 需要什么或在哪里被配置。

## 示例

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

## 修正示例

在应用程序边界处构造对象一次，然后传递给需要它的地方。

```python
class PaymentGateway:
    def charge(self, order):
        ...

# wiring, done once at startup
gateway = PaymentGateway()
order_service = OrderService(gateway)
```

## 选项

| 选项 | 默认值 | 说明 |
| --- | --- | --- |
| [`di.enabled`](../configuration/reference.md#di) | `false` | 必须为 `true` 才能在 `analyze` 中运行 DI 规则。 |
| [`di.min_severity`](../configuration/reference.md#di) | `"warning"` | 提高到 `"error"` 可抑制此规则。 |

## 参考

- 隐式依赖检测（`internal/analyzer/hidden_dependency_detector.go`）。
- [规则目录](index.md) · [Global state dependency](global-state-dependency.md) · [Module variable dependency](module-variable-dependency.md)
