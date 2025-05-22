package main

import (
	"log"
	"os"

	"github.com/jessevdk/go-flags"
)

// tOpts defines command line options for the document crawler
// Uses go-flags package for parsing and validation
type tOpts struct {
	Site    string   `short:"s" long:"site" required:"true" description:"site name"`
	Type    []string `short:"t" long:"type" choice:"pdf" choice:"docx" choice:"xlsx" choice:"pptx" description:"document type / file name extension (all if empty)"`
	Output  string   `short:"o" long:"output" default:"" description:"output stream, stdout if none"`
	Paramax int      `short:"p" long:"paramax" default:"100" description:"maximum number of parallel analysis threads"`
}

// main is the entry point of the application
// Parses command line arguments and starts the crawling engine
func main() {
	var opts tOpts

	// Initialize command line parser
	parser := flags.NewParser(&opts, flags.Default)

	// Parse command line arguments
	if _, err := parser.Parse(); err != nil {
		os.Exit(1)
	}

	// If no document types are specified, use all supported types
	typeOption := parser.FindOptionByLongName("type")
	allowDocTypes := typeOption.Choices
	if len(opts.Type) == 0 {
		opts.Type = allowDocTypes
	}

	// Initialize and run the crawler engine
	engine, err := newEngine(opts)
	if err != nil {
		log.Fatalf("Engine initialization error: %v", err)
	}

	engine.run()
}
