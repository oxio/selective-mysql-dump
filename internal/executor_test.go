package internal

import (
	"testing"
)

func TestBuildMysqldumpCommand(t *testing.T) {
	executor := NewExecutor("mysql://testuser:testpass@localhost:3306/testdb", "output.sql")

	tests := []struct {
		name     string
		parsed   *ParsedDSN
		tables   []string
		noData   bool
		wantPath string
		wantArgs []string
	}{
		{
			name: "basic command with password and data",
			parsed: &ParsedDSN{
				User:     "testuser",
				Password: "testpass",
				Host:     "localhost",
				Port:     "3306",
				Database: "testdb",
			},
			tables:   []string{"users", "posts"},
			noData:   false,
			wantPath: "mysqldump",
			wantArgs: []string{
				"mysqldump",
				"-h", "localhost",
				"-P", "3306",
				"-u", "testuser",
				"--skip-lock-tables",
				"--single-transaction",
				"-ptestpass",
				"testdb",
				"users", "posts",
			},
		},
		{
			name: "command without password",
			parsed: &ParsedDSN{
				User:     "testuser",
				Password: "",
				Host:     "localhost",
				Port:     "3306",
				Database: "testdb",
			},
			tables:   []string{"users"},
			noData:   false,
			wantPath: "mysqldump",
			wantArgs: []string{
				"mysqldump",
				"-h", "localhost",
				"-P", "3306",
				"-u", "testuser",
				"--skip-lock-tables",
				"--single-transaction",
				"testdb",
				"users",
			},
		},
		{
			name: "command with no-data flag",
			parsed: &ParsedDSN{
				User:     "testuser",
				Password: "testpass",
				Host:     "localhost",
				Port:     "3306",
				Database: "testdb",
			},
			tables:   []string{"schema_migrations"},
			noData:   true,
			wantPath: "mysqldump",
			wantArgs: []string{
				"mysqldump",
				"-h", "localhost",
				"-P", "3306",
				"-u", "testuser",
				"--skip-lock-tables",
				"--single-transaction",
				"-ptestpass",
				"--no-data",
				"testdb",
				"schema_migrations",
			},
		},
		{
			name: "command with multiple tables",
			parsed: &ParsedDSN{
				User:     "testuser",
				Password: "testpass",
				Host:     "localhost",
				Port:     "3306",
				Database: "testdb",
			},
			tables:   []string{"users", "posts", "comments", "likes"},
			noData:   false,
			wantPath: "mysqldump",
			wantArgs: []string{
				"mysqldump",
				"-h", "localhost",
				"-P", "3306",
				"-u", "testuser",
				"--skip-lock-tables",
				"--single-transaction",
				"-ptestpass",
				"testdb",
				"users", "posts", "comments", "likes",
			},
		},
		{
			name: "command with single table",
			parsed: &ParsedDSN{
				User:     "testuser",
				Password: "testpass",
				Host:     "localhost",
				Port:     "3306",
				Database: "testdb",
			},
			tables:   []string{"users"},
			noData:   false,
			wantPath: "mysqldump",
			wantArgs: []string{
				"mysqldump",
				"-h", "localhost",
				"-P", "3306",
				"-u", "testuser",
				"--skip-lock-tables",
				"--single-transaction",
				"-ptestpass",
				"testdb",
				"users",
			},
		},
		{
			name: "command with no tables",
			parsed: &ParsedDSN{
				User:     "testuser",
				Password: "testpass",
				Host:     "localhost",
				Port:     "3306",
				Database: "testdb",
			},
			tables:   []string{},
			noData:   false,
			wantPath: "mysqldump",
			wantArgs: []string{
				"mysqldump",
				"-h", "localhost",
				"-P", "3306",
				"-u", "testuser",
				"--skip-lock-tables",
				"--single-transaction",
				"-ptestpass",
				"testdb",
			},
		},
		{
			name: "command with custom host and port",
			parsed: &ParsedDSN{
				User:     "admin",
				Password: "secret",
				Host:     "db.example.com",
				Port:     "3307",
				Database: "production",
			},
			tables:   []string{"orders"},
			noData:   false,
			wantPath: "mysqldump",
			wantArgs: []string{
				"mysqldump",
				"-h", "db.example.com",
				"-P", "3307",
				"-u", "admin",
				"--skip-lock-tables",
				"--single-transaction",
				"-psecret",
				"production",
				"orders",
			},
		},
		{
			name: "command with no-data and no password",
			parsed: &ParsedDSN{
				User:     "readonly",
				Password: "",
				Host:     "localhost",
				Port:     "3306",
				Database: "testdb",
			},
			tables:   []string{"schema"},
			noData:   true,
			wantPath: "mysqldump",
			wantArgs: []string{
				"mysqldump",
				"-h", "localhost",
				"-P", "3306",
				"-u", "readonly",
				"--skip-lock-tables",
				"--single-transaction",
				"--no-data",
				"testdb",
				"schema",
			},
		},
		{
			name: "command with special characters in password",
			parsed: &ParsedDSN{
				User:     "testuser",
				Password: "p@ssw0rd!#$",
				Host:     "localhost",
				Port:     "3306",
				Database: "testdb",
			},
			tables:   []string{"users"},
			noData:   false,
			wantPath: "mysqldump",
			wantArgs: []string{
				"mysqldump",
				"-h", "localhost",
				"-P", "3306",
				"-u", "testuser",
				"--skip-lock-tables",
				"--single-transaction",
				"-pp@ssw0rd!#$",
				"testdb",
				"users",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := executor.buildMysqldumpCommand(tt.parsed, tt.tables, tt.noData)

			// Verify command path
			if cmd.Path != tt.wantPath {
				t.Errorf("buildMysqldumpCommand() Path = %v, want %v", cmd.Path, tt.wantPath)
			}

			// Verify command arguments
			if len(cmd.Args) != len(tt.wantArgs) {
				t.Errorf("buildMysqldumpCommand() Args length = %d, want %d", len(cmd.Args), len(tt.wantArgs))
			} else {
				for i, arg := range cmd.Args {
					if arg != tt.wantArgs[i] {
						t.Errorf("buildMysqldumpCommand() Args[%d] = %v, want %v", i, arg, tt.wantArgs[i])
					}
				}
			}
		})
	}
}

func TestBuildMysqldumpCommand_ArgsOrder(t *testing.T) {
	executor := NewExecutor("mysql://testuser:testpass@localhost:3306/testdb", "output.sql")

	parsed := &ParsedDSN{
		User:     "testuser",
		Password: "testpass",
		Host:     "localhost",
		Port:     "3306",
		Database: "testdb",
	}

	cmd := executor.buildMysqldumpCommand(parsed, []string{"users", "posts"}, false)

	// Verify the order of arguments
	expectedOrder := []string{
		"mysqldump",
		"-h", "localhost",
		"-P", "3306",
		"-u", "testuser",
		"--skip-lock-tables",
		"--single-transaction",
		"-ptestpass",
		"testdb",
		"users", "posts",
	}

	for i, arg := range expectedOrder {
		if i >= len(cmd.Args) {
			t.Errorf("Expected arg at index %d but command has only %d args", i, len(cmd.Args))
			break
		}
		if cmd.Args[i] != arg {
			t.Errorf("Arg at index %d = %v, want %v", i, cmd.Args[i], arg)
		}
	}
}

func TestBuildMysqldumpCommand_ReturnsExecCmd(t *testing.T) {
	executor := NewExecutor("mysql://testuser:testpass@localhost:3306/testdb", "output.sql")

	parsed := &ParsedDSN{
		User:     "testuser",
		Password: "testpass",
		Host:     "localhost",
		Port:     "3306",
		Database: "testdb",
	}

	cmd := executor.buildMysqldumpCommand(parsed, []string{"users"}, false)

	// Verify it returns a non-nil command
	if cmd == nil {
		t.Fatal("buildMysqldumpCommand() returned nil")
	}

	// Verify it has the correct command path
	if cmd.Path != "mysqldump" {
		t.Errorf("buildMysqldumpCommand() Path = %v, want mysqldump", cmd.Path)
	}
}
