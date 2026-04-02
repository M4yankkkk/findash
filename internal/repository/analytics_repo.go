package repository

import (
	"fmt"

	"github.com/M4yankkkk/findash/internal/database"
)

// AnalyticsRepository handles all aggregation queries for the analytics endpoints.
type AnalyticsRepository struct {
	db *database.DB
}

// NewAnalyticsRepository creates a new AnalyticsRepository.
func NewAnalyticsRepository(db *database.DB) *AnalyticsRepository {
	return &AnalyticsRepository{db: db}
}

// AggregateRow is a generic row returned from summary/category queries.
type AggregateRow struct {
	Type     string
	Category string
	Total    float64
	Count    int
}

// MonthlyRow is a row from the monthly trend query.
type MonthlyRow struct {
	Month string
	Type  string
	Total float64
}

// AggregateSummary returns total amount and count grouped by entry type (income/expense).
// If userID is empty, aggregates across all users (admin view).
func (r *AnalyticsRepository) AggregateSummary(userID string) ([]AggregateRow, error) {
	query := `
		SELECT
			type,
			'' AS category,
			COALESCE(SUM(amount), 0) AS total,
			COUNT(*) AS count
		FROM financial_entries
		WHERE deleted_at IS NULL`

	args := []interface{}{}
	if userID != "" {
		query += ` AND user_id = $1`
		args = append(args, userID)
	}

	query += ` GROUP BY type`

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("aggregate summary: %w", err)
	}
	defer rows.Close()

	var results []AggregateRow
	for rows.Next() {
		var row AggregateRow
		if err := rows.Scan(&row.Type, &row.Category, &row.Total, &row.Count); err != nil {
			return nil, fmt.Errorf("scan summary row: %w", err)
		}
		results = append(results, row)
	}

	return results, rows.Err()
}

// AggregateByCategory returns totals grouped by category and type, ordered by total descending.
// If userID is empty, aggregates across all users (admin view).
func (r *AnalyticsRepository) AggregateByCategory(userID string) ([]AggregateRow, error) {
	query := `
		SELECT
			type,
			category,
			COALESCE(SUM(amount), 0) AS total,
			COUNT(*) AS count
		FROM financial_entries
		WHERE deleted_at IS NULL`

	args := []interface{}{}
	if userID != "" {
		query += ` AND user_id = $1`
		args = append(args, userID)
	}

	query += ` GROUP BY type, category ORDER BY total DESC`

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("aggregate by category: %w", err)
	}
	defer rows.Close()

	var results []AggregateRow
	for rows.Next() {
		var row AggregateRow
		if err := rows.Scan(&row.Type, &row.Category, &row.Total, &row.Count); err != nil {
			return nil, fmt.Errorf("scan category row: %w", err)
		}
		results = append(results, row)
	}

	return results, rows.Err()
}

// AggregateMonthlyTrend returns monthly income and expense totals for the last N months.
// Results are ordered chronologically (oldest → newest) for chart rendering.
// If userID is empty, aggregates across all users (admin view).
func (r *AnalyticsRepository) AggregateMonthlyTrend(userID string, months int) ([]MonthlyRow, error) {
	query := `
		SELECT
			TO_CHAR(date, 'YYYY-MM') AS month,
			type,
			COALESCE(SUM(amount), 0) AS total
		FROM financial_entries
		WHERE deleted_at IS NULL
		  AND date >= DATE_TRUNC('month', NOW()) - INTERVAL '1 month' * $1`

	args := []interface{}{months - 1}
	argIdx := 2

	if userID != "" {
		query += fmt.Sprintf(` AND user_id = $%d`, argIdx)
		args = append(args, userID)
	}

	query += ` GROUP BY month, type ORDER BY month ASC`

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("aggregate monthly trend: %w", err)
	}
	defer rows.Close()

	var results []MonthlyRow
	for rows.Next() {
		var row MonthlyRow
		if err := rows.Scan(&row.Month, &row.Type, &row.Total); err != nil {
			return nil, fmt.Errorf("scan trend row: %w", err)
		}
		results = append(results, row)
	}

	return results, rows.Err()
}
