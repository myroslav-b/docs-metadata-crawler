package main

import (
	"bufio"
	"docscrawler/app/researchers"
	"errors"
	"fmt"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"
)

// Time to wait between checks for available crawl threads
const crawlSleepTime = 5 * time.Second

// tEngine represents the main crawler engine
// Manages URL and document storages, processing parameters, and output configuration
type tEngine struct {
	url            *url.URL                          // Base URL to start crawling from
	urlStorage     *tUrlStorage                      // Storage for URLs discovered during crawling
	docStorage     map[string]researchers.Researcher // Storage for processed documents
	docTypes       []string                          // Document types/extensions to look for
	outputFileName string                            // Output file name (stdout if empty)
	paramax        int                               // Maximum number of parallel threads
	mutex          sync.Mutex                        // Mutex for thread-safe operations
}

// newEngine initializes a new crawler engine with the provided options
func newEngine(opts tOpts) (*tEngine, error) {

	engine := new(tEngine)
	engine.urlStorage = newUrlStorage()
	engine.docStorage = make(map[string]researchers.Researcher)
	engine.docTypes = make([]string, len(opts.Type))

	// Validate document types
	for i, st := range opts.Type {
		ok := researchers.Is(st)
		if !ok {
			return nil, errors.New("unknown document format for analysis")
		}
		engine.docTypes[i] = st
	}

	engine.outputFileName = opts.Output

	engine.paramax = opts.Paramax

	// Parse and validate the starting URL
	var err error
	engine.url, err = url.ParseRequestURI(opts.Site)
	if err != nil {
		return engine, errors.New("invalid URL")
	}

	return engine, nil
}

// run executes the three main phases of the crawling process:
// 1. crawl - discover URLs
// 2. analyser - process documents
// 3. output - generate results
func (engine *tEngine) run() {
	engine.crawl()

	_ = engine.analyser()

	err := engine.output()
	if err != nil {
		fmt.Println(err.Error())
	}

}

// crawl recursively discovers URLs starting from the base URL
// Uses a worker pool pattern with a guard channel to limit concurrent operations
func (engine *tEngine) crawl() {
	guard := make(chan bool, engine.paramax)
	defer close(guard)

	hostname := engine.url.Hostname()
	harv(engine.url, engine.urlStorage)

	for {
		urlBase, ok := engine.urlStorage.use()
		switch {
		case !ok && (len(guard) == 0):
			// No more URLs to process and no active workers
			return
		case !ok && (len(guard) > 0):
			// No URLs to process but workers are still active, wait
			time.Sleep(crawlSleepTime)
		case ok:
			if isValidScheme(urlBase) && (hostname == urlBase.Hostname()) {
				guard <- true
				urlCopy := *urlBase
				go func(u *url.URL) {
					harv(u, engine.urlStorage)
					<-guard
				}(&urlCopy)
			}
		}
	}
}

// analyser processes discovered URLs looking for document files of specified types
// Uses a worker pool pattern with a guard channel to limit concurrent operations
func (engine *tEngine) analyser() error {

	guard := make(chan bool, engine.paramax)
	defer close(guard)

	var wg sync.WaitGroup

	for _, url := range engine.urlStorage.getAllUrls() {
		url := url
		guard <- true
		wg.Add(1)
		go func() {
			defer wg.Done()
			engine.mutex.Lock()
			defer engine.mutex.Unlock()

			// Process URL if it has a matching document extension
			for _, t := range engine.docTypes {
				if strings.HasSuffix(url.String(), "."+t) {
					eng := researchers.New(t)
					err := eng.Do(url.String())
					if err == nil {
						engine.docStorage[url.String()] = eng
					}
					break
				}
			}
			<-guard

		}()
	}

	wg.Wait()

	return nil
}

// isValidScheme checks if the URL uses a supported protocol (http or https)
func isValidScheme(u *url.URL) bool {
	return u.Scheme == "http" || u.Scheme == "https"
}

// output writes the analysis results to the specified output file or stdout
// Output is in JSON array format containing document metadata
func (engine *tEngine) output() error {
	//st := ""
	var out *os.File
	var err error

	// Determine output destination (file or stdout)
	if engine.outputFileName == "" {
		out = os.Stdout
	} else {
		out, err = os.Create(engine.outputFileName)
		if err != nil {
			return err
		}
		defer out.Close()
	}

	bufout := bufio.NewWriter(out)
	defer bufout.Flush()

	// Start JSON array
	bufout.WriteString("[")
	isFirst := true

	// Write each document's metadata as JSON object
	for _, url := range engine.urlStorage.getAllUrls() {
		rr, exists := engine.docStorage[url.String()]
		if exists {
			if !isFirst {
				bufout.WriteString(",")
			}
			isFirst = false
			_ = rr.OutJSON(bufout)
		}
	}

	// Close JSON array
	bufout.WriteString("]")

	return nil
}
