package researchers

import (
	"archive/zip"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

// tCoreProperty represents core document properties from Office Open XML format
// Found in docProps/core.xml inside Office documents
type tCoreProperty struct {
	XMLName        xml.Name `xml:"coreProperties" json:"coreProperties,omitempty"`
	Title          string   `xml:"title" json:"title,omitempty"`
	Creator        string   `xml:"creator" json:"creator,omitempty"`
	LastModifiedBy string   `xml:"lastModifiedBy" json:"lastModifiedBy,omitempty"`
	Revision       string   `xml:"revision" json:"revision,omitempty"`
	Created        string   `xml:"created" json:"created,omitempty"`
	Modified       string   `xml:"modified" json:"modified,omitempty"`
	Language       string   `xml:"language" json:"language,omitempty"`
}

// tAppProperty represents application-specific properties from Office Open XML format
// Found in docProps/app.xml inside Office documents
type tAppProperty struct {
	XMLName     xml.Name `xml:"Properties" json:"properties,omitempty"`
	Application string   `xml:"Application" json:"application,omitempty"`
	DocSecurity string   `xml:"DocSecurity" json:"doc_security,omitempty"`
	Pages       string   `xml:"Pages" json:"pages,omitempty"`
	Words       string   `xml:"Words" json:"words,omitempty"`
	Characters  string   `xml:"Characters" json:"characters,omitempty"`
	Company     string   `xml:"Company" json:"company,omitempty"`
	Lines       string   `xml:"Lines" json:"lines,omitempty" `
	Paragraphs  string   `xml:"Paragraphs" json:"paragraphs,omitempty"`
	TotalTime   string   `xml:"TotalTime" json:"total_time,omitempty"`
	SharedDoc   string   `xml:"SharedDoc" json:"shared_doc,omitempty"`
	AppVersion  string   `xml:"AppVersion" json:"app_version,omitempty"`
}

// tMsox is a researcher for Microsoft Office Open XML files (docx, xlsx, pptx)
// Extracts metadata from the Office documents
type tMsox struct {
	docType      string
	Url          string `json:"url,omitempty"`
	CoreProperty tCoreProperty
	AppProperty  tAppProperty
}

// newMsox creates a new Microsoft Office document researcher
func newMsox() *tMsox {
	return new(tMsox)
}

// OutJSON serializes the MSOX metadata to JSON and writes it to the provided writer
func (msox *tMsox) OutJSON(writer io.Writer) error {
	data, err := json.Marshal(msox)
	if err != nil {
		return err
	}
	_, err = writer.Write(data)
	return err
}

// Do performs the analysis of a Microsoft Office document at the given URL
// Downloads the file, extracts metadata from core.xml and app.xml, and stores it
func (msox *tMsox) Do(url string) error {
	msox.docType = "msox"
	msox.Url = url

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

	// Convert response body to a ReadSeeker for zip operations
	respReadSeeker, err := readCloserToReadSeekerFile(resp.Body)
	if err != nil {
		return err
	}

	// Get temporary file name
	tmpFileName := respReadSeeker.Name()

	// Open ZIP archive (Office documents are ZIP archives)
	rZip, err := zip.OpenReader(tmpFileName)
	if err != nil {
		return err
	}
	defer rZip.Close()

	// Process files inside the ZIP archive
	for _, fInZip := range rZip.File {
		switch fInZip.Name {
		case "docProps/core.xml":
			rc1, err := fInZip.Open()
			if err != nil {
				return err
			}
			defer rc1.Close()
			err = xml.NewDecoder(rc1).Decode(&msox.CoreProperty)
			//rc.Close()
			if err != nil {
				return err
			}
		case "docProps/app.xml":
			rc2, err := fInZip.Open()
			if err != nil {
				return err
			}
			defer rc2.Close()
			err = xml.NewDecoder(rc2).Decode(&msox.AppProperty)
			//rc.Close()
			if err != nil {
				return err
			}
		default:
			continue
		}
	}

	// Clean up temporary file
	respReadSeeker.Close()
	err = os.Remove(tmpFileName)
	if err != nil {
		return err
	}

	return nil
}
