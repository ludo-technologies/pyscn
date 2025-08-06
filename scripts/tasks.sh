#!/bin/bash
# Task management helper script for pyqol

case "$1" in
  "list")
    echo "📋 All Open Tasks:"
    gh issue list --limit 20
    ;;
  
  "week")
    week=${2:-1}
    echo "📅 Week $week Tasks:"
    gh issue list --milestone "Week $week*"
    ;;
  
  "p0")
    echo "🔴 Critical Priority Tasks (P0):"
    gh issue list --label "P0"
    ;;
  
  "p1")
    echo "🟡 High Priority Tasks (P1):"
    gh issue list --label "P1"
    ;;
  
  "view")
    if [ -z "$2" ]; then
      echo "Usage: ./tasks.sh view <issue-number>"
      exit 1
    fi
    gh issue view $2
    ;;
  
  "start")
    if [ -z "$2" ]; then
      echo "Usage: ./tasks.sh start <issue-number>"
      exit 1
    fi
    echo "Starting work on issue #$2..."
    gh issue edit $2 --add-assignee @me
    gh issue comment $2 --body "🚀 Starting work on this task"
    gh issue view $2
    ;;
  
  "done")
    if [ -z "$2" ]; then
      echo "Usage: ./tasks.sh done <issue-number>"
      exit 1
    fi
    echo "Closing issue #$2..."
    gh issue close $2 --comment "✅ Task completed"
    ;;
  
  "progress")
    echo "📊 Progress Overview:"
    echo ""
    echo "Total Issues:"
    gh issue list --state all --json state --jq 'group_by(.state) | map({(.[0].state): length}) | add'
    echo ""
    echo "By Milestone:"
    for milestone in "Week 1 - Foundation" "Week 2 - Dead Code" "Week 3 - Clone Detection" "Week 4 - Release"; do
      open=$(gh issue list --milestone "$milestone" --state open --json id --jq 'length')
      closed=$(gh issue list --milestone "$milestone" --state closed --json id --jq 'length')
      total=$((open + closed))
      if [ $total -gt 0 ]; then
        percent=$((closed * 100 / total))
        echo "$milestone: $closed/$total ($percent%)"
      fi
    done
    ;;
  
  *)
    echo "📚 pyqol Task Management"
    echo ""
    echo "Usage: ./tasks.sh <command> [options]"
    echo ""
    echo "Commands:"
    echo "  list          - Show all open tasks"
    echo "  week <n>      - Show tasks for week n (default: 1)"
    echo "  p0            - Show critical priority tasks"
    echo "  p1            - Show high priority tasks"
    echo "  view <n>      - View details of issue #n"
    echo "  start <n>     - Start working on issue #n"
    echo "  done <n>      - Mark issue #n as completed"
    echo "  progress      - Show overall progress"
    ;;
esac