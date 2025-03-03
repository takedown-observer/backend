package api

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/takedown-observer/backend/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func newTestDB(t *testing.T) *gorm.DB {
	dbName := fmt.Sprintf("file::memory:?db=%p", t.Name)
	db, err := gorm.Open(sqlite.Open(dbName), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}

	err = db.AutoMigrate(&models.Account{})
	if err != nil {
		t.Fatalf("Failed to migrate test database: %v", err)
	}

	return db
}

func TestReportHandler(t *testing.T) {
	tests := []struct {
		name          string
		request       models.ReportRequest
		setupDB       func(*testing.T, *gorm.DB) error
		expectedCode  int
		expectedError string
		checkDB       func(*testing.T, *gorm.DB)
	}{
		{
			name: "valid report - new account",
			request: models.ReportRequest{
				ClientID: "123e4567-e89b-12d3-a456-426614174000",
				Account: models.ReportedAccount{
					ID:        "test_account1",
					Name:      "TestAccount",
					Countries: []string{"US", "GB"},
				},
				DataFormatVersion: "1.0",
			},
			setupDB:      func(t *testing.T, db *gorm.DB) error { return nil },
			expectedCode: http.StatusOK,
			checkDB: func(t *testing.T, db *gorm.DB) {
				var account models.Account
				result := db.First(&account, "id = ?", "test_account1")
				if result.Error != nil {
					t.Errorf("Failed to find created account: %v", result.Error)
				}
				if account.Name != "TestAccount" {
					t.Errorf("Expected account name %s, got %s", "TestAccount", account.Name)
				}
				if account.ReportCount != 1 {
					t.Errorf("Expected report count 1, got %d", account.ReportCount)
				}
			},
		},
		{
			name: "valid report - existing account",
			request: models.ReportRequest{
				ClientID: "123e4567-e89b-12d3-a456-426614174000",
				Account: models.ReportedAccount{
					ID:        "test_account2",
					Name:      "UpdatedName",
					Countries: []string{"US", "GB"},
				},
				DataFormatVersion: "1.0",
			},
			setupDB: func(t *testing.T, db *gorm.DB) error {
				account := models.Account{
					ID:                "test_account2",
					Name:              "OldName",
					Countries:         []string{"US"},
					LastReportedAt:    time.Now(),
					ReportCount:       1,
					ReportedBy:        []string{"other-client"},
					DataFormatVersion: "1.0",
				}
				return db.Create(&account).Error
			},
			expectedCode: http.StatusOK,
			checkDB: func(t *testing.T, db *gorm.DB) {
				var account models.Account
				result := db.First(&account, "id = ?", "test_account2")
				if result.Error != nil {
					t.Errorf("Failed to find updated account: %v", result.Error)
				}
				if account.Name != "UpdatedName" {
					t.Errorf("Expected account name %s, got %s", "UpdatedName", account.Name)
				}
				if account.ReportCount != 2 {
					t.Errorf("Expected report count 2, got %d", account.ReportCount)
				}
			},
		},
		{
			name: "invalid UUID",
			request: models.ReportRequest{
				ClientID: "invalid-uuid",
				Account: models.ReportedAccount{
					ID:        "test_account",
					Name:      "TestAccount",
					Countries: []string{"US"},
				},
				DataFormatVersion: "1.0",
			},
			setupDB:       func(t *testing.T, db *gorm.DB) error { return nil },
			expectedCode:  http.StatusBadRequest,
			expectedError: "Invalid client ID format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := newTestDB(t)

			if err := tt.setupDB(t, db); err != nil {
				t.Fatalf("Failed to setup test database: %v", err)
			}

			handler := NewHandler(db)

			body, _ := json.Marshal(tt.request)
			req := httptest.NewRequest("POST", "/api/report", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			handler.ReportHandler(w, req)

			if w.Code != tt.expectedCode {
				t.Errorf("ReportHandler() status code = %v, want %v", w.Code, tt.expectedCode)
			}

			if tt.expectedError != "" {
				if !bytes.Contains(w.Body.Bytes(), []byte(tt.expectedError)) {
					t.Errorf("ReportHandler() error = %v, want %v", w.Body.String(), tt.expectedError)
				}
			}

			if tt.checkDB != nil {
				tt.checkDB(t, db)
			}
		})
	}
}

func TestGetAccountsHandler(t *testing.T) {
	tests := []struct {
		name          string
		queryParams   map[string]string
		setupDB       func(*testing.T, *gorm.DB) error
		expectedCode  int
		checkResponse func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name: "successful retrieval - no filters",
			queryParams: map[string]string{
				"page": "1",
			},
			setupDB: func(t *testing.T, db *gorm.DB) error {
				account := models.Account{
					ID:                "account1",
					Name:              "Account1",
					Countries:         []string{"IN", "BR"},
					LastReportedAt:    time.Now(),
					ReportCount:       1,
					ReportedBy:        []string{"client1"},
					DataFormatVersion: "1.0",
				}
				return db.Create(&account).Error
			},
			expectedCode: http.StatusOK,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response models.AccountsResponse
				if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
					t.Errorf("Failed to decode response: %v", err)
					return
				}
				if len(response.Accounts) != 1 {
					t.Errorf("Expected 1 account, got %v", response.Accounts)
				}
				if response.TotalCount != 1 {
					t.Errorf("Expected total count 1, got %d", response.TotalCount)
				}
				if len(response.UniqueCountries) != 2 {
					t.Errorf("Expected 2 unique countries, got %d", len(response.UniqueCountries))
				}
			},
		},
		{
			name: "successful retrieval - with country filter",
			queryParams: map[string]string{
				"page":    "1",
				"country": "US",
			},
			setupDB: func(t *testing.T, db *gorm.DB) error {
				accounts := []models.Account{
					{
						ID:                "account2",
						Name:              "Account1",
						Countries:         []string{"US"},
						LastReportedAt:    time.Now(),
						ReportCount:       1,
						ReportedBy:        []string{"client1"},
						DataFormatVersion: "1.0",
					},
					{
						ID:                "account3",
						Name:              "Account2",
						Countries:         []string{"GB"},
						LastReportedAt:    time.Now(),
						ReportCount:       1,
						ReportedBy:        []string{"client1"},
						DataFormatVersion: "1.0",
					},
				}
				for _, account := range accounts {
					if err := db.Create(&account).Error; err != nil {
						return err
					}
				}
				return nil
			},
			expectedCode: http.StatusOK,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response models.AccountsResponse
				if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
					t.Errorf("Failed to decode response: %v", err)
					return
				}
				if len(response.Accounts) != 1 {
					t.Errorf("Expected 1 account, got %d", len(response.Accounts))
				}
				if len(response.Accounts) > 0 && response.Accounts[0].ID != "account2" {
					t.Errorf("Expected account1, got %s", response.Accounts[0].ID)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := newTestDB(t)

			if err := tt.setupDB(t, db); err != nil {
				t.Fatalf("Failed to setup test database: %v", err)
			}

			handler := NewHandler(db)

			req := httptest.NewRequest("GET", "/api/accounts", nil)
			q := req.URL.Query()
			for key, value := range tt.queryParams {
				q.Add(key, value)
			}
			req.URL.RawQuery = q.Encode()

			w := httptest.NewRecorder()
			handler.GetAccountsHandler(w, req)

			if w.Code != tt.expectedCode {
				t.Errorf("GetAccountsHandler() status code = %v, want %v", w.Code, tt.expectedCode)
			}

			if tt.checkResponse != nil {
				tt.checkResponse(t, w)
			}
		})
	}
}

func TestDownloadCSVHandler(t *testing.T) {
	tests := []struct {
		name          string
		setupDB       func(*testing.T, *gorm.DB) error
		expectedCode  int
		checkResponse func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name: "successful download - multiple accounts",
			setupDB: func(t *testing.T, db *gorm.DB) error {
				accounts := []models.Account{
					{
						ID:                "account1",
						Name:              "Account1",
						Countries:         []string{"US", "GB"},
						LastReportedAt:    time.Date(2025, 2, 20, 14, 0, 0, 0, time.UTC),
						ReportCount:       1,
						ReportedBy:        []string{"client1"},
						DataFormatVersion: "1.0",
					},
					{
						ID:                "account2",
						Name:              "Account2",
						Countries:         []string{"FR", "DE"},
						LastReportedAt:    time.Date(2025, 2, 20, 13, 0, 0, 0, time.UTC),
						ReportCount:       2,
						ReportedBy:        []string{"client1", "client2"},
						DataFormatVersion: "1.0",
					},
				}
				for _, account := range accounts {
					if err := db.Create(&account).Error; err != nil {
						return err
					}
				}
				return nil
			},
			expectedCode: http.StatusOK,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				// Check headers
				contentType := w.Header().Get("Content-Type")
				if contentType != "text/csv" {
					t.Errorf("Expected Content-Type text/csv, got %s", contentType)
				}

				contentDisposition := w.Header().Get("Content-Disposition")
				expectedDisposition := "attachment; filename=takedowns.csv"
				if contentDisposition != expectedDisposition {
					t.Errorf("Expected Content-Disposition %s, got %s", expectedDisposition, contentDisposition)
				}

				// Parse CSV content
				body := w.Body.String()
				// Parse CSV content using encoding/csv to handle quotes correctly
				csvReader := csv.NewReader(strings.NewReader(body))
				records, err := csvReader.ReadAll()
				if err != nil {
					t.Fatalf("Failed to parse CSV: %v", err)
				}

				// Check header row
				expectedHeader := []string{"Account ID", "Username", "Countries", "Last Reported At", "Data Format Version"}
				if !reflect.DeepEqual(records[0], expectedHeader) {
					t.Errorf("Expected CSV header %v, got %v", expectedHeader, records[0])
				}

				// Check number of records (header + 2 accounts)
				if len(records) != 3 {
					t.Errorf("Expected 3 CSV records (header + 2 accounts), got %d", len(records))
				}

				// Check first account record
				if records[1][0] != "account1" {
					t.Errorf("Expected first account ID account1, got %s", records[1][0])
				}
				if records[1][1] != "Account1" {
					t.Errorf("Expected first account name Account1, got %s", records[1][1])
				}
				firstCountries := records[1][2]
				if !strings.Contains(firstCountries, "US") || !strings.Contains(firstCountries, "GB") {
					t.Errorf("Expected first account countries to contain US and GB, got %s", firstCountries)
				}

				// Check second account record
				if records[2][0] != "account2" {
					t.Errorf("Expected second account ID account2, got %s", records[2][0])
				}
				if records[2][1] != "Account2" {
					t.Errorf("Expected second account name Account2, got %s", records[2][1])
				}
				secondCountries := records[2][2]
				if !strings.Contains(secondCountries, "FR") || !strings.Contains(secondCountries, "DE") {
					t.Errorf("Expected second account countries to contain FR and DE, got %s", secondCountries)
				}
			},
		},
		{
			name: "successful download - empty database",
			setupDB: func(t *testing.T, db *gorm.DB) error {
				return nil
			},
			expectedCode: http.StatusOK,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				// Check headers
				if ct := w.Header().Get("Content-Type"); ct != "text/csv" {
					t.Errorf("Expected Content-Type text/csv, got %s", ct)
				}

				// Parse CSV content
				body := w.Body.String()
				records := strings.Split(strings.TrimSpace(body), "\n")

				// Should have only the header row
				if len(records) != 1 {
					t.Errorf("Expected 1 CSV record (header only), got %d", len(records))
				}

				expectedHeader := "Account ID,Username,Countries,Last Reported At,Data Format Version"
				if records[0] != expectedHeader {
					t.Errorf("Expected CSV header %s, got %s", expectedHeader, records[0])
				}
			},
		},
		{
			name: "successful download - special characters in data",
			setupDB: func(t *testing.T, db *gorm.DB) error {
				account := models.Account{
					ID:                "account_with,comma",
					Name:              "Name,with,commas",
					Countries:         []string{"US", "GB"},
					LastReportedAt:    time.Date(2025, 2, 20, 12, 0, 0, 0, time.UTC),
					ReportCount:       1,
					ReportedBy:        []string{"client1"},
					DataFormatVersion: "1.0",
				}
				return db.Create(&account).Error
			},
			expectedCode: http.StatusOK,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				body := w.Body.String()
				records := strings.Split(strings.TrimSpace(body), "\n")

				if len(records) != 2 {
					t.Errorf("Expected 2 CSV records (header + 1 account), got %d", len(records))
				}

				// CSV should properly escape commas
				if !strings.Contains(records[1], `"account_with,comma"`) {
					t.Errorf("Expected properly escaped account ID, got %s", records[1])
				}
				if !strings.Contains(records[1], `"Name,with,commas"`) {
					t.Errorf("Expected properly escaped account name, got %s", records[1])
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := newTestDB(t)

			if err := tt.setupDB(t, db); err != nil {
				t.Fatalf("Failed to setup test database: %v", err)
			}

			handler := NewHandler(db)

			req := httptest.NewRequest("GET", "/api/download", nil)
			w := httptest.NewRecorder()

			handler.DownloadCSVHandler(w, req)

			if w.Code != tt.expectedCode {
				t.Errorf("DownloadCSVHandler() status code = %v, want %v", w.Code, tt.expectedCode)
			}

			if tt.checkResponse != nil {
				tt.checkResponse(t, w)
			}
		})
	}
}
