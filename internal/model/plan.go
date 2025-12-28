package model

import (
	"time"
)

// Input Request (Tidak berubah, karena ini kontrak JSON API)
type ProjectionRequest struct {
	StartDate       string         `json:"start_date"`
	InitialCapital  float64        `json:"initial_capital"`
	IncomeStreams   []IncomeStream `json:"income_streams"`
	ExpenseRules    []ExpenseRule  `json:"expense_rules"`
	TargetSavings   float64        `json:"target_savings"`
}

type IncomeStream struct {
	StartMonth int     `json:"start_month"`
	EndMonth   int     `json:"end_month"`
	Amount     float64 `json:"amount"`
	Note       string  `json:"note"`
}

type ExpenseRule struct {
	Name           string  `json:"name"`
	Type           string  `json:"type"` // FIXED, LOAN_INTEREST, etc
	Amount         float64 `json:"amount"`
	Rate           float64 `json:"rate"`
	InitialBalance float64 `json:"initial_balance"`
	StartMonth     int     `json:"start_month"`
	EndMonth       int     `json:"end_month"`
}

// ==========================================
// DATABASE MODEL (POSTGRESQL TABLES)
// ==========================================

// Table Name: financial_plans
type FinancialPlan struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    uint      `json:"user_id"` // Nanti diisi ID user dari JWT
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	
	// MAGIC HAPPENS HERE:
	// Kita simpan array struct ini sebagai JSONB di Postgres
	// GORM v2 otomatis handle serialization/deserialization
	Projections []MonthlyProjection `gorm:"serializer:json" json:"projections"`
}

// Struct pendukung (Tidak jadi tabel terpisah, tapi masuk ke dalam JSONB column)
type MonthlyProjection struct {
	PeriodDate       time.Time          `json:"period_date"`
	MonthIndex       int                `json:"month_index"`
	MainIncome       float64            `json:"main_income"`
	SideIncome       float64            `json:"side_income"`
	IncomeNote       string             `json:"income_note"`
	Expenses         []ExpenseItem      `json:"expenses"` // Nested JSON again
	TotalExpense     float64            `json:"total_expense"`
	DisposableIncome float64            `json:"disposable_income"`
	SavingsBalance   float64            `json:"savings_balance"`
	RemainingDebts   map[string]float64 `json:"remaining_debts"`
}

type ExpenseItem struct {
	Name           string  `json:"name"`
	Amount         float64 `json:"amount"`
	CategoryType   string  `json:"category_type"`
	PercentageRate float64 `json:"percentage_rate,omitempty"`
	Note           string  `json:"note"`
}