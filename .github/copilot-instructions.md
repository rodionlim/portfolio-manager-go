# GitHub Copilot Instructions for portfolio-manager-go

## Guidelines

- Whenever adding a new handler endpoint, always remember to add sample curl commands in README.md at the root. Also, if we change the directory structure, also remember to update the README on the Project Structure.
- Whenever adding a new handler endpoint, always update the swagger documentation for the endpoint (using swaggo comments above the handler).
- Whenever you amend config.yaml or add new configuration to config.go, always update the Configurations section in README.md to reflect the latest config options and defaults.
- When implementing getter methods that return collections (arrays/slices), always initialize with an empty collection (`[]Type{}`) rather than a nil slice to ensure JSON serialization produces an empty array (`[]`) instead of `null`. This maintains consistency with REST API best practices.
- Whenever testing and starting the server, don't check ./portfolio.log file for logs, instead check the output from the terminal where you started the server. The logs are printed to stdout/stderr, not to the log file.

## Guidelines

- Always remember when building UI components that we support both dark and light themes. When choosing color scheme, make sure to account for both themes (preferbly via useMantineColorScheme) to ensure a consistent user experience.
- Whenever making calls to backend server from the front end, wrap the calls with getUrl function in web/ui/src/utils/url.ts

## Mocking Strategy

When writing tests for this codebase (this only includes modules that are outside web/ui, i.e. non ui components), always use the testify mocking framework located in `internal/mocks/testify/` rather than the custom mock implementations in `internal/mocks/`.

The testify mocking framework is the preferred approach for several reasons:

- It provides a more consistent and maintainable approach to mocking
- It has better support for verifying expectations and method calls
- It integrates well with the rest of our testing strategy

Example of the correct mock implementation:

```go
// Good: Using testify mocks
import (
    "testing"
    "github.com/stretchr/testify/assert"
    "portfolio-manager/internal/mocks/testify"
)

func TestSomething(t *testing.T) {
    mockService := new(testify.MockService)
    mockService.On("MethodName", arg1, arg2).Return(expectedResult)

    // Use the mock in your test

    mockService.AssertExpectations(t)
}
```

Example of what to avoid:

```go
// Avoid: Using custom mocks
import (
    "testing"
    "github.com/stretchr/testify/assert"
    "portfolio-manager/internal/mocks"
)

func TestSomething(t *testing.T) {
    mockService := &mocks.MockService{
        // Custom implementation
    }

    // Use the mock in your test
}
```

Existing tests may still use the old approach, but all new tests should use the testify framework.

## MCP Server Integration

This project includes a Model Context Protocol (MCP) server that allows LLMs to interact with portfolio data. The MCP server runs alongside the HTTP server when enabled in configuration.

### Key Points:

- MCP server is optional and can be enabled/disabled via `config.yaml` (`mcp.enabled`)
- Runs on a separate port from the main HTTP server (default: 8081, configurable via `mcp.port`)
- Provides tools for querying blotter trades and portfolio positions
- Uses the `github.com/mark3labs/mcp-go` library for MCP protocol implementation
- Tool handlers follow the pattern from `cmd/mcp/main.go` using `request.RequireString()`, `request.RequireNumber()` for parameter extraction

### Available MCP Tools:

- `query_blotter_trades`: Query trades with filters (ticker, date range, trade type, limit)
- `get_portfolio_positions`: Get portfolio positions by book

### Configuration Example:

```yaml
mcp:
  enabled: true
  port: 8081
```

When adding new MCP tools:

1. Add the tool definition in `registerTools()` method in `internal/server/mcp.go`
2. Implement the handler function following the existing pattern
3. Use `request.RequireString()` for parameter extraction
4. Return JSON responses using `mcp.NewToolResultText()`

## Copilot Agent Instructions for Portfolio Manager

### Commit Message Policy

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
