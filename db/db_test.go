package db

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNew(t *testing.T) {
	// Create a temporary directory for the test database
	tmpDir, err := os.MkdirTemp("", "takedown-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	dbPath := filepath.Join(tmpDir, "test.db")

	// Test database creation
	_, err = New(dbPath)
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}

	// Verify the database file exists
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		t.Error("Database file was not created")
	}

	// Test with invalid path
	invalidPath := filepath.Join(tmpDir, "nonexistent", "test.db")
	_, err = New(invalidPath)
	if err == nil {
		t.Error("Expected error for invalid database path")
	}
}
