package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/dropsite-ai/config"
	"github.com/dropsite-ai/yamledit"
)

func usage() {
	fmt.Println("Usage:")
	fmt.Println("  cli load -config <path>")
	fmt.Println("  cli copy -srcfile <src.yaml> -srcpath <dot.path> -dstfile <dst.yaml> -dstpath <dot.path>")
}

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(1)
	}
	cmd := os.Args[1]
	switch cmd {
	case "load":
		loadCmd(os.Args[2:])
	case "copy":
		copyCmd(os.Args[2:])
	default:
		usage()
		os.Exit(1)
	}
}

// loadCmd loads a config file, processes it, and then displays the YAML document along with
// the processed Variables and CallbackDefinition values.
func loadCmd(args []string) {
	fs := flag.NewFlagSet("load", flag.ExitOnError)
	configPath := fs.String("config", "", "Path to YAML config file")
	fs.Parse(args)
	if *configPath == "" {
		fs.Usage()
		os.Exit(1)
	}

	doc, vars, callbacks, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	// Encode the YAML document back into bytes for display.
	yamlBytes, err := yamledit.Encode(doc)
	if err != nil {
		log.Fatalf("Error encoding YAML: %v", err)
	}
	fmt.Println("YAML Document:")
	fmt.Println(string(yamlBytes))
	fmt.Println("\nProcessed Variables:")
	fmt.Printf("%+v\n", vars)
	fmt.Println("\nProcessed Callbacks:")
	fmt.Printf("%+v\n", callbacks)
}

// copyCmd copies a field from a source YAML file to a destination YAML file based on dot-notation paths.
func copyCmd(args []string) {
	fs := flag.NewFlagSet("copy", flag.ExitOnError)
	srcFile := fs.String("srcfile", "", "Source YAML file")
	srcPath := fs.String("srcpath", "", "Dot-notation path in source YAML")
	dstFile := fs.String("dstfile", "", "Destination YAML file")
	dstPath := fs.String("dstpath", "", "Dot-notation path in destination YAML")
	fs.Parse(args)
	if *srcFile == "" || *srcPath == "" || *dstFile == "" || *dstPath == "" {
		fs.Usage()
		os.Exit(1)
	}

	// Load the source YAML file.
	srcBytes, err := os.ReadFile(*srcFile)
	if err != nil {
		log.Fatalf("Error reading source file: %v", err)
	}
	srcDoc, err := yamledit.Parse(srcBytes)
	if err != nil {
		log.Fatalf("Error parsing source YAML: %v", err)
	}

	// Read the value at the given dot-notation path from the source.
	var value interface{}
	if err = yamledit.ReadNode(srcDoc, *srcPath, &value); err != nil {
		log.Fatalf("Error reading field %q from source: %v", *srcPath, err)
	}

	// Load the destination YAML file.
	dstBytes, err := os.ReadFile(*dstFile)
	if err != nil {
		log.Fatalf("Error reading destination file: %v", err)
	}
	dstDoc, err := yamledit.Parse(dstBytes)
	if err != nil {
		log.Fatalf("Error parsing destination YAML: %v", err)
	}

	// Update the destination document with the value from the source.
	if err = yamledit.UpdateNode(dstDoc, *dstPath, value); err != nil {
		log.Fatalf("Error updating destination YAML at %q: %v", *dstPath, err)
	}

	// Encode and write the updated destination YAML back to file.
	updatedBytes, err := yamledit.Encode(dstDoc)
	if err != nil {
		log.Fatalf("Error encoding updated destination YAML: %v", err)
	}
	if err := os.WriteFile(*dstFile, updatedBytes, 0644); err != nil {
		log.Fatalf("Error writing updated destination file: %v", err)
	}

	fmt.Printf("Successfully copied field %q from %q to field %q in %q\n", *srcPath, *srcFile, *dstPath, *dstFile)
}
