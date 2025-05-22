package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOptsParsing(t *testing.T) {
	// Test the defaults in the opts structure
	opts := tOpts{}

	// Assert default values
	assert.Equal(t, "", opts.Site, "Default Site should be empty string")
	assert.Empty(t, opts.Type, "Default Type should be empty slice")
	assert.Equal(t, "", opts.Output, "Default Output should be empty string")
	assert.Equal(t, 0, opts.Paramax, "Default Paramax should be 0")

	// Test with values
	opts = tOpts{
		Site:    "https://example.com",
		Type:    []string{"pdf", "docx"},
		Output:  "output.json",
		Paramax: 10,
	}

	// Verify values
	assert.Equal(t, "https://example.com", opts.Site)
	assert.Equal(t, []string{"pdf", "docx"}, opts.Type)
	assert.Equal(t, "output.json", opts.Output)
	assert.Equal(t, 10, opts.Paramax)
}

// Note: Testing the main function directly is challenging because it calls os.Exit()
// A more comprehensive test would involve capturing command line arguments and
// redirecting them to the parser. That would be more of an integration test.
