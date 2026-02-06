package internal

import (
	"testing"
)

func TestParseDSN(t *testing.T) {
	tests := []struct {
		name         string
		dsn          string
		wantUser     string
		wantPassword string
		wantHost     string
		wantPort     string
		wantDatabase string
		expectError  bool
	}{
		{
			name:         "DSN with protocol and password",
			dsn:          "mysql://user:pass@localhost:3306/mydb",
			wantUser:     "user",
			wantPassword: "pass",
			wantHost:     "localhost",
			wantPort:     "3306",
			wantDatabase: "mydb",
			expectError:  false,
		},
		{
			name:         "DSN without protocol and with password",
			dsn:          "user:pass@localhost:3306/mydb",
			wantUser:     "user",
			wantPassword: "pass",
			wantHost:     "localhost",
			wantPort:     "3306",
			wantDatabase: "mydb",
			expectError:  false,
		},
		{
			name:         "DSN without protocol and without password",
			dsn:          "user@localhost:3306/mydb",
			wantUser:     "user",
			wantPassword: "",
			wantHost:     "localhost",
			wantPort:     "3306",
			wantDatabase: "mydb",
			expectError:  false,
		},
		{
			name:         "DSN with protocol and without password",
			dsn:          "mysql://user@localhost:3306/mydb",
			wantUser:     "user",
			wantPassword: "",
			wantHost:     "localhost",
			wantPort:     "3306",
			wantDatabase: "mydb",
			expectError:  false,
		},
		{
			name:         "DSN with default port",
			dsn:          "mysql://user:pass@localhost/mydb",
			wantUser:     "user",
			wantPassword: "pass",
			wantHost:     "localhost",
			wantPort:     "3306",
			wantDatabase: "mydb",
			expectError:  false,
		},
		{
			name:         "DSN with custom port",
			dsn:          "mysql://user:pass@localhost:3307/mydb",
			wantUser:     "user",
			wantPassword: "pass",
			wantHost:     "localhost",
			wantPort:     "3307",
			wantDatabase: "mydb",
			expectError:  false,
		},
		{
			name:         "DSN with IP address",
			dsn:          "mysql://user:pass@127.0.0.1:3306/mydb",
			wantUser:     "user",
			wantPassword: "pass",
			wantHost:     "127.0.0.1",
			wantPort:     "3306",
			wantDatabase: "mydb",
			expectError:  false,
		},
		{
			name:         "DSN with special characters in password",
			dsn:          "mysql://user:p@ssw0rd@localhost:3306/mydb",
			wantUser:     "user",
			wantPassword: "p@ssw0rd",
			wantHost:     "localhost",
			wantPort:     "3306",
			wantDatabase: "mydb",
			expectError:  false,
		},
		{
			name:         "Invalid DSN - missing database",
			dsn:          "mysql://user:pass@localhost:3306/",
			wantUser:     "",
			wantPassword: "",
			wantHost:     "",
			wantPort:     "",
			wantDatabase: "",
			expectError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseDSN(tt.dsn)

			if tt.expectError {
				if err == nil {
					t.Errorf("ParseDSN() expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("ParseDSN() unexpected error: %v", err)
				return
			}

			if got.User != tt.wantUser {
				t.Errorf("ParseDSN() User = %v, want %v", got.User, tt.wantUser)
			}
			if got.Password != tt.wantPassword {
				t.Errorf("ParseDSN() Password = %v, want %v", got.Password, tt.wantPassword)
			}
			if got.Host != tt.wantHost {
				t.Errorf("ParseDSN() Host = %v, want %v", got.Host, tt.wantHost)
			}
			if got.Port != tt.wantPort {
				t.Errorf("ParseDSN() Port = %v, want %v", got.Port, tt.wantPort)
			}
			if got.Database != tt.wantDatabase {
				t.Errorf("ParseDSN() Database = %v, want %v", got.Database, tt.wantDatabase)
			}
		})
	}
}

func TestBuildDSN(t *testing.T) {
	tests := []struct {
		name     string
		parsed   *ParsedDSN
		password string
		want     string
	}{
		{
			name: "Build DSN with password",
			parsed: &ParsedDSN{
				User:     "user",
				Host:     "localhost",
				Port:     "3306",
				Database: "mydb",
			},
			password: "pass",
			want:     "mysql://user:pass@localhost:3306/mydb",
		},
		{
			name: "Build DSN without password",
			parsed: &ParsedDSN{
				User:     "user",
				Host:     "localhost",
				Port:     "3306",
				Database: "mydb",
			},
			password: "",
			want:     "mysql://user@localhost:3306/mydb",
		},
		{
			name: "Build DSN with custom port",
			parsed: &ParsedDSN{
				User:     "user",
				Host:     "localhost",
				Port:     "3307",
				Database: "mydb",
			},
			password: "pass",
			want:     "mysql://user:pass@localhost:3307/mydb",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := BuildDSN(tt.parsed, tt.password)
			if got != tt.want {
				t.Errorf("BuildDSN() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetPasswordFromDSN(t *testing.T) {
	tests := []struct {
		name         string
		dsn          string
		wantPassword string
		expectError  bool
	}{
		{
			name:         "DSN with password",
			dsn:          "mysql://user:pass@localhost:3306/mydb",
			wantPassword: "pass",
			expectError:  false,
		},
		{
			name:         "DSN without password",
			dsn:          "mysql://user@localhost:3306/mydb",
			wantPassword: "",
			expectError:  false,
		},
		{
			name:         "DSN without protocol and with password",
			dsn:          "user:pass@localhost:3306/mydb",
			wantPassword: "pass",
			expectError:  false,
		},
		{
			name:         "Invalid DSN",
			dsn:          "invalid-dsn",
			wantPassword: "",
			expectError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetPasswordFromDSN(tt.dsn)

			if tt.expectError {
				if err == nil {
					t.Errorf("GetPasswordFromDSN() expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("GetPasswordFromDSN() unexpected error: %v", err)
				return
			}

			if got != tt.wantPassword {
				t.Errorf("GetPasswordFromDSN() = %v, want %v", got, tt.wantPassword)
			}
		})
	}
}
