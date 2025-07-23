# Copilot Agent Instructions for Portfolio Manager

## PR and Branch Policy

**Default Target Branch: When creating pull requests, always target the `develop` branch by default.** We follow the git flow pattern where features are merged to `develop` first, and then `develop` is merged to `main` for releases.

## Commit Message Policy

When making commits in agent mode, **we require all commit messages to follow the [Conventional Commits](https://www.conventionalcommits.org/) specification**. This helps automate changelogs, versioning, and makes it easier to understand the history of the project.

**Format:**

```
type(scope?): subject
```

- `type`: feat, fix, docs, style, refactor, test, chore, etc.
- `scope`: (optional) area of codebase affected
- `subject`: short description

**Examples:**

- `feat: add support for new asset type`
- `fix: correct calculation for dividend yield`
- `docs: update README with new API usage`
- `test: add tests for portfolio handler`

## Additional Guidelines

- Always follow the existing code style and structure.
- Reference relevant issues in your PRs and commit messages when possible.
- Update documentation and tests as required by your changes.
