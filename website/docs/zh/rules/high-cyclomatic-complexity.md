# high-cyclomatic-complexity

**Category**: 复杂度  
**Severity**: Configurable by threshold  
**Triggered by**: `pyscn analyze`, `pyscn check`

## 检测内容

标记 McCabe 圈复杂度超过配置阈值的函数。每个 `if`、`elif`、`for`、`while`、`except`、`match case` 以及推导式中的布尔子句都会使计数加一。一个没有分支的函数从 1 开始计数。

pyscn 不会将 `and` / `or` 短路运算符计为单独的分支。

## 为什么这是一个问题

高分支数意味着：

- **需要阅读更多路径** — 每个审查者都必须在脑中模拟每个分支才能理解函数的功能。
- **需要测试更多路径** — 完整的分支覆盖要求每条路径都有一个测试；实际上大多数高分支函数的测试都不充分。
- **更高的缺陷密度** — 自 McCabe（1976）以来的实证研究表明复杂度与缺陷率相关。
- **更难安全地修改** — 在一个分支中的小改动可能会悄悄破坏另一个分支。

复杂度超过约 10 的函数通常在做多件可以命名并分离的工作。

## 示例

```python
def price_for(user, cart, coupon, region):
    total = 0
    for item in cart:
        if item.category == "book":
            if region == "EU":
                total += item.price * 0.95
            elif region == "US":
                total += item.price
            else:
                total += item.price * 1.10
        elif item.category == "food":
            if user.is_student:
                total += item.price * 0.90
            else:
                total += item.price
        else:
            total += item.price
    if coupon:
        if coupon.kind == "percent":
            total *= 1 - coupon.value
        elif coupon.kind == "fixed":
            total -= coupon.value
    if total < 0:
        total = 0
    return total
```

圈复杂度：13。

## 修正示例

提取每项定价和优惠券处理逻辑，并用分派表替换嵌套的条件判断。

```python
REGION_BOOK_MULTIPLIER = {"EU": 0.95, "US": 1.00}

def _book_price(item, region):
    return item.price * REGION_BOOK_MULTIPLIER.get(region, 1.10)

def _food_price(item, user):
    return item.price * (0.90 if user.is_student else 1.00)

PRICERS = {"book": _book_price, "food": _food_price}

def _item_price(item, user, region):
    pricer = PRICERS.get(item.category)
    return pricer(item, region) if item.category == "book" else \
           pricer(item, user)   if item.category == "food" else \
           item.price

def _apply_coupon(total, coupon):
    if coupon is None:
        return total
    if coupon.kind == "percent":
        return total * (1 - coupon.value)
    return total - coupon.value

def price_for(user, cart, coupon, region):
    subtotal = sum(_item_price(i, user, region) for i in cart)
    return max(0, _apply_coupon(subtotal, coupon))
```

每个辅助函数的复杂度现在只有 1-3，且各司其职。守卫子句（`if coupon is None: return`）展平了剩余的分支。

## 选项

| 选项 | 默认值 | 说明 |
| --- | --- | --- |
| [`complexity.max_complexity`](../configuration/reference.md#complexity) | `0` | `pyscn check` 强制执行的硬限制。`0` 表示在 `analyze` 中不强制执行；如果未设置，`pyscn check --max-complexity` 使用 `10`。 |
| [`complexity.low_threshold`](../configuration/reference.md#complexity) | `9` | 等于或低于此值的函数报告为低风险。 |
| [`complexity.medium_threshold`](../configuration/reference.md#complexity) | `19` | 高于此值的函数为高风险。 |
| [`complexity.min_complexity`](../configuration/reference.md#complexity) | `1` | 低于此值的函数将从报告中省略。 |

## 参考

- McCabe, T. J. *A Complexity Measure.* IEEE Transactions on Software Engineering, 1976.
- 控制流图构建和圈复杂度计算：`internal/analyzer/complexity.go`、`internal/analyzer/complexity_analyzer.go`、`internal/analyzer/cfg_builder.go`。
- [规则目录](index.md) · [too-many-constructor-parameters](too-many-constructor-parameters.md)
