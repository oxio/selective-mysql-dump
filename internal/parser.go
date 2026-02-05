package internal

import (
	"fmt"
	"os"
	"regexp"

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
// Format: user:password@tcp(host:port)/database
// Or: user@tcp(host:port)/database (password empty)
func ParseDSN(dsn string) (*ParsedDSN, error) {
	// Pattern to match: user:password@tcp(host:port)/database
	// This regex captures:
	// 1. username
	// 2. password (optional)
	// 3. host
	// 4. port (optional)
	// 5. database name
	pattern := `^([^:@]+)(?::([^@]*))?@tcp\(([^:]+)(?::(\d+))?\)/(.+)$`
	re := regexp.MustCompile(pattern)

	matches := re.FindStringSubmatch(dsn)
	if matches == nil {
		return nil, fmt.Errorf("invalid DSN format: %s", dsn)
	}

	parsed := &ParsedDSN{
		User:     matches[1],
		Password: matches[2], // May be empty
		Host:     matches[3],
		Port:     matches[4],
		Database: matches[5],
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
	if password == "" {
		return fmt.Sprintf("%s@tcp(%s:%s)/%s", parsed.User, parsed.Host, parsed.Port, parsed.Database)
	}
	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", parsed.User, password, parsed.Host, parsed.Port, parsed.Database)
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
