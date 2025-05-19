# Testing Guide for Pizzeria

This document provides comprehensive guidance on testing the Pizzeria application.

## Testing Structure

The project uses Go's built-in testing framework and follows standard Go testing conventions. Tests are organized into these categories:

1. **Unit Tests**: Testing individual functions and components in isolation
2. **Integration Tests**: Testing interactions between components
3. **Handler Tests**: Testing HTTP handlers with simulated requests
4. **Database Tests**: Testing database operations

## Running Tests

### Using Makefile

The project includes a Makefile with comprehensive test commands:

```bash
# Run all tests
make test

# Run tests with verbose output
make test-v

# Run tests with race detection
make test-race

# Run tests with coverage reporting
make test-cover
```

### Directly with Go

```bash
# Run all tests
go test ./...

# Run tests for a specific package
go test ./internal/models

# Run tests with verbose output
go test -v ./...

# Run tests with race detection
go test -race ./...
```

## Writing Tests

### Test File Naming

Test files should be named with a `_test.go` suffix and placed in the same package as the code they test.

Example:
- Code file: `models.go`
- Test file: `models_test.go`

### Unit Test Example

```go
func TestFlashMessage_GetStatus(t *testing.T) {
    // Arrange
    flashMsg := FlashMessage{
        Active:    true,
        StartDate: yesterday,
        EndDate:   tomorrow,
    }
    
    // Act
    status := flashMsg.GetStatus()
    
    // Assert
    if status != "Active" {
        t.Errorf("Expected status 'Active', got '%s'", status)
    }
}
```

### Handler Test Example

Use the provided test helpers to test handlers:

```go
func TestRepository_Home(t *testing.T) {
    // Create a test repository
    repo := NewTestRepo(t)
    defer CleanTestDB(repo)

    // Set handlers with our test repo
    NewHandlers(repo)

    // Create a test HTTP request
    req, rr := CreateTestRequest(t, "GET", "/", nil)

    // Call the Home handler
    http.HandlerFunc(Repo.Home).ServeHTTP(rr, req)

    // Check response status code
    if status := rr.Code; status != http.StatusOK {
        t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
    }
}
```

## Test Coverage

To generate test coverage reports:

```bash
# Generate and view coverage report in one command (using Makefile)
make test-cover

# Or manually with Go commands:
# Generate coverage profile
go test ./... -coverprofile=coverage.out

# View coverage in browser
go tool cover -html=coverage.out

# Get coverage percentage
go tool cover -func=coverage.out
```

## Mock Objects

For tests requiring external dependencies, use the provided mock objects:

- `NewTestRepo()`: Creates a test repository with a temporary SQLite database
- `CreateTestRequest()`: Creates test HTTP requests with response recorders
- `MockResponseWriter`: A mock `http.ResponseWriter` for testing

## Test Helpers

The project includes several test helper functions to simplify testing:

- `internal/handlers/test_helpers.go`: Helpers for testing HTTP handlers
- `cmd/server/main_test.go`: Integration test helpers

## Best Practices

1. **Use table-driven tests** when testing multiple cases of the same function
2. **Clean up resources** after tests (e.g., temporary databases, files)
3. **Use descriptive test names** to clearly indicate what's being tested
4. **Test both success and error cases**
5. **Avoid testing unexported functions** unless absolutely necessary
6. **Keep tests independent** - tests should not depend on the state from other tests

## Troubleshooting Common Test Issues

### Database-related Issues

If tests involving the database are failing:

1. Check if SQLite3 is properly installed
2. Ensure the test database is being created in a writable location
3. Verify that all required tables are being created in the test setup

### Authentication-related Issues

For tests involving authentication:

1. Make sure cookie secret is properly initialized
2. Set up test environment variables for OAuth
3. Mock OAuth responses rather than making actual API calls

### HTTP Handler Issues

When testing HTTP handlers:

1. Ensure the templates are properly loaded or mocked
2. Check that the middleware chain is properly set up
3. Verify response status codes and content