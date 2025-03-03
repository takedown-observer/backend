package api

import (
	"encoding/csv"
	"encoding/json"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/takedown-observer/backend/models"
	"github.com/takedown-observer/backend/validation"
	"gorm.io/gorm"
)

type Handler struct {
	db *gorm.DB
}

func NewHandler(db *gorm.DB) *Handler {
	return &Handler{db: db}
}

// ReportHandler handles POST /api/report
func (h *Handler) ReportHandler(w http.ResponseWriter, r *http.Request) {
	var report models.ReportRequest
	if err := json.NewDecoder(r.Body).Decode(&report); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate client ID
	if !validation.ValidateUUID(report.ClientID) {
		http.Error(w, "Invalid client ID format", http.StatusBadRequest)
		return
	}

	// Validate data format version
	if report.DataFormatVersion != models.DataFormatVersion {
		http.Error(w, "Unsupported data format version", http.StatusBadRequest)
		return
	}

	// Validate and sanitize account data
	if err := validation.ValidateAccountID(report.Account.ID); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := validation.ValidateAccountName(report.Account.Name); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := validation.ValidateCountries(report.Account.Countries); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Sanitize input
	sanitizedAccount := models.Account{
		ID:                validation.SanitizeString(report.Account.ID),
		Name:              validation.SanitizeString(report.Account.Name),
		Countries:         make([]string, len(report.Account.Countries)),
		LastReportedAt:    time.Now(),
		DataFormatVersion: report.DataFormatVersion,
	}

	// Sanitize countries
	for i, country := range report.Account.Countries {
		sanitizedAccount.Countries[i] = validation.SanitizeString(country)
	}

	// Database transaction
	err := h.db.Transaction(func(tx *gorm.DB) error {
		var existingAccount models.Account
		result := tx.First(&existingAccount, "id = ?", sanitizedAccount.ID)

		if result.Error == gorm.ErrRecordNotFound {
			// New account
			sanitizedAccount.ReportCount = 1
			sanitizedAccount.ReportedBy = []string{report.ClientID}

			if err := tx.Create(&sanitizedAccount).Error; err != nil {
				return err
			}
		} else if result.Error != nil {
			return result.Error
		} else {
			// Update existing account
			reported := false
			for _, id := range existingAccount.ReportedBy {
				if id == report.ClientID {
					reported = true
					break
				}
			}

			if !reported {
				existingAccount.ReportCount++
				existingAccount.ReportedBy = append(existingAccount.ReportedBy, report.ClientID)
			}

			existingAccount.Name = sanitizedAccount.Name
			existingAccount.Countries = sanitizedAccount.Countries
			existingAccount.LastReportedAt = sanitizedAccount.LastReportedAt
			existingAccount.DataFormatVersion = sanitizedAccount.DataFormatVersion

			if err := tx.Save(&existingAccount).Error; err != nil {
				return err
			}
		}
		return nil
	})

	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

// GetAccountsHandler handles GET /api/accounts
func (h *Handler) GetAccountsHandler(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	country := r.URL.Query().Get("country")
	search := r.URL.Query().Get("search")

	pageSize := 20
	offset := (page - 1) * pageSize

	// Start building the query
	query := h.db.Model(&models.Account{})

	// Apply filters
	if country != "" {
		query = query.Where("JSON_EXTRACT(countries, '$') LIKE ?", "%"+country+"%")
	}

	if search != "" {
		query = query.Where("name LIKE ?", "%"+search+"%")
	}

	// Get total count with filters
	var totalCount int64
	query.Count(&totalCount)

	// Get filtered and paginated accounts
	var accounts []models.Account
	result := query.Order("last_reported_at desc").
		Limit(pageSize).
		Offset(offset).
		Find(&accounts)

	if result.Error != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	// Get unique countries (from all accounts, not just filtered)
	var uniqueCountries []string
	uniqueCountriesMap := make(map[string]bool)

	var allAccounts []models.Account
	if err := h.db.Find(&allAccounts).Error; err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	for _, account := range allAccounts {
		for _, country := range account.Countries {
			uniqueCountriesMap[country] = true
		}
	}

	for country := range uniqueCountriesMap {
		uniqueCountries = append(uniqueCountries, country)
	}
	sort.Strings(uniqueCountries)

	totalPages := int((totalCount + int64(pageSize) - 1) / int64(pageSize))

	response := models.AccountsResponse{
		Accounts:        accounts,
		TotalCount:      totalCount,
		CurrentPage:     page,
		TotalPages:      totalPages,
		UniqueCountries: uniqueCountries,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// DownloadCSVHandler handles GET /api/download
func (h *Handler) DownloadCSVHandler(w http.ResponseWriter, r *http.Request) {
	// Set headers for CSV download
	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", "attachment; filename=takedowns.csv")

	// Get all accounts ordered by last reported time
	var accounts []models.Account
	result := h.db.Order("last_reported_at desc").Find(&accounts)
	if result.Error != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	// Create CSV writer
	writer := csv.NewWriter(w)

	// Write header
	header := []string{"Account ID", "Username", "Countries", "Last Reported At", "Data Format Version"}
	if err := writer.Write(header); err != nil {
		http.Error(w, "Error writing CSV", http.StatusInternalServerError)
		return
	}

	// Write data
	for _, account := range accounts {
		// Join countries array with commas
		countriesStr := strings.Join(account.Countries, ", ")

		row := []string{
			account.ID,
			account.Name,
			countriesStr,
			account.LastReportedAt.Format(time.RFC3339),
			account.DataFormatVersion,
		}
		if err := writer.Write(row); err != nil {
			http.Error(w, "Error writing CSV", http.StatusInternalServerError)
			return
		}
	}

	// Flush the writer
	writer.Flush()

	if err := writer.Error(); err != nil {
		http.Error(w, "Error writing CSV", http.StatusInternalServerError)
		return
	}
}
