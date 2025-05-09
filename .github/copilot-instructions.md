# GitHub Copilot Instructions for portfolio-manager-go

## Guidelines

- Whenever adding a new handler endpoint, always remember to add sample curl commands in README.md at the root. Also, if we change the directory structure, also remember to update the README on the Project Structure.
- Whenever adding a new handler endpoint, always update the swagger documentation for the endpoint (using swaggo comments above the handler).

## Mocking Strategy

When writing tests for this codebase, always use the testify mocking framework located in `internal/mocks/testify/` rather than the custom mock implementations in `internal/mocks/`.

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
