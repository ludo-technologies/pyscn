# Commit Message Convention

## Format

```
<type>: <description>

[optional body]

[optional footer]
```

## Rules

1. **Keep it short**: First line max 50 characters
2. **No scope needed**: Keep it simple
3. **Lowercase**: Start description with lowercase
4. **No period**: No period at the end
5. **Imperative mood**: "add" not "added" or "adds"

## Types (only these 8)

- `feat:` - New feature
- `fix:` - Bug fix  
- `docs:` - Documentation only
- `style:` - Formatting, no code change
- `refactor:` - Code restructuring
- `test:` - Add/update tests
- `perf:` - Performance improvement
- `chore:` - Maintenance, dependencies

## Examples

### Simple (preferred)
```
feat: add tree-sitter parser
fix: handle nil input
docs: update readme
test: add parser tests
refactor: simplify cfg logic
chore: update dependencies
```

### With body (when necessary)
```
fix: correct loop handling in cfg

Previous implementation incorrectly connected
break statements in nested loops.

Fixes #456
```

### With breaking change
```
feat: change api interface

BREAKING CHANGE: Parse() now returns (*AST, error)
instead of just *AST
```

## Quick Reference

```bash
# Feature
git commit -m "feat: add python 3.11 support"

# Fix
git commit -m "fix: handle empty files"

# Docs
git commit -m "docs: add api examples"

# Test
git commit -m "test: add benchmark tests"

# Multiple changes (use body)
git commit -m "feat: add clone detection

- implement apted algorithm
- add similarity threshold
- add tests"
```

## DON'T

❌ `feat(parser): add tree-sitter support` - No scope needed
❌ `feat: Added parser.` - No past tense, no period
❌ `feat: Add Parser` - No title case
❌ `feature: add parser` - Use standard types only
❌ `add parser` - Always include type

## Automated Validation

The CI will check commit messages. To test locally:

```bash
# Check last commit
git log -1 --pretty=format:"%s" | grep -E "^(feat|fix|docs|style|refactor|test|perf|chore): [a-z]"
```

## VS Code Integration

Add to `.vscode/settings.json`:

```json
{
  "git.inputValidation": "always",
  "git.inputValidationLength": 50,
  "git.inputValidationSubjectLength": 50
}
```

## Why Simple?

- Faster to write
- Easier to scan in git log
- Less cognitive overhead
- Focus on the change, not the format

Remember: **Keep it simple, keep it consistent**