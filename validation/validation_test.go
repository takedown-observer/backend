package validation

import (
	"strings"
	"testing"
)

func TestValidateUUID(t *testing.T) {
	tests := []struct {
		name     string
		id       string
		expected bool
	}{
		{
			name:     "valid UUID",
			id:       "123e4567-e89b-12d3-a456-426614174000",
			expected: true,
		},
		{
			name:     "valid UUID uppercase",
			id:       "123E4567-E89B-12D3-A456-426614174000",
			expected: true,
		},
		{
			name:     "invalid UUID - too short",
			id:       "123e4567",
			expected: false,
		},
		{
			name:     "invalid UUID - wrong format",
			id:       "123e4567-wrong-format",
			expected: false,
		},
		{
			name:     "empty string",
			id:       "",
			expected: false,
		},
		{
			name:     "invalid characters",
			id:       "123e4567-e89b-12d3-a456-42661417400g",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ValidateUUID(tt.id)
			if result != tt.expected {
				t.Errorf("ValidateUUID(%q) = %v, want %v", tt.id, result, tt.expected)
			}
		})
	}
}

func TestValidateAccountID(t *testing.T) {
	tests := []struct {
		name          string
		id            string
		expectError   bool
		errorContains string
	}{
		{
			name:          "valid ID",
			id:            "valid_account_123",
			expectError:   false,
			errorContains: "",
		},
		{
			name:          "empty ID",
			id:            "",
			expectError:   true,
			errorContains: "cannot be empty",
		},
		{
			name:          "too long ID",
			id:            strings.Repeat("a", MaxAccountIDLength+1),
			expectError:   true,
			errorContains: "exceeds maximum length",
		},
		{
			name:          "invalid characters",
			id:            "invalid-account!@#",
			expectError:   true,
			errorContains: "invalid characters",
		},
		{
			name:          "spaces not allowed",
			id:            "account with spaces",
			expectError:   true,
			errorContains: "invalid characters",
		},
		{
			name:          "maximum length",
			id:            strings.Repeat("a", MaxAccountIDLength),
			expectError:   false,
			errorContains: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateAccountID(tt.id)
			if tt.expectError {
				if err == nil {
					t.Errorf("ValidateAccountID(%q) expected error containing %q, got nil", tt.id, tt.errorContains)
				} else if !strings.Contains(err.Error(), tt.errorContains) {
					t.Errorf("ValidateAccountID(%q) error = %v, want error containing %q", tt.id, err, tt.errorContains)
				}
			} else if err != nil {
				t.Errorf("ValidateAccountID(%q) unexpected error: %v", tt.id, err)
			}
		})
	}
}

func TestValidateAccountName(t *testing.T) {
	tests := []struct {
		name          string
		accountName   string
		expectError   bool
		errorContains string
	}{
		{
			name:          "valid name",
			accountName:   "valid_account_123",
			expectError:   false,
			errorContains: "",
		},
		{
			name:          "account name contains underscore",
			accountName:   "valid_name",
			expectError:   false,
			errorContains: "",
		},
		{
			name:          "empty name",
			accountName:   "",
			expectError:   true,
			errorContains: "cannot be empty",
		},
		{
			name:          "too long name",
			accountName:   strings.Repeat("a", MaxAccountNameLength+1),
			expectError:   true,
			errorContains: "exceeds maximum length",
		},
		{
			name:          "invalid characters",
			accountName:   "invalid-name!@#",
			expectError:   true,
			errorContains: "invalid characters",
		},
		{
			name:          "spaces not allowed",
			accountName:   "account with spaces",
			expectError:   true,
			errorContains: "invalid characters",
		},
		{
			name:          "maximum length",
			accountName:   strings.Repeat("a", MaxAccountNameLength),
			expectError:   false,
			errorContains: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateAccountName(tt.accountName)
			if tt.expectError {
				if err == nil {
					t.Errorf("ValidateAccountName(%q) expected error containing %q, got nil", tt.accountName, tt.errorContains)
				} else if !strings.Contains(err.Error(), tt.errorContains) {
					t.Errorf("ValidateAccountName(%q) error = %v, want error containing %q", tt.accountName, err, tt.errorContains)
				}
			} else if err != nil {
				t.Errorf("ValidateAccountName(%q) unexpected error: %v", tt.accountName, err)
			}
		})
	}
}

func TestValidateCountries(t *testing.T) {
	tests := []struct {
		name          string
		countries     []string
		expectError   bool
		errorContains string
	}{
		{
			name:          "valid countries",
			countries:     []string{"US", "GB", "FR"},
			expectError:   false,
			errorContains: "",
		},
		{
			name:          "empty list",
			countries:     []string{},
			expectError:   true,
			errorContains: "cannot be empty",
		},
		{
			name:          "too many countries",
			countries:     make([]string, MaxCountries+1),
			expectError:   true,
			errorContains: "exceeds maximum",
		},
		{
			name:          "invalid country code length",
			countries:     []string{"USA"},
			expectError:   true,
			errorContains: "not 2 characters",
		},
		{
			name:          "invalid country code format",
			countries:     []string{"U1"},
			expectError:   true,
			errorContains: "not valid",
		},
		{
			name:          "duplicate countries",
			countries:     []string{"US", "GB", "US"},
			expectError:   true,
			errorContains: "duplicate country code",
		},
		{
			name:          "lowercase not allowed",
			countries:     []string{"us", "gb"},
			expectError:   true,
			errorContains: "not valid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateCountries(tt.countries)
			if tt.expectError {
				if err == nil {
					t.Errorf("ValidateCountries(%v) expected error containing %q, got nil", tt.countries, tt.errorContains)
				} else if !strings.Contains(err.Error(), tt.errorContains) {
					t.Errorf("ValidateCountries(%v) error = %v, want error containing %q", tt.countries, err, tt.errorContains)
				}
			} else if err != nil {
				t.Errorf("ValidateCountries(%v) unexpected error: %v", tt.countries, err)
			}
		})
	}
}

func TestSanitizeString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "normal string",
			input:    "normal string",
			expected: "normal string",
		},
		{
			name:     "string with control characters",
			input:    "test\x00string\x01with\x02control\x03chars",
			expected: "teststringwithcontrolchars",
		},
		{
			name:     "string with HTML characters",
			input:    "<script>alert('xss')</script>",
			expected: "&lt;script&gt;alert(&#39;xss&#39;)&lt;/script&gt;",
		},
		{
			name:     "string with spaces and tabs",
			input:    "  test  string\t",
			expected: "test  string",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "string with mixed content",
			input:    "  <b>Test</b>\x00\x01String  ",
			expected: "&lt;b&gt;Test&lt;/b&gt;String",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SanitizeString(tt.input)
			if result != tt.expected {
				t.Errorf("SanitizeString(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}
