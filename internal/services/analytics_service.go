package services

import (
	"fmt"

	"github.com/M4yankkkk/findash/internal/models"
	"github.com/M4yankkkk/findash/internal/repository"
)

// SummaryResult holds the top-level financial summary.
type SummaryResult struct {
	TotalIncome   float64 `json:"total_income"`
	TotalExpenses float64 `json:"total_expenses"`
	NetBalance    float64 `json:"net_balance"`
	EntryCount    int     `json:"entry_count"`
}

// CategoryResult holds aggregated data for a single category.
type CategoryResult struct {
	Category string  `json:"category"`
	Type     string  `json:"type"`
	Total    float64 `json:"total"`
	Count    int     `json:"count"`
}

// TrendResult holds aggregated data for a single month.
type TrendResult struct {
	Month         string  `json:"month"` // e.g. "2026-03"
	TotalIncome   float64 `json:"total_income"`
	TotalExpenses float64 `json:"total_expenses"`
	NetBalance    float64 `json:"net_balance"`
}

// AnalyticsService handles all analytics and aggregation logic.
type AnalyticsService struct {
	analyticsRepo *repository.AnalyticsRepository
}

// NewAnalyticsService creates a new AnalyticsService.
func NewAnalyticsService(analyticsRepo *repository.AnalyticsRepository) *AnalyticsService {
	return &AnalyticsService{analyticsRepo: analyticsRepo}
}

// GetSummary returns total income, expenses, and net balance.
// Admins see the global summary; others see only their own data.
func (s *AnalyticsService) GetSummary(userID string, role models.Role) (*SummaryResult, error) {
	rows, err := s.getSummaryRows(userID, role)
	if err != nil {
		return nil, fmt.Errorf("fetching summary: %w", err)
	}

	result := &SummaryResult{}
	for _, row := range rows {
		result.EntryCount += row.Count
		if row.Type == string(models.EntryTypeIncome) {
			result.TotalIncome = row.Total
		} else {
			result.TotalExpenses = row.Total
		}
	}
	result.NetBalance = result.TotalIncome - result.TotalExpenses

	return result, nil
}

// GetDashboardSummary returns summary cards scoped to the requester's dashboard visibility.
func (s *AnalyticsService) GetDashboardSummary(userID string, role models.Role) (*SummaryResult, error) {
	rows, err := s.getSummaryRows(userID, role)
	if err != nil {
		return nil, fmt.Errorf("fetching dashboard summary: %w", err)
	}

	result := &SummaryResult{}
	for _, row := range rows {
		result.EntryCount += row.Count
		if row.Type == string(models.EntryTypeIncome) {
			result.TotalIncome = row.Total
		} else {
			result.TotalExpenses = row.Total
		}
	}
	result.NetBalance = result.TotalIncome - result.TotalExpenses

	return result, nil
}

func (s *AnalyticsService) getSummaryRows(userID string, role models.Role) ([]repository.AggregateRow, error) {
	if role == models.RoleViewer {
		return s.analyticsRepo.AggregateSummaryForViewer(userID)
	}

	scopedUserID := userID
	if role == models.RoleAdmin {
		scopedUserID = ""
	}

	return s.analyticsRepo.AggregateSummary(scopedUserID)
}

// GetByCategory returns a breakdown of totals grouped by category and type.
// Admins see global data; others see only their own.
func (s *AnalyticsService) GetByCategory(userID string, role models.Role) ([]CategoryResult, error) {
	scopedUserID := userID
	if role == models.RoleAdmin {
		scopedUserID = ""
	}

	rows, err := s.analyticsRepo.AggregateByCategory(scopedUserID)
	if err != nil {
		return nil, fmt.Errorf("fetching category breakdown: %w", err)
	}

	results := make([]CategoryResult, len(rows))
	for i, row := range rows {
		results[i] = CategoryResult{
			Category: row.Category,
			Type:     row.Type,
			Total:    row.Total,
			Count:    row.Count,
		}
	}

	return results, nil
}

// GetTrend returns month-by-month income vs expense data for the last N months.
// Admins see global data; others see only their own.
func (s *AnalyticsService) GetTrend(userID string, role models.Role, months int) ([]TrendResult, error) {
	if months < 1 || months > 24 {
		months = 6 // sensible default for a dashboard chart
	}

	scopedUserID := userID
	if role == models.RoleAdmin {
		scopedUserID = ""
	}

	rows, err := s.analyticsRepo.AggregateMonthlyTrend(scopedUserID, months)
	if err != nil {
		return nil, fmt.Errorf("fetching trend: %w", err)
	}

	// Group by month — each month has two rows (income + expense)
	monthMap := make(map[string]*TrendResult)
	order := []string{}

	for _, row := range rows {
		if _, exists := monthMap[row.Month]; !exists {
			monthMap[row.Month] = &TrendResult{Month: row.Month}
			order = append(order, row.Month)
		}
		if row.Type == string(models.EntryTypeIncome) {
			monthMap[row.Month].TotalIncome = row.Total
		} else {
			monthMap[row.Month].TotalExpenses = row.Total
		}
		monthMap[row.Month].NetBalance = monthMap[row.Month].TotalIncome - monthMap[row.Month].TotalExpenses
	}

	results := make([]TrendResult, 0, len(order))
	for _, month := range order {
		results = append(results, *monthMap[month])
	}

	return results, nil
}
