package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/dropsite-ai/config"
	"gopkg.in/yaml.v2"
)

// Config is an example configuration struct used by the CLI.
type Config struct {
	Secret string `yaml:"secret"`
	Path   string `yaml:"path"`
	User   string `yaml:"user"`
	URL    string `yaml:"url"`
}

func usage() {
	fmt.Println(`Usage: llmfs-config <command> [options]

Commands:
  load      Load a YAML config from file (creates file if not exists)
            Options:
              -file string   Path to config file

  save      Save a YAML config to file
            Options:
              -file string   Path to config file
              -secret string Secret value
              -path string   Path value (will be expanded)
              -user string   Username (validated)
              -url string    URL (validated)

  process   Process a YAML config from file and print the processed result
            Options:
              -file string   Path to config file

  generate  Generate a JWT secret and print it

  expand    Expand a given path (handles "~")
            Options:
              -path string   Path to expand

  validate  Validate a value as username or URL
            Options:
              -type string   "username" or "url"
              -value string  Value to validate

  copy      Copy a property from one YAML config to another
            Options:
              -src string      Source config file
              -srcField string Source field name
              -dst string      Destination config file
              -dstField string Destination field name`)
}

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(1)
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
		os.Exit(1)
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
	// Define a default config.
	defaultCfg := Config{
		Secret: "",
		Path:   "~/default/path",
		User:   "defaultuser",
		URL:    "http://example.com",
	}
	cfg, err := config.Load[Config](*file, defaultCfg)
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}
	data, err := yaml.Marshal(cfg)
	if err != nil {
		log.Fatalf("Error marshalling config: %v", err)
	}
	fmt.Println(string(data))
}

// saveCmd saves a configuration provided via flags to a YAML file.
func saveCmd(args []string) {
	fs := flag.NewFlagSet("save", flag.ExitOnError)
	file := fs.String("file", "", "Path to config file")
	secret := fs.String("secret", "", "Secret value")
	pathVal := fs.String("path", "", "Path value")
	user := fs.String("user", "", "Username")
	urlVal := fs.String("url", "", "URL")
	fs.Parse(args)
	if *file == "" {
		fmt.Println("Please provide -file")
		fs.Usage()
		os.Exit(1)
	}
	cfg := Config{
		Secret: *secret,
		Path:   *pathVal,
		User:   *user,
		URL:    *urlVal,
	}
	// Process the config to apply custom logic (expanding paths, generating secrets, etc.).
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
	// Use the same default as for load.
	defaultCfg := Config{
		Secret: "",
		Path:   "~/default/path",
		User:   "defaultuser",
		URL:    "http://example.com",
	}
	cfg, err := config.Load[Config](*file, defaultCfg)
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
	// Load source config.
	defaultSrc := Config{}
	srcCfg, err := config.Load[Config](*srcFile, defaultSrc)
	if err != nil {
		log.Fatalf("Error loading source config: %v", err)
	}
	// Load destination config.
	defaultDst := Config{}
	dstCfg, err := config.Load[Config](*dstFile, defaultDst)
	if err != nil {
		log.Fatalf("Error loading destination config: %v", err)
	}
	// Copy the property.
	if err := config.CopyProperty(&srcCfg, *srcField, &dstCfg, *dstField); err != nil {
		log.Fatalf("Error copying property: %v", err)
	}
	// Save the updated destination config.
	if err := config.Save(*dstFile, dstCfg); err != nil {
		log.Fatalf("Error saving destination config: %v", err)
	}
	fmt.Printf("Copied field %q from %s to field %q in %s\n", *srcField, *srcFile, *dstField, *dstFile)
}
