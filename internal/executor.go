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

// executeMysqldump executes mysqldump command, writing output to the given writer
func (e *Executor) executeMysqldump(w io.Writer, tables []string, noData bool) error {
	parsed, err := ParseDSN(e.DSN)
	if err != nil {
		return fmt.Errorf("failed to parse DSN: %w", err)
	}

	if _, err := exec.LookPath("mysqldump"); err != nil {
		return fmt.Errorf("mysqldump command not found in PATH: %w", err)
	}

	cmd := e.buildMysqldumpCommand(parsed, tables, noData)
	cmd.Stdout = w
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

	if parsed.Password != "" {
		args = append(args, fmt.Sprintf("-p%s", parsed.Password))
	}

	if noData {
		args = append(args, "--no-data")
	}

	args = append(args, parsed.Database)
	args = append(args, tables...)

	return exec.Command("mysqldump", args...)
}

// DumpAll dumps all configured tables using a staged approach.
// When an output file is specified, each stage writes to a temp file and
// the results are merged into the final output. Without an output file,
// each stage streams directly to stdout.
func (e *Executor) DumpAll(withData, structureOnly []string) error {
	if e.OutputFile != "" {
		return e.dumpToFile(withData, structureOnly)
	}
	return e.dumpToStdout(withData, structureOnly)
}

// dumpToStdout streams each stage directly to stdout
func (e *Executor) dumpToStdout(withData, structureOnly []string) error {
	// Stage 1: structure of with_data tables
	if len(withData) > 0 {
		if err := e.executeMysqldump(os.Stdout, withData, true); err != nil {
			return fmt.Errorf("failed to dump structure of with_data tables: %w", err)
		}
	}

	// Stage 2: structure of structure_only tables
	if len(structureOnly) > 0 {
		if err := e.executeMysqldump(os.Stdout, structureOnly, true); err != nil {
			return fmt.Errorf("failed to dump structure_only tables: %w", err)
		}
	}

	// Stage 3: data of with_data tables
	if len(withData) > 0 {
		if err := e.executeMysqldump(os.Stdout, withData, false); err != nil {
			return fmt.Errorf("failed to dump data of with_data tables: %w", err)
		}
	}

	return nil
}

// dumpToFile writes each stage to a temp file, then merges them into the output file
func (e *Executor) dumpToFile(withData, structureOnly []string) error {
	var tempFiles []string
	defer func() {
		for _, f := range tempFiles {
			os.Remove(f)
		}
	}()

	// Stage 1: structure of with_data tables
	if len(withData) > 0 {
		fmt.Println("Dumping structure of tables with data...")
		tmp, err := e.dumpToTempFile(withData, true)
		if err != nil {
			return fmt.Errorf("failed to dump structure of with_data tables: %w", err)
		}
		tempFiles = append(tempFiles, tmp)
	}

	// Stage 2: structure of structure_only tables
	if len(structureOnly) > 0 {
		fmt.Println("Dumping structure only tables...")
		tmp, err := e.dumpToTempFile(structureOnly, true)
		if err != nil {
			return fmt.Errorf("failed to dump structure_only tables: %w", err)
		}
		tempFiles = append(tempFiles, tmp)
	}

	// Stage 3: data of with_data tables
	if len(withData) > 0 {
		fmt.Println("Dumping data of tables with data...")
		tmp, err := e.dumpToTempFile(withData, false)
		if err != nil {
			return fmt.Errorf("failed to dump data of with_data tables: %w", err)
		}
		tempFiles = append(tempFiles, tmp)
	}

	// Merge all temp files into the output file
	fmt.Println("Merging dump files...")
	if err := e.mergeTempFiles(tempFiles); err != nil {
		return fmt.Errorf("failed to merge dump files: %w", err)
	}

	fmt.Println("Done.")
	return nil
}

// dumpToTempFile runs mysqldump and writes the output to a new temp file,
// returning the path to the temp file
func (e *Executor) dumpToTempFile(tables []string, noData bool) (string, error) {
	tmp, err := os.CreateTemp("", "smdump-*.sql")
	if err != nil {
		return "", fmt.Errorf("failed to create temp file: %w", err)
	}
	defer tmp.Close()

	if err := e.executeMysqldump(tmp, tables, noData); err != nil {
		os.Remove(tmp.Name())
		return "", err
	}

	return tmp.Name(), nil
}

// mergeTempFiles concatenates the given temp files into the output file
func (e *Executor) mergeTempFiles(tempFiles []string) error {
	out, err := os.Create(e.OutputFile)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer out.Close()

	for _, path := range tempFiles {
		f, err := os.Open(path)
		if err != nil {
			return fmt.Errorf("failed to open temp file %s: %w", path, err)
		}
		if _, err := io.Copy(out, f); err != nil {
			f.Close()
			return fmt.Errorf("failed to copy temp file %s: %w", path, err)
		}
		f.Close()
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
