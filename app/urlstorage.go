package main

import (
	"net/url"
	"sync"
)

// tUrlStorage manages URL collection, status tracking, and processing queue
// with thread-safe operations using RWMutex for concurrent access control
type tUrlStorage struct {
	mu         sync.RWMutex        // RWMutex for concurrent access control
	urlStatus  map[string]bool     // URL status map (true = used/processed)
	urlObjects map[string]*url.URL // Map of string keys to URL objects
	queue      []string            // Queue of URLs to be processed
}

// newUrlStorage creates and initializes a new URL storage instance
func newUrlStorage() *tUrlStorage {
	return &tUrlStorage{
		urlStatus:  make(map[string]bool),
		urlObjects: make(map[string]*url.URL),
		queue:      make([]string, 0, 100),
	}
}

// Add adds a new URL to the storage if it doesn't already exist
// Returns true if URL was added, false if it already existed or is nil
func (us *tUrlStorage) add(u *url.URL) bool {
	if u == nil {
		return false
	}

	us.mu.Lock()
	defer us.mu.Unlock()

	key := u.String()

	// Check if URL already exists
	if _, exists := us.urlStatus[key]; exists {
		return false
	}

	// Store a copy of the URL
	urlCopy := *u // Create a copy of the URL structure
	us.urlObjects[key] = &urlCopy
	us.urlStatus[key] = false // false = unused
	us.queue = append(us.queue, key)

	return true
}

// Use returns an unused URL and marks it as used
// Returns the URL and true if successful, nil and false if no unused URLs exist
func (us *tUrlStorage) use() (*url.URL, bool) {
	us.mu.Lock()
	defer us.mu.Unlock()

	// Find an unused URL in the queue
	for i := 0; i < len(us.queue); i++ {
		key := us.queue[i]

		if !us.urlStatus[key] {
			// Mark as used
			us.urlStatus[key] = true

			// Remove from queue (fast removal without preserving order)
			us.queue[i] = us.queue[len(us.queue)-1]
			us.queue = us.queue[:len(us.queue)-1]

			return us.urlObjects[key], true
		}
	}

	return nil, false
}

// GetAllURLs returns all URLs stored in the storage
func (us *tUrlStorage) getAllUrls() []*url.URL {
	us.mu.RLock()
	defer us.mu.RUnlock()

	result := make([]*url.URL, 0, len(us.urlObjects))

	for _, urlObj := range us.urlObjects {
		// Important: return the stored pointers, not creating new ones
		result = append(result, urlObj)
	}

	return result
}

// Check verifies if a URL exists in storage and whether it's already used
// Returns (exists, used) as booleans
func (us *tUrlStorage) check(u *url.URL) (exists bool, used bool) {
	if u == nil {
		return false, false
	}

	us.mu.RLock()
	defer us.mu.RUnlock()

	key := u.String()
	used, exists = us.urlStatus[key]
	return exists, used
}

// Count returns the total number of URLs in storage and how many are used
func (us *tUrlStorage) count() (total int, used int) {
	us.mu.RLock()
	defer us.mu.RUnlock()

	total = len(us.urlStatus)

	for _, isUsed := range us.urlStatus {
		if isUsed {
			used++
		}
	}

	return total, used
}
