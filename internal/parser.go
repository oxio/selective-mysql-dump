package internal

import (
	"fmt"
	"net/url"
	"os"
	"strings"

	"golang.org/x/term"
)

// ParsedDSN represents a parsed MySQL DSN
type ParsedDSN struct {
	User     string
	Password string
	Host     string
	Port     string
	Database string
}

// ParseDSN parses a MySQL DSN string
// Format: mysql://user:password@host:port/database
// Or: user:password@host:port/database (protocol auto-added)
// Or: user@host:port/database (password empty)
func ParseDSN(dsn string) (*ParsedDSN, error) {
	// Add mysql:// prefix if not present
	if !strings.Contains(dsn, "://") {
		dsn = "mysql://" + dsn
	}

	// Parse using net/url
	u, err := url.Parse(dsn)
	if err != nil {
		return nil, fmt.Errorf("invalid DSN format: %w", err)
	}

	// Extract password from URL user info
	var password string
	if u.User != nil {
		password, _ = u.User.Password()
	}

	// Extract database name from path
	database := strings.TrimPrefix(u.Path, "/")
	if database == "" {
		return nil, fmt.Errorf("invalid DSN format: database name is required")
	}

	parsed := &ParsedDSN{
		User:     u.User.Username(),
		Password: password,
		Host:     u.Hostname(),
		Port:     u.Port(),
		Database: database,
	}

	// Set default port if not specified
	if parsed.Port == "" {
		parsed.Port = "3306"
	}

	return parsed, nil
}

// PromptForPassword prompts the user for a password
func PromptForPassword() (string, error) {
	fmt.Print("Enter password: ")

	// Read password without echoing
	bytePassword, err := term.ReadPassword(int(os.Stdin.Fd()))
	if err != nil {
		return "", fmt.Errorf("failed to read password: %w", err)
	}
	fmt.Println() // Add newline after password input

	return string(bytePassword), nil
}

// BuildDSN builds a DSN string from parsed components
func BuildDSN(parsed *ParsedDSN, password string) string {
	u := &url.URL{
		Scheme: "mysql",
		Host:   fmt.Sprintf("%s:%s", parsed.Host, parsed.Port),
		Path:   "/" + parsed.Database,
	}

	if password != "" {
		u.User = url.UserPassword(parsed.User, password)
	} else {
		u.User = url.User(parsed.User)
	}

	return u.String()
}

// GetPasswordFromDSN extracts the password from a DSN if present
// Returns empty string if password is not set
func GetPasswordFromDSN(dsn string) (string, error) {
	parsed, err := ParseDSN(dsn)
	if err != nil {
		return "", err
	}
	return parsed.Password, nil
}
