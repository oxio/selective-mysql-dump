package internal

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Config represents the YAML configuration structure
type Config struct {
	DSN    string `yaml:"dsn"`
	Tables Tables `yaml:"tables"`
}

// Tables represents the table lists in the config
type Tables struct {
	WithData      []string `yaml:"with_data"`
	StructureOnly []string `yaml:"structure_only"`
}

// LoadConfig loads the configuration from the specified path
func LoadConfig(configPath string) (*Config, error) {
	// If no path provided, use default
	if configPath == "" {
		configPath = GetDefaultConfigPath()
	}

	// Check if file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return nil, &ConfigNotFoundError{
			Path: configPath,
		}
	}

	// Read the file
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Parse YAML
	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Validate config
	if config.DSN == "" {
		return nil, fmt.Errorf("config error: dsn is required")
	}

	return &config, nil
}

// GetDefaultConfigPath returns the default config file path
func GetDefaultConfigPath() string {
	return filepath.Join(".", ".smdump.yaml")
}

// ShowExampleConfig returns an example configuration
func ShowExampleConfig() string {
	return `# smdump configuration file
# Place this file as .smdump.yaml in your project root

# MySQL/MariaDB DSN (Data Source Name)
# Format: user:password@tcp(host:port)/database
# Leave password empty to be prompted for it
dsn: "user:password@tcp(localhost:3306)/database"

# Tables to dump
tables:
  # Tables to dump with structure AND data
  with_data:
    - users
    - orders
    - products

  # Tables to dump with structure ONLY (--no-data)
  structure_only:
    - migrations
    - cache
    - sessions
`
}

// ConfigNotFoundError is returned when the config file is not found
type ConfigNotFoundError struct {
	Path string
}

func (e *ConfigNotFoundError) Error() string {
	return fmt.Sprintf("config file not found: %s", e.Path)
}

// ShowHelpMessage returns a helpful error message with example config
func (e *ConfigNotFoundError) ShowHelpMessage() string {
	return fmt.Sprintf(`Config file not found: %s

Please create a config file. Example:

%s
`, e.Path, ShowExampleConfig())
}
