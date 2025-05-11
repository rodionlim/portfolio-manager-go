# Contributing to Portfolio Manager

Thank you for considering contributing to Portfolio Manager! To help us maintain a high-quality codebase and a smooth workflow, please follow these guidelines when making contributions.

## Commit Messages: Conventional Commits

We require all commit messages to follow the [Conventional Commits](https://www.conventionalcommits.org/) specification. This helps automate changelogs, versioning, and makes it easier to understand the history of the project.

**Examples:**
- `feat: add support for new asset type`
- `fix: correct calculation for dividend yield`
- `docs: update README with new API usage`
- `test: add tests for portfolio handler`

**Format:**
```
type(scope?): subject
```
- `type`: feat, fix, docs, style, refactor, test, chore, etc.
- `scope`: (optional) area of codebase affected
- `subject`: short description

## Tests Required for Features and Bug Fixes

- **All new features and bug fixes must include appropriate tests.**
- Use the [testify mocking framework](internal/mocks/testify/) for all new tests. Do not use custom mocks in `internal/mocks/`.
- Place tests alongside the code they cover, following the existing structure (e.g., `*_test.go` files).
- Run `make test` to ensure all tests pass before submitting a pull request.

## Pull Requests

- Ensure your branch is up to date with `main` before opening a PR.
- Describe your changes clearly in the PR description.
- **Reference the relevant issue(s) in your PR description.** If there is no existing issue for your change, please create one and link it in your PR.
- Update documentation (README, Swagger, etc.) if your change affects APIs or configuration.
- **When adding new handler endpoints:**
  - Add Swagger documentation comments above your handler (using swaggo format).
  - Regenerate the Swagger documentation by running `make run`.
  - Commit the updated Swagger files (e.g., `docs/swagger.yaml`, `docs/swagger.json`) to the repository.
  - Add sample curl commands for the new endpoint to the `README.md`.

## Code Style

- Follow the existing code style and structure.
- Use `go fmt` to format your code.

## Questions?

If you have any questions, please open an issue or start a discussion.

Thank you for helping make Portfolio Manager better!
