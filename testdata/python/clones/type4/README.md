# Type-4 Clone Test Data

Type-4 clones are **semantically similar** but **syntactically different** code fragments.
They compute the same result using different algorithms or control flow structures.

## Test Pairs

### 1. Sum of Numbers
- `sum_iterative.py` - Uses loops (for/while)
- `sum_recursive.py` - Uses recursion

### 2. Find Maximum
- `find_max_a.py` - Uses explicit index tracking with if statements
- `find_max_b.py` - Uses enumerate with ternary expressions

### 3. Filter and Transform
- `filter_transform_a.py` - Single-pass with accumulator
- `filter_transform_b.py` - Two-pass (filter then transform)

## Expected Behavior

When running clone detection with DFA enabled:
```bash
pyscn analyze testdata/python/clones/type4/
```

Functions with matching semantics should be detected as Type-4 clones,
even though their syntax differs significantly.

## DFA Features That Help Detection

- **Def-Use Chain Patterns**: Similar variable definition and usage patterns
- **Cross-Block Pairs**: Similar data flow across control structures
- **Chain Length Distribution**: Similar variable reuse patterns
