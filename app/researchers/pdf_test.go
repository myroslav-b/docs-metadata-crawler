package researchers

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPdfResearcher(t *testing.T) {
	t.Run("PDF initialization", func(t *testing.T) {
		pdf := newPdf()
		assert.NotNil(t, pdf, "PDF researcher should be initialized")
		assert.IsType(t, &tPdf{}, pdf, "Should return correct type")
		assert.Empty(t, pdf.Url, "URL should be empty initially")
		assert.Empty(t, pdf.Title, "Title should be empty initially")
		assert.Empty(t, pdf.Author, "Author should be empty initially")
	})

	t.Run("Output to JSON", func(t *testing.T) {
		// Create PDF researcher with test data
		pdf := newPdf()
		pdf.Url = "https://example.com/test.pdf"
		pdf.Title = "Test Document"
		pdf.Author = "Test Author"
		pdf.Subject = "Test Subject"
		pdf.Creator = "Test Creator"
		pdf.Producer = "Test Producer"
		pdf.CreationDate = "2023-01-01"
		pdf.ModDate = "2023-01-02"

		// Write to buffer
		var buf bytes.Buffer
		err := pdf.OutJSON(&buf)
		require.NoError(t, err, "JSON output should not error")

		// Check JSON output
		jsonOutput := buf.String()
		assert.Contains(t, jsonOutput, "\"url\":\"https://example.com/test.pdf\"", "JSON should contain URL")
		assert.Contains(t, jsonOutput, "\"title\":\"Test Document\"", "JSON should contain title")
		assert.Contains(t, jsonOutput, "\"author\":\"Test Author\"", "JSON should contain author")
		assert.Contains(t, jsonOutput, "\"subject\":\"Test Subject\"", "JSON should contain subject")
		assert.Contains(t, jsonOutput, "\"creator\":\"Test Creator\"", "JSON should contain creator")
		assert.Contains(t, jsonOutput, "\"producer\":\"Test Producer\"", "JSON should contain producer")
		assert.Contains(t, jsonOutput, "\"creation_date\":\"2023-01-01\"", "JSON should contain creation date")
		assert.Contains(t, jsonOutput, "\"mod_date\":\"2023-01-02\"", "JSON should contain modification date")
	})

	t.Run("Error handling for HTTP issues", func(t *testing.T) {
		// Create test server that returns error status
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		}))
		defer ts.Close()

		pdf := newPdf()
		err := pdf.Do(ts.URL)
		assert.Error(t, err, "Should return error for non-200 HTTP status")
		assert.Contains(t, err.Error(), "failed to download file", "Error should indicate download failure")
	})

	// Note: Complete PDF parsing tests would require actual PDF files
	// Below is a mock test - in a real environment, consider using testdata with real PDFs

	t.Run("Do method sets URL", func(t *testing.T) {
		// This minimal test just verifies the URL is set, without testing actual PDF parsing
		pdf := newPdf()

		// Mock server that returns invalid data (not a real PDF)
		// This will cause errors in the PDF parsing, but we can still check that URL is set
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("Not a real PDF"))
		}))
		defer ts.Close()

		// Call will fail due to invalid PDF data, but URL should be set
		_ = pdf.Do(ts.URL)
		assert.Equal(t, ts.URL, pdf.Url, "URL should be set even if processing fails")
		assert.Equal(t, "pdf", pdf.docType, "Document type should be set to pdf")
	})
}

// TestIntegrationPDF is a mock for what an integration test might look like
// For a real test, you would need actual PDF files and would enable this test conditionally
func TestIntegrationPDF(t *testing.T) {
	// Skip this test by default since it requires actual PDF files
	t.Skip("Integration test requires actual PDF files")

	// In a real test, you might:
	// 1. Set up a test server that serves a real PDF file
	// 2. Create a PDF researcher and process it
	// 3. Verify the metadata is correctly extracted

	/*
		// Example of what this might look like:
		pdfData, err := os.ReadFile("testdata/sample.pdf")
		require.NoError(t, err)

		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/pdf")
			w.Write(pdfData)
		}))
		defer ts.Close()

		pdf := newPdf()
		err = pdf.Do(ts.URL)
		require.NoError(t, err)

		assert.Equal(t, "Expected Title", pdf.Title)
		assert.Equal(t, "Expected Author", pdf.Author)
	*/
}
