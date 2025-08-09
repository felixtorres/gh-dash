#!/bin/bash

echo "=== Azure DevOps Debug Script ==="
echo

# Check if we're in a git repository
if ! git rev-parse --git-dir > /dev/null 2>&1; then
    echo "❌ Not in a git repository"
    exit 1
fi

echo "✅ In a git repository"

# Check git remote
echo "Git remote origin:"
git remote get-url origin 2>/dev/null || echo "❌ No origin remote found"

# Check for Azure DevOps tokens
echo
echo "Environment variables:"
if [ -n "$AZURE_DEVOPS_TOKEN" ]; then
    echo "✅ AZURE_DEVOPS_TOKEN is set"
else
    echo "❌ AZURE_DEVOPS_TOKEN is not set"
fi

if [ -n "$ADO_PAT" ]; then
    echo "✅ ADO_PAT is set"
else
    echo "❌ ADO_PAT is not set"
fi

if [ -n "$AZURE_PAT" ]; then
    echo "✅ AZURE_PAT is set"
else
    echo "❌ AZURE_PAT is not set"
fi

echo
echo "=== Running gh-dash with debug logging ==="
echo "Check debug.log for detailed information"
echo

# Run with debug flag
./gh-dash --debug

echo
echo "=== Debug log contents ==="
if [ -f debug.log ]; then
    echo "Last 50 lines of debug.log:"
    tail -50 debug.log
else
    echo "❌ debug.log not found"
fi