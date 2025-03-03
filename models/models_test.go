package models

import (
	"encoding/json"
	"testing"
)

func TestAccountJSONSerialization(t *testing.T) {
	account := Account{
		ID:                "test_account",
		Name:              "TestAccount",
		Countries:         []string{"US", "GB"},
		ReportedBy:        []string{"client1"},
		DataFormatVersion: "1.0",
	}

	// Test marshaling
	data, err := json.Marshal(account)
	if err != nil {
		t.Fatalf("Failed to marshal Account: %v", err)
	}

	// Test unmarshaling
	var unmarshaled Account
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("Failed to unmarshal Account: %v", err)
	}

	// Verify fields
	if unmarshaled.ID != account.ID {
		t.Errorf("Expected ID %s, got %s", account.ID, unmarshaled.ID)
	}
	if unmarshaled.Name != account.Name {
		t.Errorf("Expected Name %s, got %s", account.Name, unmarshaled.Name)
	}
	if len(unmarshaled.Countries) != len(account.Countries) {
		t.Errorf("Expected %d countries, got %d", len(account.Countries), len(unmarshaled.Countries))
	}
	// ReportedBy should not be included in JSON
	if string(data) != "" && string(data) != "{}" {
		if json.Valid(data) {
			var jsonMap map[string]interface{}
			json.Unmarshal(data, &jsonMap)
			if _, exists := jsonMap["reported_by"]; exists {
				t.Error("ReportedBy field should not be present in JSON")
			}
		}
	}
}
