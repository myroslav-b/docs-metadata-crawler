# DocsCrawler Project Testing

This document describes the structure and execution of tests for the DocsCrawler project.

## Test Structure

Tests are organized according to the project structure:

1. **Root package (main)**
   - `urlstorage_test.go` - tests for URL storage
   - `crawler_test.go` - tests for site crawling functionality
   - `engine_test.go` - tests for the main engine
   - `main_test.go` - tests for command line parameter parsing

2. **Researchers package**
   - `researcher_test.go` - tests for basic analyzer functionality
   - `pdf_test.go` - tests for PDF analyzer
   - `msox_test.go` - tests for Microsoft Office file analyzer

## Running Tests

### Requirements

To run tests, you need:

1. Go version 1.13 or higher
2. Testify library: `go get github.com/stretchr/testify`
3. Other project dependencies

### Running All Tests

```bash
go test ./...
```

### Running with Code Coverage

To run tests with code coverage report generation:

```bash
./run_tests.sh
```

This script will execute tests and generate an HTML coverage report (coverage.html).

### Running Specific Tests

To run tests for a specific file:

```bash
go test -v ./... -run TestUrlStorage
```

## Testing Features

### Mocks and Stubs

Tests use mocks to simulate:
- HTTP servers (using `httptest`)
- Document analyzers (through the `Researcher` interface)
- Input/output streams

### Integration Tests

Some integration tests (especially tests requiring real PDF/DOCX files) 
are disabled by default (via `t.Skip`). To run them:

1. Create a `testdata` directory and add real files to it
2. Remove the `t.Skip()` call from the corresponding tests

## Test File Structure

Each test file follows this structure:

1. Unit tests for individual functions
2. Mocks for external dependencies
3. Integration tests (when needed)

## Known Limitations

- Document analyzer tests require real files for full testing
- Tests for `main.go` don't cover actual program entry due to testing limitations with functions using `os.Exit()`