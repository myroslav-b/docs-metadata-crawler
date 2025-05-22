package researchers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
)

// tPdf is a researcher for PDF documents
// Extracts metadata from PDF files using pdfcpu library
type tPdf struct {
	docType      string
	Url          string `json:"url,omitempty"`
	FileName     string `json:"source,omitempty"`
	Version      string `json:"version,omitempty"`
	Title        string `json:"title,omitempty"`
	Author       string `json:"author,omitempty"`
	Subject      string `json:"subject,omitempty"`
	Producer     string `json:"producer,omitempty"`
	Creator      string `json:"creator,omitempty"`
	CreationDate string `json:"creation_date,omitempty"`
	ModDate      string `json:"mod_date,omitempty"`
}

// newPdf creates a new PDF document researcher
func newPdf() *tPdf {
	return new(tPdf)
}

// OutJSON serializes the PDF metadata to JSON and writes it to the provided writer
func (pdf *tPdf) OutJSON(writer io.Writer) error {
	data, err := json.Marshal(pdf)
	if err != nil {
		return err
	}
	_, err = writer.Write(data)
	return err
}

// Do performs the analysis of a PDF document at the given URL
// Downloads the file, extracts metadata, and stores it
func (pdf *tPdf) Do(url string) error {
	pdf.docType = "pdf"
	pdf.Url = url

	// Initialize HTTP client with timeout
	client := http.Client{
		Timeout: httpGetTimeout * time.Second,
	}
	resp, err := client.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK { // Check for 200 OK status
		// Can read response body for more detailed error if needed
		return fmt.Errorf("failed to download file: status code %d", resp.StatusCode)
	}

	// Convert response body to a ReadSeeker for PDF operations
	respReadSeeker, err := readCloserToReadSeekerFile(resp.Body)
	if err != nil {
		return err
	}

	// Get PDF information using pdfcpu library
	tmpFileName := respReadSeeker.Name()
	info, err := api.PDFInfo(respReadSeeker, tmpFileName, nil, model.NewDefaultConfiguration())
	if err != nil {
		return err
	}

	// Clean up temporary file
	respReadSeeker.Close()
	err = os.Remove(tmpFileName)
	if err != nil {
		return err
	}

	// Store extracted metadata
	pdf.Title = info.Title
	pdf.Author = info.Author
	pdf.Subject = info.Subject
	pdf.Creator = info.Creator
	pdf.Producer = info.Producer
	pdf.CreationDate = info.CreationDate
	pdf.ModDate = info.ModificationDate

	return nil
}
