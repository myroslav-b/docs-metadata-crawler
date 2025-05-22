package researchers

import (
	"bytes"
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResearcherInterfaces(t *testing.T) {
	// Test if the file types are properly registered
	t.Run("Check registered file types", func(t *testing.T) {
		expectedTypes := []string{"pdf", "docx", "xlsx", "pptx"}

		for _, fileType := range expectedTypes {
			assert.True(t, Is(fileType), "Type %s should be registered", fileType)
		}

		assert.False(t, Is("unknown"), "Unknown type should not be registered")
	})

	t.Run("Factory method returns correct types", func(t *testing.T) {
		// PDF researcher
		pdfResearcher := New("pdf")
		assert.NotNil(t, pdfResearcher, "PDF researcher should not be nil")
		assert.IsType(t, &tPdf{}, pdfResearcher, "Should return PDF researcher type")

		// MSOX researchers (docx, xlsx, pptx)
		docxResearcher := New("docx")
		assert.NotNil(t, docxResearcher, "DOCX researcher should not be nil")
		assert.IsType(t, &tMsox{}, docxResearcher, "Should return MSOX researcher type")

		xlsxResearcher := New("xlsx")
		assert.NotNil(t, xlsxResearcher, "XLSX researcher should not be nil")
		assert.IsType(t, &tMsox{}, xlsxResearcher, "Should return MSOX researcher type")

		pptxResearcher := New("pptx")
		assert.NotNil(t, pptxResearcher, "PPTX researcher should not be nil")
		assert.IsType(t, &tMsox{}, pptxResearcher, "Should return MSOX researcher type")
	})
}

func TestReadCloserToReadSeekerFile(t *testing.T) {
	t.Run("Successful conversion", func(t *testing.T) {
		// Create a ReadCloser with test data
		testData := []byte("This is test data for ReadCloser to ReadSeeker conversion")
		reader := io.NopCloser(bytes.NewReader(testData))

		// Convert to ReadSeeker
		readSeeker, err := readCloserToReadSeekerFile(reader)
		require.NoError(t, err, "Should convert without error")
		require.NotNil(t, readSeeker, "ReadSeeker should not be nil")

		// Clean up
		tmpFileName := readSeeker.Name()
		readSeeker.Close()
		os.Remove(tmpFileName)

		// Check if temp file was created in the expected location
		assert.Contains(t, tmpFileName, os.TempDir(), "Temp file should be created in temp directory")
	})

	t.Run("File size limit", func(t *testing.T) {
		// Create a ReadCloser with data larger than the limit
		oversizedData := make([]byte, maxFileSize+1)
		reader := io.NopCloser(bytes.NewReader(oversizedData))

		// Try to convert to ReadSeeker
		readSeeker, err := readCloserToReadSeekerFile(reader)
		assert.Error(t, err, "Should return error for oversized file")
		assert.Nil(t, readSeeker, "ReadSeeker should be nil for oversized file")
		assert.Contains(t, err.Error(), "exceeds maximum allowed size", "Error should mention size limit")
	})

	t.Run("File operations", func(t *testing.T) {
		// Create a small file for testing
		testData := []byte("File operation test data")
		reader := io.NopCloser(bytes.NewReader(testData))

		// Convert to ReadSeeker
		readSeeker, err := readCloserToReadSeekerFile(reader)
		require.NoError(t, err, "Should convert without error")

		// Test seeking and reading
		position, err := readSeeker.Seek(5, io.SeekStart)
		assert.NoError(t, err, "Should seek without error")
		assert.Equal(t, int64(5), position, "Should return correct position")

		buffer := make([]byte, 4)
		n, err := readSeeker.Read(buffer)
		assert.NoError(t, err, "Should read without error")
		assert.Equal(t, 4, n, "Should read requested number of bytes")
		assert.Equal(t, []byte("oper"), buffer, "Should read correct data")

		// Clean up
		tmpFileName := readSeeker.Name()
		readSeeker.Close()
		os.Remove(tmpFileName)
	})
}

// Mock implementation for ReadCloser that returns error on Read
type errorReader struct{}

func (r *errorReader) Read(p []byte) (n int, err error) {
	return 0, io.ErrUnexpectedEOF
}

func (r *errorReader) Close() error {
	return nil
}

func TestReadCloserToReadSeekerFileErrors(t *testing.T) {
	t.Run("Read error", func(t *testing.T) {
		// Create a ReadCloser that returns error
		reader := &errorReader{}

		// Try to convert to ReadSeeker
		readSeeker, err := readCloserToReadSeekerFile(reader)
		assert.Error(t, err, "Should return error when read fails")
		assert.Nil(t, readSeeker, "ReadSeeker should be nil when read fails")
	})
}
