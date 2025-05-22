package researchers

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMsoxResearcher(t *testing.T) {
	t.Run("MSOX initialization", func(t *testing.T) {
		msox := newMsox()
		assert.NotNil(t, msox, "MSOX researcher should be initialized")
		assert.IsType(t, &tMsox{}, msox, "Should return correct type")
		assert.Empty(t, msox.Url, "URL should be empty initially")
		assert.Empty(t, msox.CoreProperty.Title, "Title should be empty initially")
		assert.Empty(t, msox.CoreProperty.Creator, "Creator should be empty initially")
	})

	t.Run("Output to JSON", func(t *testing.T) {
		// Create MSOX researcher with test data
		msox := newMsox()
		msox.Url = "https://example.com/test.docx"
		msox.CoreProperty = tCoreProperty{
			Title:          "Test Document",
			Creator:        "Test Creator",
			LastModifiedBy: "Test Modifier",
			Revision:       "1",
			Created:        "2023-01-01T10:00:00Z",
			Modified:       "2023-01-02T11:00:00Z",
			Language:       "en-US",
		}
		msox.AppProperty = tAppProperty{
			Application: "Test App",
			DocSecurity: "0",
			Pages:       "10",
			Words:       "1000",
			Characters:  "5000",
			Company:     "Test Company",
			Lines:       "100",
			Paragraphs:  "50",
			TotalTime:   "60",
			SharedDoc:   "false",
			AppVersion:  "16.0",
		}

		// Write to buffer
		var buf bytes.Buffer
		err := msox.OutJSON(&buf)
		require.NoError(t, err, "JSON output should not error")

		// Check JSON output
		jsonOutput := buf.String()
		assert.Contains(t, jsonOutput, "\"url\":\"https://example.com/test.docx\"", "JSON should contain URL")
		assert.Contains(t, jsonOutput, "\"title\":\"Test Document\"", "JSON should contain title")
		assert.Contains(t, jsonOutput, "\"creator\":\"Test Creator\"", "JSON should contain creator")
		assert.Contains(t, jsonOutput, "\"lastModifiedBy\":\"Test Modifier\"", "JSON should contain lastModifiedBy")
		assert.Contains(t, jsonOutput, "\"application\":\"Test App\"", "JSON should contain application")
		assert.Contains(t, jsonOutput, "\"pages\":\"10\"", "JSON should contain pages count")
		assert.Contains(t, jsonOutput, "\"words\":\"1000\"", "JSON should contain words count")
	})

	t.Run("Error handling for HTTP issues", func(t *testing.T) {
		// Create test server that returns error status
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		}))
		defer ts.Close()

		msox := newMsox()
		err := msox.Do(ts.URL)
		assert.Error(t, err, "Should return error for non-200 HTTP status")
		assert.Contains(t, err.Error(), "failed to download file", "Error should indicate download failure")
	})

	// Note: Complete MSOX parsing tests would require actual Office files
	// Below is a mock test - in a real environment, consider using testdata with real files

	t.Run("Do method sets URL and docType", func(t *testing.T) {
		// This minimal test just verifies the URL and docType are set
		msox := newMsox()

		// Mock server that returns invalid data (not a real Office file)
		// This will cause errors in the ZIP parsing, but we can still check some basic setup
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("Not a real Office file"))
		}))
		defer ts.Close()

		// Call will fail due to invalid data, but URL and docType should be set
		_ = msox.Do(ts.URL)
		assert.Equal(t, ts.URL, msox.Url, "URL should be set even if processing fails")
		assert.Equal(t, "msox", msox.docType, "Document type should be set to msox")
	})
}

// TestIntegrationMSOX is a mock for what an integration test might look like
// For a real test, you would need actual Office files and would enable this test conditionally
func TestIntegrationMSOX(t *testing.T) {
	// Skip this test by default since it requires actual Office files
	t.Skip("Integration test requires actual Office files")

	// In a real test, you might:
	// 1. Set up a test server that serves a real Office file (DOCX, XLSX, PPTX)
	// 2. Create a MSOX researcher and process it
	// 3. Verify the metadata is correctly extracted

	/*
		// Example of what this might look like:
		docxData, err := os.ReadFile("testdata/sample.docx")
		require.NoError(t, err)

		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.wordprocessingml.document")
			w.Write(docxData)
		}))
		defer ts.Close()

		msox := newMsox()
		err = msox.Do(ts.URL)
		require.NoError(t, err)

		assert.Equal(t, "Expected Title", msox.CoreProperty.Title)
		assert.Equal(t, "Expected Creator", msox.CoreProperty.Creator)
		assert.Equal(t, "Expected Pages", msox.AppProperty.Pages)
	*/
}
