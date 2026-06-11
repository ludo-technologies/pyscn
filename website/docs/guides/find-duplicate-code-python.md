---
title: How to Find Duplicate Code in Python
description: A practical guide to finding duplicate code in Python. Learn what code clones are (Type 1-4), which tools detect them, and how to automate detection in CI.
---

# How to Find Duplicate Code in Python

Duplicate code causes real problems. A bug fixed in one copy survives in the others. Every change has to be made in several places. Reviewers waste time re-reading logic they have already seen. And now that AI assistants write a lot of our code, near-identical blocks appear in codebases faster than humans ever copy-pasted them.

This guide explains what counts as "duplicate" code, which tools can find it in Python, and how to run detection locally and in CI.

## Quick answer

To scan a project right now, run:

```bash
uvx pyscn@latest analyze --select clones .
```

This runs [pyscn](https://github.com/ludo-technologies/pyscn)'s clone detection without installing anything. It opens an HTML report that lists every group of duplicated code, with similarity scores and file locations.

## What counts as "duplicate"? The four clone types

Most developers picture duplicate code as literal copy-paste. In research, duplicates are called *code clones*, and they come in four types. The difference matters because most tools only catch the first one or two.

Here is the same function in four variants.

**Type-1: identical code.** A copy-paste, where only whitespace and comments differ:

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

If this block is pasted into another file, perhaps with a comment added, that is a Type-1 clone. It is the easiest type to detect and the easiest to fix: extract a shared function. ([rule: duplicate-code-identical](../rules/duplicate-code-identical.md))

**Type-2: renamed identifiers.** The structure is untouched. Only the names changed:

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

Line-based tools miss this, because no two lines are textually equal. But if you compare the syntax trees with names normalized away, the two functions have exactly the same shape. ([rule: duplicate-code-renamed](../rules/duplicate-code-renamed.md))

**Type-3: modified copies.** Someone copied the function, then added or removed a few statements:

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
    subtotal += shipping        # <- new
    tax = subtotal * 0.1
    total = subtotal + tax
    return round(total, 2)
```

This is the most common clone in real codebases. Someone copied a function, tweaked it for a new case, and moved on. Detecting it means measuring how far apart two trees are (tree edit distance), not just whether they match. ([rule: duplicate-code-modified](../rules/duplicate-code-modified.md))

**Type-4: same behavior, different implementation.** The code was rewritten from scratch but computes the same thing:

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

No text matching or tree matching will connect this to the original. Comparing control-flow structure can. ([rule: duplicate-code-semantic](../rules/duplicate-code-semantic.md))

## Tools that detect duplicate code in Python

Here are the main options and what each one can catch:

| Tool | Detects | Notes |
| --- | --- | --- |
| [pylint](https://pylint.readthedocs.io/) (`R0801`) | Type-1 | Line-based similarity check, bundled with pylint. Catches copy-paste. Renames defeat it. |
| [jscpd](https://github.com/kucherenko/jscpd) | Type-1, partial Type-2 | Token-based, supports 150+ languages. A good fit if one detector has to cover a multi-language repo. |
| [SonarQube](https://www.sonarsource.com/products/sonarqube/) | Type-1, partial Type-2 | A full platform with dashboards and history. Heavier to set up and host. |
| [PMD CPD](https://pmd.github.io/pmd/pmd_userdocs_cpd.html) | Type-1, Type-2 | The classic copy-paste detector. Requires a JVM. |
| [pyscn](https://github.com/ludo-technologies/pyscn) | Type-1 through Type-4 | Python-specific. Uses AST hashing for Types 1-2, tree edit distance (APTED) for Type-3, and control-flow comparison for Type-4. |

One common point of confusion: **[ruff](https://docs.astral.sh/ruff/) does not detect duplicate code.** Ruff is a linter and formatter. It checks how individual lines and statements are written, which is a different job from comparing functions against each other across files. The two kinds of tools complement each other rather than compete.

## Walkthrough: detecting clones with pyscn

Take the four variants above and spread them across two files, `orders.py` and `invoices.py`, with the original pasted into both. Then run:

```bash
uvx pyscn@latest analyze --select clones .
```

pyscn parses every Python file, extracts code fragments, and compares them pairwise. On large codebases it uses [LSH](https://en.wikipedia.org/wiki/Locality-sensitive_hashing) to keep this fast, at over 100,000 lines per second. The summary appears in the terminal:

```text
📊 Analysis Summary:
Health Score: 80/100 (Grade: B)

📈 Detailed Scores:
  Duplication:      0/100 ❌  (10.0% duplication, 1 groups)
```

The HTML report shows all five fragments collected into one clone group, with each pair classified and scored:

| Pair | Classified as | Similarity |
| --- | --- | --- |
| exact copy across the two files | Type-1 | 1.00 |
| original vs. modified copy | Type-2 | 0.85 |
| original vs. rewritten version | Type-4 | 0.94 |

Note the last row. The rewritten `total_for_order` is the variant no text-based tool can connect to the original. pyscn catches it at 0.94 similarity from its control-flow structure.

### Tuning the threshold

The `--clone-threshold` flag (default `0.65`) sets the minimum similarity for a pair to be reported:

```bash
pyscn analyze --select clones --clone-threshold 0.8 .   # stricter: fewer, closer matches
```

For permanent settings, create a `.pyscn.toml` file (or use `[tool.pyscn]` in `pyproject.toml`):

```toml
[clones]
similarity_threshold = 0.8
min_lines = 15        # ignore fragments smaller than this
```

Very short functions are skipped by default (`min_lines`). Below a certain size, similarity stops meaning much, because every two-line getter looks like every other. See the [configuration reference](../configuration/reference.md#clones) for all options, including turning individual clone types on or off.

## Automating detection in CI

`pyscn check` is the CI version of `analyze`. It produces no report, just a pass/fail exit code:

```bash
pyscn check --select clones .
```

As a GitHub Actions step:

```yaml
- uses: actions/setup-python@v5
  with:
    python-version: "3.12"
- run: pipx run pyscn check --select clones .
```

The job fails when new duplication crosses your thresholds. That is the point: detection you have to remember to run is detection that stops happening. See [CI/CD integration](../integrations/ci-cd.md) for complete workflows, and [Pyscn Bot](https://github.com/marketplace/pyscn-bot) if you want reviews posted on pull requests automatically.

## What to do with the findings

Not every clone needs to be removed. A useful order of attack:

1. **Type-1 and Type-2 clones in production code.** Extract a shared function. The fix is mechanical and low-risk precisely because the copies are nearly identical.
2. **Type-3 clones.** Look at what differs between the copies. If the differences are data, extract a function that takes parameters. If the differences are behavior, the copies may be diverging on purpose. Sometimes two call sites genuinely want to evolve separately, and merging them would couple things that should stay independent.
3. **Type-4 clones.** Treat these as signals rather than action items. Two independent implementations of the same logic often mean two people did not know about each other's work. Pick one, or document why both exist.
4. **Clones in test code.** Be more tolerant here. Tests value explicitness over DRY-ness, and some repetition that keeps each test readable on its own is usually worth it.

In practice, it works better to set a strict threshold so the report stays short, fix the top group, and run again. Trying to clear a 40-group report in one big refactor rarely ends well.

## FAQ

**Does ruff detect duplicate code?**
No. Ruff is a linter and formatter and has no clone detection rules. Finding duplicates requires comparing code fragments against each other across files, which is outside what a linter does. Use ruff for style and correctness checks, and a clone detector for duplication. They work well together.

**How much duplication is acceptable?**
There is no universal number. As a rough guide, under 5% duplicated lines is typical for a well-maintained codebase, and anything above 15% usually means systematic copy-paste development. The trend matters more than the number itself. Duplication that grows release after release is the real warning sign.

**Can I detect duplicates across multiple repositories?**
Yes. Point the analyzer at a directory containing both checkouts: `pyscn analyze --select clones repo-a/ repo-b/`. Fragments are compared across everything in scope, so cross-repo clones show up like any other pair.

**Why isn't my duplicated snippet being reported?**
Most likely it is below the minimum fragment size (`min_lines` / `min_nodes` in the configuration). Detectors skip tiny fragments on purpose. At five lines, half the codebase resembles the other half. Lower the limits in `.pyscn.toml` if you want short fragments compared too.

---

*Next: browse the [duplicate code rule catalog](../rules/index.md) to see how each clone type is scored, or the [health score documentation](../output/health-score.md) to see how duplication affects your project grade.*
