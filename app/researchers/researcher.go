package researchers

import (
	"fmt"
	"io"
	"os"
)

// Constants for HTTP timeout and file size limits
const (
	httpGetTimeout = 30                // HTTP request timeout in seconds
	maxFileSize    = 100 * 1024 * 1024 // Maximum file size (100MB)
)

// Map of supported file types to their researcher factory functions
var allFileTypes = map[string]func() Researcher{
	"pdf":  func() Researcher { return newPdf() },
	"docx": func() Researcher { return newMsox() },
	"xlsx": func() Researcher { return newMsox() },
	"pptx": func() Researcher { return newMsox() },
}

// Is checks if the specified file type/extension is supported
func Is(st string) bool {
	_, exist := allFileTypes[st]
	return exist
}

// New creates a new researcher instance for the specified file type
func New(st string) Researcher {
	f := allFileTypes[st]
	return f()
}

// Researcher interface defines the common operations for document metadata extraction
// Implementations should be able to analyze documents and output results as JSON
type Researcher interface {
	OutJSON(writer io.Writer) error // Write metadata as JSON to the provided writer
	Do(url string) error            // Process document at the given URL
}

// readCloserToReadSeekerFile converts an io.ReadCloser to an os.File (which implements io.ReadSeeker)
// This is necessary because many document processing libraries require io.ReadSeeker functionality
// The function creates a temporary file, copies content from the reader, and returns the file
// Caller is responsible for closing and removing the temporary file when finished
func readCloserToReadSeekerFile(rc io.ReadCloser) (*os.File, error) {

	// Create a temporary file
	tmpFile, err := os.CreateTemp("", "readseeker-*")
	if err != nil {
		return nil, err
	}

	// Copy data with size limit
	limitedReader := &io.LimitedReader{R: rc, N: maxFileSize}
	_, err = io.Copy(tmpFile, limitedReader)
	if err != nil {
		tmpFileName := tmpFile.Name()
		tmpFile.Close()
		os.Remove(tmpFileName)
		return nil, err
	}

	// Check if size limit was reached (indicates file is too large)
	if limitedReader.N == 0 {
		tmpFileName := tmpFile.Name()
		tmpFile.Close()
		os.Remove(tmpFileName)
		return nil, fmt.Errorf("file exceeds maximum allowed size of %d bytes", maxFileSize)
	}

	// Seek to beginning of file
	_, err = tmpFile.Seek(0, io.SeekStart)
	if err != nil {
		tmpFileName := tmpFile.Name()
		tmpFile.Close()
		os.Remove(tmpFileName)
		return nil, err
	}

	return tmpFile, nil
}
