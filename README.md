# config

Go package and CLI tool for processing YAML configuration files for [LLMFS](https://github.com/dropsite-ai/llmfs).

## Format

The configuration file is written in YAML and may have any fields with two optional sections: **variables** and **callbacks**.

### Variables

Under the `variables` key, you define mappings for endpoints, secrets, users, and paths.

- **endpoints:**  
  A mapping of endpoint names to their corresponding URL strings. Each URL must include a valid scheme (like `http` or `https`) and a host.

  ```yaml
  endpoints:
    service1: "http://example.com"
  ```

- **secrets:**  
  A mapping of secret names to their values. If a secret is left empty (`""`), the loader automatically generates a new secret (a 64-character hexadecimal string).  
 
  ```yaml
  secrets:
    secret1: ""          # This is replaced with a generated secret.
    secret2: "mysecret"  # This secret remains unchanged.
  ```

- **users:**  
  A mapping of user keys to their usernames. Usernames are validated using Linux-style naming rules (must start with a lowercase letter or underscore, and contain only lowercase letters, numbers, underscores, or dashes; up to 32 characters).

  ```yaml
  users:
    user1: "root"
  ```

- **paths:**  
  A mapping of path keys to filesystem paths. Paths starting with a tilde (`~`) will be automatically expanded to the current user’s home directory.

  ```yaml
  paths:
    path1: "/.llmfs.yml"
    path2: "~/folder"
  ```

### Callbacks

The `callbacks` section defines an array of callback definitions. Each callback must include the following fields:

- **name:**  
  A unique identifier for the callback.

- **events:**  
  A list of event names that trigger the callback.

- **timing:**  
  Specifies when the callback runs. Only two values are allowed: `"pre"` or `"post"`.

- **target:**  
  A mapping that describes the callback’s target. It includes:
  
  - **type:** The target type, which can be either `"file"` or `"directory"`.
  - **path:** The filesystem path to the target.

- **endpoints:**  
  A list of endpoint keys (defined under `variables.endpoints`) associated with the callback.

Example callback configuration:

```yaml
callbacks:
  - name: "callback1"
    events: ["event1", "event2"]
    timing: "pre"
    target:
      type: "file"
      path: "some/path"
    endpoints: ["service1"]
```

### Overall Structure

A complete configuration file might look like this:

```yaml
variables:
  endpoints:
    service1: "http://example.com"
  secrets:
    secret1: ""
    secret2: "existingsecret"
  users:
    user1: "root"
  paths:
    path1: "~"
    path2: "~/folder"

callbacks:
  - name: "callback1"
    events: ["event1", "event2"]
    timing: "pre"
    target:
      type: "file"
      path: "some/path"
    endpoints: ["service1"]
```

## Installation

### Go Package

```bash
go get github.com/dropsite-ai/config
```

### Homebrew (macOS or Compatible)

```bash
brew tap dropsite-ai/homebrew-tap
brew install config
```

### Download Binaries

Grab the latest pre-built binaries from the [GitHub Releases](https://github.com/dropsite-ai/config/releases). Extract them, then run the `config` executable directly.

### Build from Source

```bash
git clone https://github.com/dropsite-ai/config.git
cd config
go build -o config cmd/main.go
```

## Programmatic Usage

You can use the package directly in your Go code to load and process YAML configuration files. For example:

```go
package main

import (
	"fmt"
	"log"

	"github.com/dropsite-ai/config"
)

func main() {
	// Load configuration from a YAML file.
	doc, vars, callbacks, err := config.Load("path/to/config.yaml")
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	// Display the processed configuration.
	fmt.Printf("Processed Variables:\n%+v\n", vars)
	fmt.Printf("Processed Callbacks:\n%+v\n", callbacks)

	// Save the updated document back to the file.
	if err := config.Save("path/to/config.yaml", doc); err != nil {
		log.Fatalf("Error saving config: %v", err)
	}
}
```

## CLI Usage

The CLI tool (`config`) provides two main commands: `load` and `copy`.

### Load Command

Loads a configuration file, processes it, and displays the YAML document along with the processed variables and callbacks.

#### Example

```bash
config load -config path/to/config.yaml
```

**Output:**  
- The complete YAML document (including any modifications, such as generated secrets).  
- Processed variables and callback definitions printed to the console.

### Copy Command

Copies a value from one YAML file to another using dot-notation to specify the source and destination fields.

#### Example

```bash
config copy -srcfile src.yaml -srcpath variables.secrets.secret1 \
            -dstfile dst.yaml -dstpath config.secret
```

**Flags:**

- `-srcfile`: Path to the source YAML file.
- `-srcpath`: Dot-notation path to the field in the source YAML.
- `-dstfile`: Path to the destination YAML file.
- `-dstpath`: Dot-notation path where the value should be written in the destination YAML.

This command reads the specified field from the source, updates the destination YAML file at the given path, and writes the changes back to disk.

## License

This project is licensed under the [MIT License](LICENSE).