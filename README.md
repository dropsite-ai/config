# config

A flexible YAML configuration utility for Go, with optional CLI usage. This repository provides:

- **Programmatic usage** to load, save, process, and validate YAML configs within your Go code.
- A **CLI** tool (`config`) for simple config file manipulation.

---

## Installation

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

You can define a struct for your config. Any field ending with `Secret`, `Path`, `User`, or `URL` receives special processing:

- **Secret**: If empty, generates a new random JWT secret (hex-encoded).
- **Path**: Expands `~` to the user’s home directory.
- **User**: Validates it against a simple Linux-style username rule.
- **URL**: Validates the string as a well-formed URL.

```go
type MyAppConfig struct {
    DBUser   string // Will be validated if the field name ends with "User"
    DBURL    string // Will be validated if the field name ends with "URL"
    AppPath  string // Will be expanded if the field name ends with "Path"
    APISecret string // Will be auto-generated if empty (ends with "Secret")
}
```

> **Note**: If you don't need struct-based reflection or processing, you can also load into a `map[string]interface{}` or any generic type.

#### Loading a Config File

```go
// Default config if file doesn't exist:
defaultCfg := MyAppConfig{
    DBUser:    "defaultuser",
    DBURL:     "https://example.com",
    AppPath:   "~/myapp",
    APISecret: "",
}

// Load or create the YAML file:
cfg, err := config.Load[MyAppConfig]("config.yaml", defaultCfg)
if err != nil {
    panic(err)
}
```

- If `config.yaml` does **not** exist, `Load` will:
  1. **Process** your `defaultCfg` (auto-generate secrets, expand paths, etc.).
  2. **Save** that as a brand-new `config.yaml`.
  3. Return the processed `defaultCfg`.

- If `config.yaml` **does** exist, `Load` will:
  1. Read and unmarshal the existing YAML.
  2. **Process** it (auto-generate secrets if empty, expand paths, etc.).
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

When you call `Load` (or manually call `config.Process`), the following rules apply to fields in a struct:

- **`<Name>Secret`** (string): If empty, generates a 32-byte random JWT secret, hex-encoded.
- **`<Name>Path`** (string): Expands `~` to the user’s home directory.
- **`<Name>User`** (string): Must pass a simple Linux username check.
- **`<Name>URL`** (string): Must be a valid URL with scheme and host.

If you only have a generic map (e.g., `map[string]interface{}`), then these rules do not apply automatically. However, you can still call helper functions (like `config.ExpandPath`, `config.ValidateUsername`) manually if needed.

### Other Helpers

- **`GenerateJWTSecret()`**  
  Creates a 32-byte secure random key, returns hex-encoded string.
  
- **`ExpandPath(path string)`**  
  Expands a `~` to the user’s home directory, e.g. `"~/myapp" -> "/Users/joe/myapp"`.
  
- **`ValidateUsername(username string)`**  
  Validates against Linux-style username rules. Returns an error if invalid.
  
- **`ValidateURL(url string)`**  
  Checks if the string is a valid URL with both scheme and host set.

- **`CopyProperty(src, srcField, dst, dstField)`**  
  Copies a named field from one struct to another (both must be pointer-to-struct). Useful if you prefer a more structured approach (rather than map usage).

---

## CLI Usage

The `config` CLI mirrors much of this functionality. You can see usage by simply running:

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
              -srcField string Source field name
              -dst string      Destination config file
              -dstField string Destination field name
```

---

## License

[MIT](LICENSE)