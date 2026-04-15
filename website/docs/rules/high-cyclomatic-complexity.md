# high-cyclomatic-complexity

**Category**: Complexity  
**Severity**: Configurable by threshold  
**Triggered by**: `pyscn analyze`, `pyscn check`

## What it does

Flags functions whose McCabe cyclomatic complexity exceeds the configured threshold. Each `if`, `elif`, `for`, `while`, `except`, `match case`, and boolean clause inside a comprehension adds one to the count. A straight-line function starts at 1.

pyscn does not count `and` / `or` short-circuit operators as separate branches.

## Why is this a problem?

A high branch count means:

- **More paths to read** — every reviewer has to mentally simulate each branch to understand what the function does.
- **More paths to test** — full branch coverage requires one test per path; most highly-branched functions are under-tested in practice.
- **Higher defect density** — empirical studies since McCabe (1976) correlate complexity with bug rate.
- **Harder to change safely** — a small edit in one branch can silently break another.

Functions above ~10 are usually doing several jobs that could be named and separated.

## Example

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

Cyclomatic complexity: 13.

## Use instead

Extract the per-item pricing and coupon handling, and replace the nested conditional with a dispatch table.

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

Each helper now has complexity 1–3 and a single responsibility. Guard clauses (`if coupon is None: return`) flatten the remaining branches.

## Options

| Option | Default | Description |
| --- | --- | --- |
| [`complexity.max_complexity`](../configuration/reference.md#complexity) | `0` | Hard limit enforced by `pyscn check`. `0` means no enforcement in `analyze`; `pyscn check --max-complexity` uses `10` if unset. |
| [`complexity.low_threshold`](../configuration/reference.md#complexity) | `9` | Functions at or below this are reported as low risk. |
| [`complexity.medium_threshold`](../configuration/reference.md#complexity) | `19` | Above this, a function is high risk. |
| [`complexity.min_complexity`](../configuration/reference.md#complexity) | `1` | Functions below this value are omitted from the report. |

## References

- McCabe, T. J. *A Complexity Measure.* IEEE Transactions on Software Engineering, 1976.
- Control-flow graph construction and cyclomatic counting: `internal/analyzer/complexity.go`, `internal/analyzer/complexity_analyzer.go`, `internal/analyzer/cfg_builder.go`.
- [Rule catalog](index.md) · [too-many-constructor-parameters](too-many-constructor-parameters.md)
