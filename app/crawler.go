package main

import (
	"net/http"
	"net/url"
	"time"

	"golang.org/x/net/html"
)

// harv (harvest) extracts all links from the HTML document at the provided URL
// and adds them to the URL storage for further processing
func harv(baseUrl *url.URL, urlStorage *tUrlStorage) {
	// Initialize HTTP client with timeout
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(baseUrl.String())
	if err != nil {
		return
	}
	defer resp.Body.Close()

	// Check if the response is successful
	if resp.StatusCode != http.StatusOK {
		return
	}

	// Parse HTML content
	z := html.NewTokenizer(resp.Body)
	for {
		tt := z.Next()
		switch tt {
		case html.ErrorToken:
			return // End of document
		case html.StartTagToken:
			token := z.Token()

			// Look for <a> tags
			if token.Data == "a" {
				for _, attr := range token.Attr {
					if attr.Key == "href" {
						link := attr.Val

						// Handle relative URLs
						url, err := resolveUrl(baseUrl.String(), link)
						if err != nil {
							continue
						}

						// Add link to results if it's new
						urlStorage.add(url)
					}
				}
			}
		}
	}
}

// resolveUrl converts a relative URL to an absolute URL using the base URL
// Returns a parsed URL object or an error if parsing fails
func resolveUrl(baseStr string, href string) (*url.URL, error) {
	u, err := url.Parse(href)
	if err != nil {
		return nil, err
	}
	base, err := url.Parse(baseStr)
	if err != nil {
		return nil, err
	}
	return base.ResolveReference(u), nil
}
