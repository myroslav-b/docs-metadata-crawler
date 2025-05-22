package main

import (
	"fmt"
	"net/url"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUrlStorage(t *testing.T) {
	storage := newUrlStorage()
	require.NotNil(t, storage, "Storage should be initialized")

	t.Run("Add and check URLs", func(t *testing.T) {
		url1, err := url.Parse("https://example.com/page1")
		require.NoError(t, err)

		// Add first URL
		added := storage.add(url1)
		assert.True(t, added, "URL should be added successfully")

		// Check if URL exists and is not used
		exists, used := storage.check(url1)
		assert.True(t, exists, "URL should exist in storage")
		assert.False(t, used, "URL should not be marked as used")

		// Try adding same URL again
		added = storage.add(url1)
		assert.False(t, added, "Same URL should not be added twice")

		// Add nil URL
		added = storage.add(nil)
		assert.False(t, added, "Nil URL should not be added")
	})

	t.Run("Use URLs", func(t *testing.T) {
		storage := newUrlStorage()

		// Add multiple URLs
		url1, _ := url.Parse("https://example.com/doc1.pdf")
		url2, _ := url.Parse("https://example.com/doc2.pdf")
		url3, _ := url.Parse("https://example.com/doc3.pdf")

		storage.add(url1)
		storage.add(url2)
		storage.add(url3)

		// Use first URL
		usedUrl, ok := storage.use()
		assert.True(t, ok, "Should return a URL")
		assert.NotNil(t, usedUrl, "URL should not be nil")

		// Check if URL is marked as used
		exists, used := storage.check(usedUrl)
		assert.True(t, exists, "URL should exist")
		assert.True(t, used, "URL should be marked as used")

		// Use all URLs
		storage.use() // second URL
		storage.use() // third URL

		// Try to use when no URLs left
		usedUrl, ok = storage.use()
		assert.False(t, ok, "Should return false when no URLs left")
		assert.Nil(t, usedUrl, "URL should be nil when no URLs left")
	})

	t.Run("Get all URLs", func(t *testing.T) {
		storage := newUrlStorage()

		// Add URLs
		url1, _ := url.Parse("https://example.com/path1")
		url2, _ := url.Parse("https://example.com/path2")

		storage.add(url1)
		storage.add(url2)

		// Get all URLs
		urls := storage.getAllUrls()
		assert.Len(t, urls, 2, "Should return 2 URLs")

		// Verify URLs are in the list
		foundUrl1 := false
		foundUrl2 := false

		for _, u := range urls {
			if u.String() == url1.String() {
				foundUrl1 = true
			}
			if u.String() == url2.String() {
				foundUrl2 = true
			}
		}

		assert.True(t, foundUrl1, "URL1 should be in the list")
		assert.True(t, foundUrl2, "URL2 should be in the list")
	})

	t.Run("Count URLs", func(t *testing.T) {
		storage := newUrlStorage()

		// Initial count
		total, used := storage.count()
		assert.Equal(t, 0, total, "Initial total count should be 0")
		assert.Equal(t, 0, used, "Initial used count should be 0")

		// Add URLs
		url1, _ := url.Parse("https://example.com/doc1")
		url2, _ := url.Parse("https://example.com/doc2")
		storage.add(url1)
		storage.add(url2)

		// Count after adding
		total, used = storage.count()
		assert.Equal(t, 2, total, "Total count should be 2")
		assert.Equal(t, 0, used, "Used count should be 0")

		// Use one URL
		storage.use()

		// Count after using
		total, used = storage.count()
		assert.Equal(t, 2, total, "Total count should remain 2")
		assert.Equal(t, 1, used, "Used count should be 1")
	})
}

func TestUrlStorage_Add_Concurrency(t *testing.T) {
	us := newUrlStorage()
	numGoroutines := 100
	numUrlsPerGoroutine := 10

	var wg sync.WaitGroup
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(goroutineID int) {
			defer wg.Done()
			for j := 0; j < numUrlsPerGoroutine; j++ {
				uStr := fmt.Sprintf("http://example.com/concurrent/%d/%d", goroutineID, j)
				u, err := url.Parse(uStr)
				require.NoError(t, err)
				us.add(u)
			}
		}(i)
	}
	wg.Wait()

	// Перевірка кількості доданих URL
	expectedTotal := numGoroutines * numUrlsPerGoroutine
	total, used := us.count()
	assert.Equal(t, expectedTotal, total, "Concurrent adds should result in correct total count")
	assert.Equal(t, 0, used, "No URLs should be marked as used")

	// Перевірка унікальності доданих URL
	allUrls := us.getAllUrls()
	urlSet := make(map[string]bool)
	for _, u := range allUrls {
		urlStr := u.String()
		assert.False(t, urlSet[urlStr], "URL %s was duplicated", urlStr)
		urlSet[urlStr] = true
	}
	assert.Equal(t, expectedTotal, len(urlSet), "All URLs should be unique")
}

func TestUrlStorage_Use_Concurrency(t *testing.T) {
	us := newUrlStorage()
	numUrls := 200

	// Додаємо URL
	for i := 0; i < numUrls; i++ {
		uStr := fmt.Sprintf("http://example.com/use_concurrent/%d", i)
		u, err := url.Parse(uStr)
		require.NoError(t, err)
		added := us.add(u)
		assert.True(t, added, "URL should be added successfully")
	}

	// Перевіряємо, що всі URL додані
	total, used := us.count()
	assert.Equal(t, numUrls, total, "All URLs should be added")
	assert.Equal(t, 0, used, "No URLs should be used initially")

	// Паралельно використовуємо URL
	numGoroutines := 50
	var wg sync.WaitGroup
	usedUrls := sync.Map{}

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				u, ok := us.use()
				if !ok {
					break // Немає більше URL
				}
				if u != nil {
					// Перевіряємо, що URL не був використаний раніше
					_, loaded := usedUrls.LoadOrStore(u.String(), true)
					assert.False(t, loaded, "URL %s was used more than once", u.String())
				}
			}
		}()
	}
	wg.Wait()

	// Перевіряємо, що всі URL використані
	total, remainingUsed := us.count()
	assert.Equal(t, numUrls, total, "Total URL count should remain the same")
	assert.Equal(t, numUrls, remainingUsed, "All URLs should be marked as used")

	// Перевіряємо, що немає доступних URL
	_, ok := us.use()
	assert.False(t, ok, "No more URLs should be available after concurrent use")

	// Перевіряємо, що кількість унікальних використаних URL відповідає очікуванню
	count := 0
	usedUrls.Range(func(_, _ interface{}) bool {
		count++
		return true
	})
	assert.Equal(t, numUrls, count, "The number of unique used URLs should match")
}

// Додатковий тест на паралельні операції читання-запису
func TestUrlStorage_ConcurrentReadWrite(t *testing.T) {
	us := newUrlStorage()
	numUrls := 100

	// Додаємо початкові URL
	for i := 0; i < numUrls/2; i++ {
		uStr := fmt.Sprintf("http://example.com/readwrite/%d", i)
		u, err := url.Parse(uStr)
		require.NoError(t, err)
		us.add(u)
	}

	var wg sync.WaitGroup
	// Паралельно додаємо URL
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := numUrls / 2; i < numUrls; i++ {
			uStr := fmt.Sprintf("http://example.com/readwrite/%d", i)
			u, err := url.Parse(uStr)
			require.NoError(t, err)
			us.add(u)
			time.Sleep(time.Millisecond) // Невелика затримка для збільшення шансу паралельного виконання
		}
	}()

	// Паралельно читаємо стан URL
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 50; i++ {
			total, used := us.count()
			assert.True(t, total >= numUrls/2, "Should have at least initial URLs")
			assert.True(t, used <= total, "Used count should not exceed total")
			time.Sleep(time.Millisecond)
		}
	}()

	// Паралельно отримуємо всі URL
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 20; i++ {
			urls := us.getAllUrls()
			assert.True(t, len(urls) >= numUrls/2, "Should return at least initial URLs")
			time.Sleep(2 * time.Millisecond)
		}
	}()

	// Паралельно використовуємо URL
	wg.Add(1)
	go func() {
		defer wg.Done()
		usedCount := 0
		for {
			u, ok := us.use()
			if !ok || usedCount >= numUrls/2 {
				break
			}
			if u != nil {
				usedCount++
			}
			time.Sleep(time.Millisecond)
		}
	}()

	wg.Wait()

	// Перевіряємо фінальний стан
	total, used := us.count()
	assert.Equal(t, numUrls, total, "Should have all URLs added")
	assert.True(t, used > 0, "Some URLs should be used")
	assert.True(t, used < numUrls, "Not all URLs should be used")
}
