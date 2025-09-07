#!/bin/bash
# Branch management helper script for pyscn

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to validate branch name
validate_branch_name() {
    local branch=$1
    local valid=0
    
    # Check against patterns
    if [[ $branch =~ ^feature/issue-[0-9]+-[a-z-]+$ ]]; then
        echo -e "${GREEN}✓${NC} Valid feature branch"
        valid=1
    elif [[ $branch =~ ^fix/issue-[0-9]+-[a-z-]+$ ]]; then
        echo -e "${GREEN}✓${NC} Valid fix branch"
        valid=1
    elif [[ $branch =~ ^docs/[a-z-]+$ ]]; then
        echo -e "${GREEN}✓${NC} Valid docs branch"
        valid=1
    elif [[ $branch =~ ^refactor/[a-z-]+$ ]]; then
        echo -e "${GREEN}✓${NC} Valid refactor branch"
        valid=1
    elif [[ $branch =~ ^chore/[a-z-]+$ ]]; then
        echo -e "${GREEN}✓${NC} Valid chore branch"
        valid=1
    elif [[ $branch =~ ^hotfix/v[0-9]+\.[0-9]+\.[0-9]+-[a-z-]+$ ]]; then
        echo -e "${GREEN}✓${NC} Valid hotfix branch"
        valid=1
    elif [[ $branch =~ ^experiment/[a-z-]+$ ]]; then
        echo -e "${YELLOW}⚠${NC} Experimental branch - not for direct merge"
        valid=1
    else
        echo -e "${RED}✗${NC} Invalid branch name: $branch"
        echo ""
        echo "Valid patterns:"
        echo "  feature/issue-{number}-{description}"
        echo "  fix/issue-{number}-{description}"
        echo "  docs/{description}"
        echo "  refactor/{description}"
        echo "  chore/{description}"
        echo "  hotfix/v{version}-{description}"
        echo "  experiment/{description}"
        valid=0
    fi
    
    return $((1 - valid))
}

# Function to create a new branch
create_branch() {
    local type=$1
    local issue=$2
    local description=$3
    local branch_name=""
    
    # Ensure we're on main and up to date
    echo "Updating main branch..."
    git checkout main
    git pull origin main
    
    # Create branch name based on type
    case $type in
        feature|fix)
            if [ -z "$issue" ] || [ -z "$description" ]; then
                echo -e "${RED}Error:${NC} Issue number and description required for $type branch"
                echo "Usage: ./branch.sh create $type <issue-number> <description>"
                exit 1
            fi
            branch_name="$type/issue-$issue-$description"
            ;;
        docs|refactor|chore|experiment)
            if [ -z "$issue" ]; then
                echo -e "${RED}Error:${NC} Description required for $type branch"
                echo "Usage: ./branch.sh create $type <description>"
                exit 1
            fi
            branch_name="$type/$issue"  # issue is actually description in this case
            ;;
        *)
            echo -e "${RED}Error:${NC} Unknown branch type: $type"
            echo "Valid types: feature, fix, docs, refactor, chore, experiment"
            exit 1
            ;;
    esac
    
    # Validate branch name
    if validate_branch_name "$branch_name"; then
        echo -e "${GREEN}Creating branch:${NC} $branch_name"
        git checkout -b "$branch_name"
        echo -e "${GREEN}✓${NC} Branch created and checked out"
        
        # Optional: push to remote
        echo ""
        read -p "Push branch to remote? (y/n) " -n 1 -r
        echo ""
        if [[ $REPLY =~ ^[Yy]$ ]]; then
            git push -u origin "$branch_name"
            echo -e "${GREEN}✓${NC} Branch pushed to remote"
        fi
    fi
}

# Function to list branches
list_branches() {
    local filter=$1
    
    echo -e "${BLUE}Local branches:${NC}"
    if [ -n "$filter" ]; then
        git branch | grep "$filter" || echo "  No branches matching '$filter'"
    else
        git branch
    fi
    
    echo ""
    echo -e "${BLUE}Remote branches:${NC}"
    if [ -n "$filter" ]; then
        git branch -r | grep "$filter" || echo "  No remote branches matching '$filter'"
    else
        git branch -r | head -20
    fi
}

# Function to clean up merged branches
cleanup_branches() {
    echo "Fetching latest from remote..."
    git fetch --prune
    
    echo ""
    echo -e "${BLUE}Merged branches that can be deleted:${NC}"
    git branch --merged main | grep -v "main\|develop\|\*"
    
    echo ""
    read -p "Delete all merged branches? (y/n) " -n 1 -r
    echo ""
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        git branch --merged main | grep -v "main\|develop\|\*" | xargs -r git branch -d
        echo -e "${GREEN}✓${NC} Merged branches deleted"
    fi
    
    # Check for stale branches
    echo ""
    echo -e "${BLUE}Branches not updated in 30+ days:${NC}"
    for branch in $(git for-each-ref --format='%(refname:short)' refs/heads/); do
        last_commit=$(git log -1 --format='%cr' "$branch")
        days_ago=$(git log -1 --format='%cr' "$branch" | grep -o '[0-9]\+' | head -1)
        
        if [[ "$last_commit" == *"month"* ]] || ([[ "$last_commit" == *"day"* ]] && [ "${days_ago:-0}" -gt 30 ]); then
            echo "  $branch (last commit: $last_commit)"
        fi
    done
}

# Function to validate current branch
validate_current() {
    current_branch=$(git branch --show-current)
    echo -e "${BLUE}Current branch:${NC} $current_branch"
    validate_branch_name "$current_branch"
}

# Function to update current branch with main
update_branch() {
    current_branch=$(git branch --show-current)
    
    if [ "$current_branch" == "main" ]; then
        echo "Updating main branch..."
        git pull origin main
    else
        echo "Updating $current_branch with latest main..."
        git fetch origin
        git rebase origin/main
        
        if [ $? -ne 0 ]; then
            echo -e "${YELLOW}⚠${NC} Rebase conflicts detected. Resolve conflicts and run:"
            echo "  git rebase --continue"
            echo "Or abort with:"
            echo "  git rebase --abort"
        else
            echo -e "${GREEN}✓${NC} Branch updated successfully"
        fi
    fi
}

# Main command handler
case "$1" in
    create)
        shift
        create_branch "$@"
        ;;
    
    validate)
        if [ -z "$2" ]; then
            validate_current
        else
            validate_branch_name "$2"
        fi
        ;;
    
    list)
        list_branches "$2"
        ;;
    
    cleanup)
        cleanup_branches
        ;;
    
    update)
        update_branch
        ;;
    
    *)
        echo "🌿 pyscn Branch Management"
        echo ""
        echo "Usage: ./branch.sh <command> [options]"
        echo ""
        echo "Commands:"
        echo "  create <type> <issue> <desc>  - Create a new branch"
        echo "    Types: feature, fix, docs, refactor, chore, experiment"
        echo "    Example: ./branch.sh create feature 1 tree-sitter-integration"
        echo ""
        echo "  validate [branch]             - Validate branch name"
        echo "    Example: ./branch.sh validate feature/issue-1-parser"
        echo ""
        echo "  list [filter]                 - List branches"
        echo "    Example: ./branch.sh list feature"
        echo ""
        echo "  update                        - Update current branch with main"
        echo "    Example: ./branch.sh update"
        echo ""
        echo "  cleanup                       - Remove merged branches"
        echo "    Example: ./branch.sh cleanup"
        echo ""
        echo "Quick Start:"
        echo "  ./branch.sh create feature 1 tree-sitter    # Create feature branch"
        echo "  ./branch.sh create fix 42 parser-panic      # Create fix branch"
        echo "  ./branch.sh create docs api-guide           # Create docs branch"
        ;;
esac