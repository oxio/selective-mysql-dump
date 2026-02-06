package internal

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
)

// Executor handles mysqldump command execution
type Executor struct {
	DSN        string
	OutputFile string
}

// NewExecutor creates a new Executor instance
func NewExecutor(dsn, outputFile string) *Executor {
	return &Executor{
		DSN:        dsn,
		OutputFile: outputFile,
	}
}

// DumpTablesWithData dumps the specified tables with structure and data
func (e *Executor) DumpTablesWithData(tables []string) error {
	if len(tables) == 0 {
		return nil
	}
	if e.OutputFile != "" {
		fmt.Println("Dumping tables with data...")
	}
	return e.executeMysqldump(tables, false)
}

// DumpTablesStructureOnly dumps the specified tables with structure only (--no-data)
func (e *Executor) DumpTablesStructureOnly(tables []string) error {
	if len(tables) == 0 {
		return nil
	}
	if e.OutputFile != "" {
		fmt.Println("Dumping tables structure only...")
	}
	return e.executeMysqldump(tables, true)
}

// executeMysqldump executes mysqldump command
func (e *Executor) executeMysqldump(tables []string, noData bool) error {
	// Parse DSN to get connection parameters
	parsed, err := ParseDSN(e.DSN)
	if err != nil {
		return fmt.Errorf("failed to parse DSN: %w", err)
	}

	// Check if mysqldump is available
	if _, err := exec.LookPath("mysqldump"); err != nil {
		return fmt.Errorf("mysqldump command not found in PATH: %w", err)
	}

	// Build mysqldump command
	cmd := e.buildMysqldumpCommand(parsed, tables, noData)

	// Set up output
	var output io.Writer
	if e.OutputFile != "" {
		file, err := os.Create(e.OutputFile)
		if err != nil {
			return fmt.Errorf("failed to create output file: %w", err)
		}
		defer file.Close()
		output = file
	} else {
		output = os.Stdout
	}

	// Execute command and capture output
	cmd.Stdout = output
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("mysqldump failed: %w", err)
	}

	return nil
}

// buildMysqldumpCommand builds the mysqldump command
func (e *Executor) buildMysqldumpCommand(parsed *ParsedDSN, tables []string, noData bool) *exec.Cmd {
	args := []string{
		"-h", parsed.Host,
		"-P", parsed.Port,
		"-u", parsed.User,
		"--skip-lock-tables",
		"--single-transaction",
	}

	// Add password if present
	if parsed.Password != "" {
		args = append(args, fmt.Sprintf("-p%s", parsed.Password))
	}

	// Add --no-data flag if structure only
	if noData {
		args = append(args, "--no-data")
	}

	// Add database name
	args = append(args, parsed.Database)

	// Add tables
	args = append(args, tables...)

	return exec.Command("mysqldump", args...)
}

// DumpAll dumps all configured tables
func (e *Executor) DumpAll(withData, structureOnly []string) error {
	// Dump tables with structure and data first
	if len(withData) > 0 {
		if err := e.DumpTablesWithData(withData); err != nil {
			return fmt.Errorf("failed to dump tables with data: %w", err)
		}
	}

	// Dump tables with structure only
	if len(structureOnly) > 0 {
		if err := e.DumpTablesStructureOnly(structureOnly); err != nil {
			return fmt.Errorf("failed to dump tables structure only: %w", err)
		}
	}

	// Print completion message if output file is specified
	if e.OutputFile != "" {
		fmt.Println("Done.")
	}

	return nil
}

// ValidateTables checks if the table lists are valid
func ValidateTables(withData, structureOnly []string) error {
	// Check for duplicates
	tableMap := make(map[string]bool)
	for _, table := range withData {
		if tableMap[table] {
			return fmt.Errorf("duplicate table in with_data: %s", table)
		}
		tableMap[table] = true
	}

	for _, table := range structureOnly {
		if tableMap[table] {
			return fmt.Errorf("table %s appears in both with_data and structure_only", table)
		}
		tableMap[table] = true
	}

	// Check for empty table names
	for _, table := range withData {
		if strings.TrimSpace(table) == "" {
			return fmt.Errorf("empty table name in with_data")
		}
	}

	for _, table := range structureOnly {
		if strings.TrimSpace(table) == "" {
			return fmt.Errorf("empty table name in structure_only")
		}
	}

	return nil
}
