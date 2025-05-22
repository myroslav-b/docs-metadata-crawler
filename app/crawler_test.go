package main

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResolveUrl(t *testing.T) {
	testCases := []struct {
		name     string
		baseURL  string
		href     string
		expected string
		hasError bool
	}{
		{
			name:     "Absolute URL",
			baseURL:  "https://example.com/page",
			href:     "https://example.org/another",
			expected: "https://example.org/another",
			hasError: false,
		},
		{
			name:     "Relative URL",
			baseURL:  "https://example.com/page",
			href:     "documents/doc.pdf",
			expected: "https://example.com/documents/doc.pdf",
			hasError: false,
		},
		{
			name:     "Root relative URL",
			baseURL:  "https://example.com/page/subpage",
			href:     "/documents/doc.pdf",
			expected: "https://example.com/documents/doc.pdf",
			hasError: false,
		},
		{
			name:     "URL with query parameters",
			baseURL:  "https://example.com/page",
			href:     "document?format=pdf&id=123",
			expected: "https://example.com/document?format=pdf&id=123",
			hasError: false,
		},
		{
			name:     "Invalid base URL",
			baseURL:  "://invalid.url",
			href:     "/document.pdf",
			hasError: true,
		},
		{
			name:     "Invalid href URL",
			baseURL:  "https://example.com",
			href:     "://invalid.url",
			hasError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			resolvedURL, err := resolveUrl(tc.baseURL, tc.href)

			if tc.hasError {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.expected, resolvedURL.String())
			}
		})
	}
}

func TestHarv(t *testing.T) {
	// Create a test server with a mock HTML page
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		html := `
		<!DOCTYPE html>
		<html>
		<head>
			<title>Test Page</title>
		</head>
		<body>
			<a href="https://example.com/document1.pdf">Document 1</a>
			<a href="/document2.pdf">Document 2</a>
			<a href="document3.pdf">Document 3</a>
			<a href="https://different-domain.com/doc.pdf">External Document</a>
			<a href="invalid:url">Invalid URL</a>
		</body>
		</html>
		`
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(html))
	}))
	defer ts.Close()

	baseURL, err := url.Parse(ts.URL)
	require.NoError(t, err)

	urlStorage := newUrlStorage()

	// Run the crawler
	harv(baseURL, urlStorage)

	// Check the collected URLs
	urls := urlStorage.getAllUrls()

	// Helper function to check if a URL is in the collection
	findURL := func(target string) bool {
		for _, u := range urls {
			if u.String() == target {
				return true
			}
		}
		return false
	}

	// URLs that should be found
	assert.True(t, findURL("https://example.com/document1.pdf"), "Should find absolute URL")
	assert.True(t, findURL(ts.URL+"/document2.pdf"), "Should find root-relative URL")
	assert.True(t, findURL(ts.URL+"/document3.pdf"), "Should find relative URL")
	assert.True(t, findURL("https://different-domain.com/doc.pdf"), "Should find external domain URL")

	// Don't test for invalid URL since it depends on the implementation
	// whether invalid URLs are silently ignored or added

	// Test harvesting a non-existent URL
	invalidURL, _ := url.Parse("http://non-existent-domain-that-should-fail.example")
	urlStorage2 := newUrlStorage()
	harv(invalidURL, urlStorage2)

	// Should not cause panic and should not add any URLs
	assert.Len(t, urlStorage2.getAllUrls(), 0, "Should not add URLs from non-existent site")
}
