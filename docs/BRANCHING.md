# Git Branching Strategy

## Branch Types and Naming Conventions

### Main Branches

#### `main`
- **Purpose**: Production-ready code
- **Protected**: Yes
- **Direct commits**: Not allowed
- **Merge requirements**: PR with approval and passing CI

#### `develop` (optional for future)
- **Purpose**: Integration branch for features
- **Protected**: Yes
- **Direct commits**: Not allowed
- **Merge requirements**: PR with passing CI

### Feature Branches

#### Pattern: `feature/issue-{number}-{short-description}`

**Examples:**
```
feature/issue-1-tree-sitter-integration
feature/issue-3-cfg-implementation
feature/issue-6-apted-algorithm
```

**Rules:**
- Always created from `main`
- Must reference an issue number
- Use kebab-case for description
- Keep description under 30 characters
- Delete after merging

### Bug Fix Branches

#### Pattern: `fix/issue-{number}-{short-description}`

**Examples:**
```
fix/issue-42-parser-panic
fix/issue-55-memory-leak
fix/issue-67-incorrect-cfg-edge
```

**Rules:**
- Created from `main` for production bugs
- Must reference an issue number
- Use clear, specific descriptions
- Delete after merging

### Hotfix Branches

#### Pattern: `hotfix/{version}-{description}`

**Examples:**
```
hotfix/v0.1.1-critical-crash
hotfix/v0.2.0-security-patch
```

**Rules:**
- Only for critical production issues
- Created from `main`
- Merged to `main` immediately after fix
- Tagged with patch version

### Documentation Branches

#### Pattern: `docs/{description}`

**Examples:**
```
docs/api-reference
docs/contributing-guide
docs/architecture-update
```

**Rules:**
- For documentation-only changes
- No issue number required
- Can skip some CI checks

### Refactor Branches

#### Pattern: `refactor/{description}`

**Examples:**
```
refactor/parser-performance
refactor/cfg-memory-optimization
refactor/test-structure
```

**Rules:**
- For code improvements without functionality changes
- Should have an issue for tracking
- Requires thorough testing

### Chore Branches

#### Pattern: `chore/{description}`

**Examples:**
```
chore/update-dependencies
chore/ci-configuration
chore/lint-fixes
```

**Rules:**
- For maintenance tasks
- No functionality changes
- Can be created without issue

### Experimental Branches

#### Pattern: `experiment/{description}`

**Examples:**
```
experiment/llm-integration
experiment/rust-parser
experiment/parallel-analysis
```

**Rules:**
- For proof-of-concept work
- Not intended for direct merge
- May be long-lived
- Should be prefixed with [WIP] in PR

## Branch Lifecycle

### 1. Creation

```bash
# Ensure you have the latest main
git checkout main
git pull origin main

# Create feature branch
git checkout -b feature/issue-1-tree-sitter-integration

# Or use the gh CLI
gh issue develop 1 --checkout
```

### 2. Development

```bash
# Regular commits during development
git add .
git commit -m "feat(parser): add tree-sitter initialization"

# Keep branch updated with main
git fetch origin
git rebase origin/main

# Push to remote
git push -u origin feature/issue-1-tree-sitter-integration
```

### 3. Pull Request

```bash
# Create PR using gh CLI
gh pr create \
  --title "feat: Add tree-sitter integration (#1)" \
  --body "Closes #1" \
  --base main \
  --milestone "v1.1.0"

# Or push and create PR via GitHub web
git push origin feature/issue-1-tree-sitter-integration
```

### 4. Merge

**Merge Strategy**: Squash and merge (default)

```bash
# After PR approval and CI passes
# Via GitHub web interface or:
gh pr merge --squash --delete-branch
```

### 5. Cleanup

```bash
# Delete local branch after merge
git checkout main
git pull origin main
git branch -d feature/issue-1-tree-sitter-integration

# Remote branch is auto-deleted on merge
```

## Commit Message Format

Follow [Conventional Commits](https://www.conventionalcommits.org/):

### Format
```
<type>(<scope>): <subject>

<body>

<footer>
```

### Types
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation
- `style`: Formatting
- `refactor`: Code restructuring
- `perf`: Performance improvement
- `test`: Testing
- `build`: Build system
- `ci`: CI/CD
- `chore`: Maintenance

### Examples

```bash
# Feature
git commit -m "feat(parser): add support for Python 3.11 syntax

- Add pattern matching support
- Handle exception groups
- Update tree-sitter grammar

Closes #123"

# Bug fix
git commit -m "fix(cfg): correctly handle nested loops

The CFG builder was incorrectly connecting break statements
in nested loops. This fix ensures breaks connect to the
correct loop exit block.

Fixes #456"

# Documentation
git commit -m "docs: update installation instructions

Add instructions for Windows users and clarify
Go version requirements."
```

## Branch Protection Rules

### For `main` branch:

1. **Required reviews**: At least 1 approval
2. **Required status checks**:
   - CI/Build must pass
   - Tests must pass
   - Coverage threshold met
3. **Enforce up-to-date**: Branch must be up-to-date with base
4. **No force pushes**: Prevent history rewriting
5. **No deletions**: Prevent branch deletion

## Workflow Examples

### Starting a New Feature

```bash
# 1. Pick an issue
./scripts/tasks.sh view 1

# 2. Assign to yourself
./scripts/tasks.sh start 1

# 3. Create branch
git checkout main
git pull origin main
git checkout -b feature/issue-1-tree-sitter-integration

# 4. Develop with TDD
echo "Write tests first!"

# 5. Commit regularly
git add .
git commit -m "feat(parser): implement basic tree-sitter setup"

# 6. Push and create PR
git push -u origin feature/issue-1-tree-sitter-integration
gh pr create --fill

# 7. After merge, cleanup
git checkout main
git pull origin main
git branch -d feature/issue-1-tree-sitter-integration
```

### Quick Fix

```bash
# 1. Create fix branch
git checkout -b fix/issue-99-parser-panic

# 2. Make fix with test
# ... edit files ...
git add .
git commit -m "fix(parser): handle nil input gracefully"

# 3. Push and create PR
git push -u origin fix/issue-99-parser-panic
gh pr create --title "fix: Handle nil input in parser (#99)" --body "Fixes #99"
```

### Working on Long Feature

```bash
# Keep branch updated with main
git checkout feature/issue-6-apted-algorithm
git fetch origin
git rebase origin/main

# If conflicts occur
git status
# ... resolve conflicts ...
git add .
git rebase --continue

# Force push after rebase (only for feature branches!)
git push --force-with-lease origin feature/issue-6-apted-algorithm
```

## Git Aliases (Optional)

Add to your `.gitconfig`:

```ini
[alias]
    # Create feature branch
    feature = "!f() { git checkout -b feature/issue-$1-$2; }; f"
    
    # Create fix branch
    fix = "!f() { git checkout -b fix/issue-$1-$2; }; f"
    
    # Update branch with main
    update = "!git fetch origin && git rebase origin/main"
    
    # Create PR
    pr = "!gh pr create --fill"
    
    # Cleanup merged branches
    cleanup = "!git branch --merged | grep -v '\\*\\|main' | xargs -n 1 git branch -d"
```

Usage:
```bash
git feature 1 tree-sitter-integration
git fix 42 parser-panic
git update
git pr
git cleanup
```

## Best Practices

### DO:
- ✅ Create a branch for every change
- ✅ Keep branches small and focused
- ✅ Reference issue numbers in branch names
- ✅ Keep branches up-to-date with main
- ✅ Delete branches after merging
- ✅ Write clear commit messages
- ✅ Squash commits before merging
- ✅ Test before pushing

### DON'T:
- ❌ Commit directly to main
- ❌ Leave stale branches
- ❌ Create branches without issues (except docs/chore)
- ❌ Force push to main
- ❌ Merge without CI passing
- ❌ Use generic branch names
- ❌ Keep long-lived feature branches
- ❌ Merge conflicts without testing

## Branch Status Check

```bash
# List all branches
git branch -a

# List merged branches
git branch --merged

# List branches with last commit date
git for-each-ref --sort=-committerdate refs/heads/ --format='%(committerdate:short) %(refname:short)'

# Find stale branches (not updated in 30 days)
git for-each-ref --format='%(committerdate:raw)%(refname:short)' refs/heads/ | \
  awk '$1 < '$(date -d "30 days ago" +%s)' {print $2}'

# Check branch divergence
git log --oneline --graph --decorate --all
```

## Troubleshooting

### Merge Conflicts

```bash
# During rebase
git rebase origin/main
# ... resolve conflicts in editor ...
git add .
git rebase --continue

# Or abort if needed
git rebase --abort
```

### Accidentally Committed to Main

```bash
# Create a branch from current state
git branch feature/issue-X-description

# Reset main to origin
git reset --hard origin/main

# Switch to feature branch
git checkout feature/issue-X-description
```

### Need to Change Branch Name

```bash
# Rename local branch
git branch -m old-name new-name

# Delete old remote branch and push new one
git push origin :old-name
git push -u origin new-name
```

## Summary

This branching strategy ensures:
1. **Clean history**: Through squash merging
2. **Traceability**: Every change linked to an issue
3. **Quality**: CI/CD gates on all changes
4. **Collaboration**: Clear naming and PR process
5. **Flexibility**: Different branch types for different needs

Follow these conventions to maintain a clean, organized repository!