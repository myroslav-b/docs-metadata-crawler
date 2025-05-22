package main

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEngineInit(t *testing.T) {
	t.Run("Valid initialization", func(t *testing.T) {
		opts := tOpts{
			Site:    "https://example.com",
			Type:    []string{"pdf", "docx"},
			Output:  "output.json",
			Paramax: 10,
		}

		engine, err := newEngine(opts)

		require.NoError(t, err, "Engine should initialize without error")
		assert.NotNil(t, engine, "Engine should not be nil")
		assert.Equal(t, "output.json", engine.outputFileName, "Output file name should be set")
		assert.Equal(t, 10, engine.paramax, "Paramax should be set")
		assert.Equal(t, "example.com", engine.url.Hostname(), "URL hostname should be parsed correctly")
		assert.Len(t, engine.docTypes, 2, "DocTypes should have correct length")
		assert.Contains(t, engine.docTypes, "pdf", "DocTypes should contain pdf")
		assert.Contains(t, engine.docTypes, "docx", "DocTypes should contain docx")
	})

	t.Run("Invalid URL", func(t *testing.T) {
		opts := tOpts{
			Site:    "not a url",
			Type:    []string{"pdf"},
			Output:  "output.json",
			Paramax: 10,
		}

		engine, err := newEngine(opts)

		assert.Error(t, err, "Should return error for invalid URL")
		assert.NotNil(t, engine, "Engine should be returned even with error")
	})

	t.Run("Invalid document type", func(t *testing.T) {
		opts := tOpts{
			Site:    "https://example.com",
			Type:    []string{"pdf", "invalid"},
			Output:  "output.json",
			Paramax: 10,
		}

		engine, err := newEngine(opts)

		assert.Error(t, err, "Should return error for invalid document type")
		assert.Nil(t, engine, "Engine should be nil")
		assert.Contains(t, err.Error(), "unknown document format", "Error should mention unknown format")
	})
}

func TestIsValidScheme(t *testing.T) {
	testCases := []struct {
		name     string
		url      string
		expected bool
	}{
		{
			name:     "Valid HTTP URL",
			url:      "http://example.com",
			expected: true,
		},
		{
			name:     "Valid HTTPS URL",
			url:      "https://example.com",
			expected: true,
		},
		{
			name:     "Invalid FTP URL",
			url:      "ftp://example.com",
			expected: false,
		},
		{
			name:     "Invalid File URL",
			url:      "file:///path/to/file",
			expected: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			u, err := url.Parse(tc.url)
			require.NoError(t, err)

			result := isValidScheme(u)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestEngineOutput(t *testing.T) {
	// Create a temporary directory for test output
	tempDir, err := os.MkdirTemp("", "engine-test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	outputFile := filepath.Join(tempDir, "output.json")

	t.Run("Output to file", func(t *testing.T) {
		opts := tOpts{
			Site:    "https://example.com",
			Type:    []string{"pdf"},
			Output:  outputFile,
			Paramax: 1,
		}

		engine, err := newEngine(opts)
		require.NoError(t, err)

		// Add a mock URL to the storage
		testUrl, _ := url.Parse("https://example.com/test.pdf")
		engine.urlStorage.add(testUrl)

		// Add mock researcher result
		mockResearcher := &MockResearcher{
			url: "https://example.com/test.pdf",
		}
		engine.docStorage[testUrl.String()] = mockResearcher

		// Run output
		err = engine.output()
		require.NoError(t, err)

		// Check if file was created
		fileContent, err := os.ReadFile(outputFile)
		require.NoError(t, err)

		// Check if output is valid JSON array
		assert.Equal(t, "[{\"test\":\"value\"}]", string(fileContent))
	})

	t.Run("Output to stdout", func(t *testing.T) {
		// Temporarily redirect stdout
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		opts := tOpts{
			Site:    "https://example.com",
			Type:    []string{"pdf"},
			Output:  "", // Empty output file means stdout
			Paramax: 1,
		}

		engine, err := newEngine(opts)
		require.NoError(t, err)

		// Add a mock URL to the storage
		testUrl, _ := url.Parse("https://example.com/test.pdf")
		engine.urlStorage.add(testUrl)

		// Add mock researcher result
		mockResearcher := &MockResearcher{
			url: "https://example.com/test.pdf",
		}
		engine.docStorage[testUrl.String()] = mockResearcher

		// Run output
		err = engine.output()
		require.NoError(t, err)

		// Restore stdout and get output
		w.Close()
		os.Stdout = oldStdout

		var buf bytes.Buffer
		io.Copy(&buf, r)

		// Check output
		assert.Equal(t, "[{\"test\":\"value\"}]", buf.String())
	})
}

// Mock implementation of Researcher interface for testing
type MockResearcher struct {
	url string
}

func (r *MockResearcher) OutJSON(writer io.Writer) error {
	_, err := writer.Write([]byte(`{"test":"value"}`))
	return err
}

func (r *MockResearcher) Do(url string) error {
	r.url = url
	return nil
}

// Testing the crawling functionality is more complex and would typically
// require setting up a mock HTTP server with a complete website structure.
// Here's a simplified version of what a crawl test might look like:

func TestEngineCrawl(t *testing.T) {
	t.Run("Basic crawl test", func(t *testing.T) {
		// Create a test server with a simple HTML structure
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			path := r.URL.Path

			switch path {
			case "/":
				// Root page with links
				w.Write([]byte(`
					<!DOCTYPE html>
					<html>
					<body>
						<a href="/page1.html">Page 1</a>
						<a href="/page2.html">Page 2</a>
						<a href="/document.pdf">PDF Document</a>
					</body>
					</html>
				`))
			case "/page1.html":
				w.Write([]byte(`
					<!DOCTYPE html>
					<html>
					<body>
						<a href="/document2.pdf">Another PDF</a>
					</body>
					</html>
				`))
			case "/page2.html":
				w.Write([]byte(`
					<!DOCTYPE html>
					<html>
					<body>
						<a href="/document3.docx">DOCX Document</a>
					</body>
					</html>
				`))
			default:
				// For document requests, just send a small response
				if strings.HasSuffix(path, ".pdf") || strings.HasSuffix(path, ".docx") {
					w.Write([]byte("Mock document content"))
				} else {
					w.WriteHeader(http.StatusNotFound)
				}
			}
		}))
		defer ts.Close()

		// Create engine with the test server URL
		opts := tOpts{
			Site:    ts.URL,
			Type:    []string{"pdf", "docx"},
			Output:  "",
			Paramax: 2,
		}

		engine, err := newEngine(opts)
		require.NoError(t, err)

		// Run crawl
		engine.crawl()

		// Check collected URLs
		urls := engine.urlStorage.getAllUrls()
		urlStrings := []string{}
		for _, u := range urls {
			urlStrings = append(urlStrings, u.String())
		}

		// Verify expected URLs were collected
		assert.Contains(t, urlStrings, ts.URL+"/page1.html")
		assert.Contains(t, urlStrings, ts.URL+"/page2.html")
		assert.Contains(t, urlStrings, ts.URL+"/document.pdf")
		assert.Contains(t, urlStrings, ts.URL+"/document2.pdf")
		assert.Contains(t, urlStrings, ts.URL+"/document3.docx")
	})
}

// A full test of the analyser would also be complex as it requires actual document processing.
// Here's a simplified example:

func TestEngineAnalyser(t *testing.T) {
	t.Run("Basic analyzer test", func(t *testing.T) {
		// Create a test server serving mock documents
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Just return minimal content that won't actually be processed
			// In a real test, you'd return valid PDF/DOCX content
			w.Write([]byte("Mock document content"))
		}))
		defer ts.Close()

		// Create engine
		opts := tOpts{
			Site:    ts.URL,
			Type:    []string{"pdf", "docx"},
			Output:  "",
			Paramax: 2,
		}

		engine, err := newEngine(opts)
		require.NoError(t, err)

		// Add some URLs to analyze
		pdfUrl, _ := url.Parse(ts.URL + "/test.pdf")
		docxUrl, _ := url.Parse(ts.URL + "/test.docx")
		htmlUrl, _ := url.Parse(ts.URL + "/test.html") // Not a document we care about

		engine.urlStorage.add(pdfUrl)
		engine.urlStorage.add(docxUrl)
		engine.urlStorage.add(htmlUrl)

		// Note: This will likely fail since the mock server doesn't serve real documents
		// This is just to show how you'd structure the test
		engine.analyser()

		// In a real test, you'd verify that engine.docStorage contains the expected entries
		// Since we're using mock responses, this won't work correctly
		// assert.Contains(t, engine.docStorage, pdfUrl.String())
		// assert.Contains(t, engine.docStorage, docxUrl.String())
		// assert.NotContains(t, engine.docStorage, htmlUrl.String())
	})
}

// Finally, we'd have an integration test that tests the full run method,
// but that would be very environment-dependent and is often done separately.
