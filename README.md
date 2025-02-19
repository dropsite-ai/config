# config

A flexible YAML configuration utility for Go, with optional CLI usage. This repository provides:

- **Programmatic usage** to load, save, process, and validate YAML configs within your Go code.
- A **CLI** tool (`config`) for simple config file manipulation.

---

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

---

## Programmatic Usage

This package is designed to simplify reading and writing YAML configuration files in Go, with additional logic for secrets, path expansion, username validation, and URL validation.

### Importing

```go
import "github.com/dropsite-ai/config"
```

### Loading and Saving Configs

#### Defining Your Config Struct

To take advantage of automatic processing, structure your configuration to include a dedicated `variables` section. Within this section you can define maps for endpoints, secrets, users, and paths:

```go
type Variables struct {
    Endpoints map[string]string `yaml:"endpoints"`
    Secrets   map[string]string `yaml:"secrets"`
    Users     map[string]string `yaml:"users"`
    Paths     map[string]string `yaml:"paths"`
}

type MyAppConfig struct {
    Variables Variables `yaml:"variables"`
    LogLevel  string    `yaml:"logLevel"`
}
```

#### YAML Configuration Example

```yaml
variables:
  endpoints:
    user: http://localhost:9000/callback
  secrets:
    root: ""
  users:
    owner: root
  paths:
    database: ~/llmfs.db
logLevel: debug
```

In this configuration:

- **endpoints:** Each value is validated as a well-formed URL.
- **secrets:** Any empty secret value is replaced with a newly generated 32-byte JWT secret (hex-encoded).
- **users:** Each value is validated as a Linux-style username.
- **paths:** Each value is expanded (e.g. a leading `~` is replaced with the user’s home directory).

> **Note:** If your configuration does not include a `variables` section, then these automatic validations and transformations are not applied. You can still use helper functions (such as `ExpandPath` or `ValidateUsername`) manually.

#### Loading a Config File

```go
// Default config if file doesn't exist:
defaultCfg := MyAppConfig{
    LogLevel: "info",
    Variables: Variables{
        Endpoints: map[string]string{"user": "http://localhost:9000/callback"},
        Secrets:   map[string]string{"root": ""},
        Users:     map[string]string{"owner": "root"},
        Paths:     map[string]string{"database": "~/llmfs.db"},
    },
}

// Load or create the YAML file:
cfg, err := config.Load("config.yaml", defaultCfg)
if err != nil {
    panic(err)
}
```

- If `config.yaml` does **not** exist, `Load` will:
  1. **Process** your `defaultCfg` (auto-generating secrets, expanding paths, and validating endpoints and usernames).
  2. **Save** that as a brand-new `config.yaml`.
  3. Return the processed `defaultCfg`.

- If `config.yaml` **does** exist, `Load` will:
  1. Read and unmarshal the existing YAML.
  2. **Process** it (auto-generating secrets if empty, expanding paths, and validating endpoints and usernames).
  3. Return the processed config.

#### Saving a Config File

After modifying your config in code, you can save it:

```go
err = config.Save("config.yaml", cfg)
if err != nil {
    panic(err)
}
```

### Processing Config Fields

When you call `Load` (or manually invoke `config.Process`), the package looks for a `variables` block in your configuration. If found, it processes the following sub-sections:

- **endpoints:**  
  Each value is validated as a well-formed URL (the URL must have both a scheme and host).

- **secrets:**  
  If a secret value is empty, a new 32-byte JWT secret is generated and inserted.

- **users:**  
  Each value is validated as a Linux-style username.

- **paths:**  
  Each value is expanded, for example replacing a leading `~` with the user’s home directory.

### Other Helpers

- **`GenerateJWTSecret()`**  
  Creates a 32-byte secure random key and returns it as a hex-encoded string.
  
- **`ExpandPath(path string)`**  
  Expands a leading `~` to the user’s home directory (e.g. `"~/myapp"` becomes `"/Users/joe/myapp"`).
  
- **`ValidateUsername(username string)`**  
  Validates the username against Linux-style rules. Returns an error if the username is invalid.
  
- **`ValidateURL(url string)`**  
  Checks that the string is a valid URL with both a scheme and host.
  
- **`CopyProperty(src, srcField, dst, dstField)`**  
  Copies a property from one configuration to another. This function supports nested fields using dot‑notation (for example, `"Variables.Secrets.api"`).

---

## CLI Usage

The `config` CLI mirrors much of the functionality available in the package. You can see usage by simply running:

```bash
$ config -h
Usage: config <command> [options]

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
              -srcField string Source field name (supports nested dot‑notation)
              -dst string      Destination config file
              -dstField string Destination field name (supports nested dot‑notation)
```

---

## Processing via `variables` Section

This package processes configuration fields using a dedicated `variables` section. Your configuration should include a block like the following:

```yaml
variables:
  endpoints:
    user: http://localhost:9000/callback
  secrets:
    root: ""
  users:
    owner: root
  paths:
    database: ~/llmfs.db
```

Within this section:

- **endpoints:** Each value is validated as a well-formed URL.
- **secrets:** Empty secret values are replaced with a generated 32-byte JWT secret (hex-encoded).
- **users:** Each value is validated as a Linux-style username.
- **paths:** Each value is expanded (for example, a leading `~` is replaced with the user’s home directory).

These processing rules are applied automatically when you call `Load` or `Process`.

---

## License

[MIT](LICENSE)