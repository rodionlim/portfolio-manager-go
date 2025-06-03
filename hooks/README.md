# Git Hooks

This directory contains Git hooks for the Portfolio Manager Go project. Git hooks are scripts that run automatically at certain points in the Git workflow to help maintain code quality and security.

## Available Hooks

### pre-commit

- **Purpose**: Prevents accidentally committing sensitive data like API keys
- **Triggers**: Before each commit
- **Checks**:
  - Scans `config.yaml` for real Gemini API keys
  - Warns about other potentially sensitive data patterns
  - Blocks commits containing real API keys
  - Provides helpful guidance on proper API key management

## Setup

### For New Developers

When you clone this repository, run the setup script to install the hooks:

```bash
# From the project root directory
cd hooks
chmod +x setup.sh
./setup.sh
```

Or run it directly from the project root:

```bash
bash hooks/setup.sh
```

### Manual Installation

If you prefer to install hooks manually:

```bash
# Copy hooks to .git/hooks directory
cp hooks/pre-commit .git/hooks/pre-commit
chmod +x .git/hooks/pre-commit
```

## Testing the Hooks

### Testing the pre-commit hook

1. **Test with a real API key** (should be blocked):

   ```bash
   # Edit config.yaml and add a real-looking API key
   echo "geminiApiKey: AIzaSyBxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx" >> config.yaml
   git add config.yaml
   git commit -m "test commit"  # This should be blocked
   ```

2. **Test with a placeholder** (should succeed):
   ```bash
   # Edit config.yaml with a placeholder
   echo "geminiApiKey: \"\"  # Set via GEMINI_API_KEY environment variable" > temp_config.yaml
   mv temp_config.yaml config.yaml
   git add config.yaml
   git commit -m "test commit"  # This should succeed
   ```

## How It Works

### Pre-commit Hook Security Checks

The pre-commit hook performs the following security checks on `config.yaml`:

1. **API Key Detection**: Uses regex patterns to identify potential real API keys
2. **Placeholder Filtering**: Allows common placeholder values like empty strings, "your-api-key", etc.
3. **Length Validation**: Checks for strings that look like real API keys (20+ characters)
4. **Interactive Warnings**: Prompts user for confirmation on potentially sensitive data

### Bypassing Hooks (Emergency Only)

If you need to bypass hooks in an emergency:

```bash
git commit --no-verify -m "emergency commit"
```

**⚠️ Warning**: Only use `--no-verify` when absolutely necessary and ensure no sensitive data is being committed.

## Environment Variable Setup

Instead of putting API keys in `config.yaml`, use environment variables:

```bash
# Add to your shell profile (.bashrc, .zshrc, etc.)
export GEMINI_API_KEY="your-actual-api-key-here"

# Or set it temporarily for one session
export GEMINI_API_KEY="your-actual-api-key-here"
./cmd/portfolio/main
```

The application automatically reads from `GEMINI_API_KEY` environment variable when `geminiApiKey` is empty in the config file.

## Troubleshooting

### Hook Not Running

- Ensure the hook file is executable: `chmod +x .git/hooks/pre-commit`
- Check if the hook file exists: `ls -la .git/hooks/pre-commit`
- Verify you're in the correct Git repository

### False Positives

If the hook incorrectly flags legitimate data:

1. Check if the pattern matches the false positive criteria
2. Update the hook's regex patterns if needed
3. Use the interactive prompt to proceed (for non-API-key warnings)

### Hook Updates

When hooks are updated in the repository:

1. Run `hooks/setup.sh` again to install the latest versions
2. Or manually copy the updated hook files to `.git/hooks/`

## Contributing

When adding new hooks:

1. Add the hook script to the `hooks/` directory
2. Update this README with documentation
3. Update the `setup.sh` script to handle the new hook
4. Test the hook thoroughly before committing
