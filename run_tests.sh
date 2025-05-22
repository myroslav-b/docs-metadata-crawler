#!/bin/bash

# Run tests and generate coverage for the root package
go test -v -coverprofile=coverage.out ./...

# If tests passed successfully, show coverage as HTML
if [ $? -eq 0 ]; then
    go tool cover -html=coverage.out -o coverage.html
    echo "Tests executed successfully. Coverage report saved to coverage.html"
    
    # Show overall coverage percentage
    go tool cover -func=coverage.out
else
    echo "Tests completed with errors"
fi
