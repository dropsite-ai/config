package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/dropsite-ai/config"
	"gopkg.in/yaml.v2"
)

// usage prints the CLI usage.
func usage() {
	fmt.Println(`Usage: config <command> [options]

Commands:
  load      Load a YAML config from file (creates file if not exists)
              -file string   Path to config file

  save      Save/update a YAML config file.
              -file string    Path to config file
              -update string  Update a field in the form key=value (can be repeated)

  process   Process a YAML config from file and print the processed result
              -file string   Path to config file

  generate  Generate a JWT secret and print it

  expand    Expand a given path (handles "~")
              -path string   Path to expand

  validate  Validate a value as username or URL
              -type string   "username" or "url"
              -value string  Value to validate

  copy      Copy a property from one YAML config to another
              -src string      Source config file
              -srcField string Source field name
              -dst string      Destination config file
              -dstField string Destination field name`)
	os.Exit(1)
}

// updateFlag is a custom flag type that collects repeated -update flags.
type updateFlag []string

func (u *updateFlag) String() string {
	return fmt.Sprintf("%v", *u)
}

func (u *updateFlag) Set(value string) error {
	*u = append(*u, value)
	return nil
}

func main() {
	if len(os.Args) < 2 {
		usage()
	}

	command := os.Args[1]
	switch command {
	case "load":
		loadCmd(os.Args[2:])
	case "save":
		saveCmd(os.Args[2:])
	case "process":
		processCmd(os.Args[2:])
	case "generate":
		generateCmd(os.Args[2:])
	case "expand":
		expandCmd(os.Args[2:])
	case "validate":
		validateCmd(os.Args[2:])
	case "copy":
		copyCmd(os.Args[2:])
	default:
		usage()
	}
}

// loadCmd loads a YAML configuration file (creating it if necessary)
// and prints the configuration.
func loadCmd(args []string) {
	fs := flag.NewFlagSet("load", flag.ExitOnError)
	file := fs.String("file", "", "Path to config file")
	fs.Parse(args)
	if *file == "" {
		fmt.Println("Please provide -file")
		fs.Usage()
		os.Exit(1)
	}
	// For a generic config, use a map as default.
	defaultCfg := make(map[string]interface{})
	cfg, err := config.Load(*file, defaultCfg)
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}
	data, err := yaml.Marshal(cfg)
	if err != nil {
		log.Fatalf("Error marshalling config: %v", err)
	}
	fmt.Println(string(data))
}

// saveCmd loads an existing YAML config (or creates a default empty one),
// applies one or more field updates (using -update key=value flags),
// processes the config, and saves it.
func saveCmd(args []string) {
	fs := flag.NewFlagSet("save", flag.ExitOnError)
	file := fs.String("file", "", "Path to config file")
	var updates updateFlag
	fs.Var(&updates, "update", "Update a field in the form key=value (can be repeated)")
	fs.Parse(args)
	if *file == "" {
		fmt.Println("Please provide -file")
		fs.Usage()
		os.Exit(1)
	}

	// Default config is an empty map.
	defaultCfg := make(map[string]interface{})
	cfg, err := config.Load(*file, defaultCfg)
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	// Process each update flag.
	for _, upd := range updates {
		parts := strings.SplitN(upd, "=", 2)
		if len(parts) != 2 {
			log.Fatalf("Invalid update format %q. Expected key=value", upd)
		}
		key, value := parts[0], parts[1]
		cfg[key] = value
	}

	// Process the config (if any processing is defined, else this is a no-op for maps).
	if err := config.Process(&cfg); err != nil {
		log.Fatalf("Error processing config: %v", err)
	}
	if err := config.Save(*file, cfg); err != nil {
		log.Fatalf("Error saving config: %v", err)
	}
	fmt.Printf("Config saved to %s\n", *file)
}

// processCmd loads and processes a config file then prints the processed config.
func processCmd(args []string) {
	fs := flag.NewFlagSet("process", flag.ExitOnError)
	file := fs.String("file", "", "Path to config file")
	fs.Parse(args)
	if *file == "" {
		fmt.Println("Please provide -file")
		fs.Usage()
		os.Exit(1)
	}
	defaultCfg := make(map[string]interface{})
	cfg, err := config.Load(*file, defaultCfg)
	if err != nil {
		log.Fatalf("Error processing config: %v", err)
	}
	data, err := yaml.Marshal(cfg)
	if err != nil {
		log.Fatalf("Error marshalling config: %v", err)
	}
	fmt.Println(string(data))
}

// generateCmd generates a new JWT secret and prints it.
func generateCmd(args []string) {
	fs := flag.NewFlagSet("generate", flag.ExitOnError)
	fs.Parse(args)
	secret, err := config.GenerateJWTSecret()
	if err != nil {
		log.Fatalf("Error generating JWT secret: %v", err)
	}
	fmt.Println(secret)
}

// expandCmd expands a provided path (e.g. handling "~") and prints the result.
func expandCmd(args []string) {
	fs := flag.NewFlagSet("expand", flag.ExitOnError)
	pathStr := fs.String("path", "", "Path to expand")
	fs.Parse(args)
	if *pathStr == "" {
		fmt.Println("Please provide -path")
		fs.Usage()
		os.Exit(1)
	}
	expanded, err := config.ExpandPath(*pathStr)
	if err != nil {
		log.Fatalf("Error expanding path: %v", err)
	}
	fmt.Println(expanded)
}

// validateCmd validates a value as either a username or a URL.
func validateCmd(args []string) {
	fs := flag.NewFlagSet("validate", flag.ExitOnError)
	typ := fs.String("type", "", "Type to validate: 'username' or 'url'")
	value := fs.String("value", "", "Value to validate")
	fs.Parse(args)
	if *typ == "" || *value == "" {
		fmt.Println("Please provide -type and -value")
		fs.Usage()
		os.Exit(1)
	}
	switch *typ {
	case "username":
		if err := config.ValidateUsername(*value); err != nil {
			log.Fatalf("Invalid username: %v", err)
		}
		fmt.Println("Username is valid")
	case "url":
		if err := config.ValidateURL(*value); err != nil {
			log.Fatalf("Invalid URL: %v", err)
		}
		fmt.Println("URL is valid")
	default:
		fmt.Println("Unknown type. Use 'username' or 'url'")
		fs.Usage()
		os.Exit(1)
	}
}

// copyCmd copies a property from one YAML config to another.
// Because the generic config is a map, we handle this without the struct-based CopyProperty.
func copyCmd(args []string) {
	fs := flag.NewFlagSet("copy", flag.ExitOnError)
	srcFile := fs.String("src", "", "Source config file")
	srcField := fs.String("srcField", "", "Source field name")
	dstFile := fs.String("dst", "", "Destination config file")
	dstField := fs.String("dstField", "", "Destination field name")
	fs.Parse(args)
	if *srcFile == "" || *srcField == "" || *dstFile == "" || *dstField == "" {
		fmt.Println("Please provide -src, -srcField, -dst, and -dstField")
		fs.Usage()
		os.Exit(1)
	}
	defaultCfg := make(map[string]interface{})
	srcCfg, err := config.Load(*srcFile, defaultCfg)
	if err != nil {
		log.Fatalf("Error loading source config: %v", err)
	}
	dstCfg, err := config.Load(*dstFile, defaultCfg)
	if err != nil {
		log.Fatalf("Error loading destination config: %v", err)
	}

	val, ok := srcCfg[*srcField]
	if !ok {
		log.Fatalf("Field %q not found in source config", *srcField)
	}
	dstCfg[*dstField] = val

	if err := config.Save(*dstFile, dstCfg); err != nil {
		log.Fatalf("Error saving destination config: %v", err)
	}
	fmt.Printf("Copied field %q from %s to field %q in %s\n", *srcField, *srcFile, *dstField, *dstFile)
}
