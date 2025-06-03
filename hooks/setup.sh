#!/bin/bash

# Git Hooks Setup Script for Portfolio Manager Go
# This script installs Git hooks for the project

set -e

HOOKS_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$HOOKS_DIR")"
GIT_HOOKS_DIR="$PROJECT_ROOT/.git/hooks"

echo "🔧 Setting up Git hooks for Portfolio Manager Go..."
echo "   Project root: $PROJECT_ROOT"
echo "   Git hooks directory: $GIT_HOOKS_DIR"

# Check if we're in a Git repository
if [ ! -d "$PROJECT_ROOT/.git" ]; then
    echo "❌ Error: Not in a Git repository. Please run this script from the project root."
    exit 1
fi

# Create git hooks directory if it doesn't exist
mkdir -p "$GIT_HOOKS_DIR"

# Install pre-commit hook
if [ -f "$HOOKS_DIR/pre-commit" ]; then
    echo "📋 Installing pre-commit hook..."
    cp "$HOOKS_DIR/pre-commit" "$GIT_HOOKS_DIR/pre-commit"
    chmod +x "$GIT_HOOKS_DIR/pre-commit"
    echo "   ✅ pre-commit hook installed"
else
    echo "   ⚠️  pre-commit hook not found in hooks/ directory"
fi

# Check for other hook files and install them
for hook_file in "$HOOKS_DIR"/*; do
    if [ -f "$hook_file" ] && [ "$(basename "$hook_file")" != "setup.sh" ] && [ "$(basename "$hook_file")" != "README.md" ]; then
        hook_name=$(basename "$hook_file")
        if [ "$hook_name" != "pre-commit" ]; then  # Already handled above
            echo "📋 Installing $hook_name hook..."
            cp "$hook_file" "$GIT_HOOKS_DIR/$hook_name"
            chmod +x "$GIT_HOOKS_DIR/$hook_name"
            echo "   ✅ $hook_name hook installed"
        fi
    fi
done

echo ""
echo "🎉 Git hooks setup completed!"
echo ""
echo "📋 Installed hooks:"
for hook in "$GIT_HOOKS_DIR"/*; do
    if [ -f "$hook" ] && [ -x "$hook" ]; then
        echo "   • $(basename "$hook")"
    fi
done

echo ""
echo "🔒 Security features enabled:"
echo "   • Pre-commit hook will prevent committing API keys in config.yaml"
echo "   • Checks for other potentially sensitive data in configuration files"
echo ""
echo "💡 To test the pre-commit hook:"
echo "   1. Add a real API key to config.yaml"
echo "   2. Try to commit the file - it should be blocked"
echo "   3. Remove the API key and commit again - it should succeed"
echo ""
echo "🚀 You're all set! The hooks will now run automatically on git operations."
