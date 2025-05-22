# docs-metadata-crawler
Concurrent web crawler specializing in document discovery and metadata extraction

### Overview

DocsCrawler is an educational web crawler written in Go that discovers and analyzes document files (PDF, DOCX, XLSX, PPTX) from websites. The project demonstrates concurrent programming patterns, HTTP client usage, document processing, and modular architecture design in Go.

**⚠️ Educational Purpose**: This project is primarily designed for learning and demonstration purposes rather than production use.

### Features

- **Concurrent Web Crawling**: Multi-threaded URL discovery with configurable parallelism
- **Document Analysis**: Metadata extraction from PDF and Microsoft Office documents
- **Extensible Architecture**: Easy addition of new document format analyzers
- **JSON Output**: Structured metadata output in JSON format
- **Configurable Parameters**: Command-line interface for customization

### Supported Document Formats

- **PDF**: Title, author, creator, creation date, modification date, etc.
- **Microsoft Office** (DOCX/XLSX/PPTX): Core properties, application properties, statistics

### Installation

```bash
# Clone the repository
git clone https://github.com/yourusername/docscrawler.git
cd docscrawler

# Install dependencies
go mod tidy

# Build the application
go build -o docscrawler
```

### Usage

```bash
# Basic usage - crawl site for all supported document types
./docscrawler -s https://example.com

# Specify document types
./docscrawler -s https://example.com -t pdf -t docx

# Save output to file
./docscrawler -s https://example.com -o results.json

# Configure parallel threads
./docscrawler -s https://example.com -p 50
```

#### Command Line Options

- `-s, --site`: Target website URL (required)
- `-t, --type`: Document types to analyze (pdf, docx, xlsx, pptx). All types if empty
- `-o, --output`: Output file path. Prints to stdout if not specified
- `-p, --paramax`: Maximum number of parallel threads (default: 100)

### Architecture

The project follows a modular architecture with clear separation of concerns:

```
├── main.go              # CLI parsing and application entry point
├── engine.go            # Main crawler engine coordination
├── crawler.go           # URL discovery and HTML parsing
├── urlstorage.go        # Thread-safe URL management
└── researchers/         # Document analysis modules
    ├── researcher.go    # Common interface and utilities
    ├── pdf.go          # PDF document analyzer
    └── msox.go         # Microsoft Office analyzer
```

#### Key Components

- **Engine**: Orchestrates the crawling process through three phases: crawl, analyze, output
- **URL Storage**: Thread-safe storage for discovered URLs with status tracking
- **Researchers**: Pluggable document analyzers implementing a common interface
- **Crawler**: HTML parsing and link extraction functionality

### Extending Document Support

Adding support for new document formats is straightforward:

1. Create a new analyzer implementing the `Researcher` interface:
```go
type NewDocAnalyzer struct {
    // Document-specific fields
}

func (nda *NewDocAnalyzer) Do(url string) error {
    // Implementation for document processing
}

func (nda *NewDocAnalyzer) OutJSON(writer io.Writer) error {
    // JSON serialization implementation
}
```

2. Register the analyzer in `researchers/researcher.go`:
```go
var allFileTypes = map[string]func() Researcher{
    "pdf":  func() Researcher { return newPdf() },
    "docx": func() Researcher { return newMsox() },
    "newext": func() Researcher { return newNewDocAnalyzer() },
}
```

3. Add the file extension to CLI options in `main.go`.

### Testing

The project includes comprehensive tests covering all major components:

```bash
# Run all tests
go test ./...

# Run tests with coverage
./run_tests.sh  # Generates coverage.html
```

**Note**: Tests were primarily written with AI assistance to demonstrate testing patterns and ensure code reliability.

### Dependencies

- `golang.org/x/net/html` - HTML parsing
- `github.com/pdfcpu/pdfcpu` - PDF processing
- `github.com/jessevdk/go-flags` - CLI argument parsing
- `github.com/stretchr/testify` - Testing framework
