package models

import (
	"time"
)

const DataFormatVersion = "1.0"

// Account represents a reported account in the database
type Account struct {
	ID                string    `gorm:"primarykey" json:"id"`
	Name              string    `json:"name"`
	Countries         []string  `gorm:"serializer:json" json:"countries"`
	LastReportedAt    time.Time `json:"last_reported_at"`
	ReportCount       int       `json:"report_count"`
	ReportedBy        []string  `gorm:"serializer:json" json:"-"`
	DataFormatVersion string    `json:"data_format_version"`
}

// ReportedAccount represents the account data in a report request
type ReportedAccount struct {
	ID        string   `json:"id"`
	Name      string   `json:"name"`
	Countries []string `json:"countries"`
}

// ReportRequest represents the full report request from a client
type ReportRequest struct {
	ClientID          string          `json:"client_id"`
	Account           ReportedAccount `json:"account"`
	DataFormatVersion string          `json:"data_format_version"`
}

// AccountsResponse represents the response for the accounts listing endpoint
type AccountsResponse struct {
	Accounts        []Account `json:"accounts"`
	TotalCount      int64     `json:"totalCount"`
	CurrentPage     int       `json:"currentPage"`
	TotalPages      int       `json:"totalPages"`
	UniqueCountries []string  `json:"uniqueCountries"`
}
