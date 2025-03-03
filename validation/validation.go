package validation

import (
	"fmt"
	"html"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/google/uuid"
)

// Validation constants
const (
	MaxAccountIDLength   = 40
	MaxAccountNameLength = 30
	MaxCountries         = 252
	CountryCodeLength    = 2
)

var countryCodePattern = regexp.MustCompile("^[A-Z]{2}$")

// ValidateUUID checks if a string is a valid UUID
func ValidateUUID(id string) bool {
	_, err := uuid.Parse(id)
	return err == nil
}

// ValidateAccountID validates the format of an account ID
func ValidateAccountID(id string) error {
	if id == "" {
		return fmt.Errorf("account ID cannot be empty")
	}

	if utf8.RuneCountInString(id) > MaxAccountIDLength {
		return fmt.Errorf("account ID exceeds maximum length of %d characters", MaxAccountIDLength)
	}

	// Check for valid characters (alphanumeric and underscores)
	matched, _ := regexp.MatchString("^[a-zA-Z0-9_=]+$", id)
	if !matched {
		return fmt.Errorf("account ID contains invalid characters")
	}

	return nil
}

// ValidateAccountName validates the format of an account name
func ValidateAccountName(name string) error {
	if name == "" {
		return fmt.Errorf("account name cannot be empty")
	}

	if utf8.RuneCountInString(name) > MaxAccountNameLength {
		return fmt.Errorf("account name exceeds maximum length of %d characters", MaxAccountNameLength)
	}

	// Allow letters, numbers, underscores
	matched, _ := regexp.MatchString("^[a-zA-Z0-9_]+$", name)
	if !matched {
		return fmt.Errorf("account name contains invalid characters")
	}

	return nil
}

// ValidateCountries validates a list of country codes
func ValidateCountries(countries []string) error {
	if len(countries) == 0 {
		return fmt.Errorf("countries list cannot be empty")
	}

	if len(countries) > MaxCountries {
		return fmt.Errorf("number of countries exceeds maximum of %d", MaxCountries)
	}

	// Create a map to check for duplicates
	seen := make(map[string]bool)

	for _, country := range countries {
		// Check length
		if len(country) != CountryCodeLength {
			return fmt.Errorf("country code '%s' is not 2 characters", country)
		}

		// Check if it's uppercase letters only using pre-compiled pattern
		if !countryCodePattern.MatchString(country) {
			return fmt.Errorf("country code '%s' is not valid", country)
		}

		// Check for duplicates
		if seen[country] {
			return fmt.Errorf("duplicate country code '%s'", country)
		}
		seen[country] = true
	}

	return nil
}

// SanitizeString removes control characters and escapes HTML special characters
func SanitizeString(input string) string {
	// Remove any control characters and trim spaces
	sanitized := strings.Map(func(r rune) rune {
		if r < 32 || r == 127 {
			return -1
		}
		return r
	}, strings.TrimSpace(input))

	// Escape HTML special characters
	sanitized = html.EscapeString(sanitized)

	return sanitized
}
