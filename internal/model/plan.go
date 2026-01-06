package model

import (
	"time"
)

// ============================================================================
// CONSTANTS
// ============================================================================
const (
	ExpenseTypeFixed        = "FIXED"
	ExpenseTypeLoanFlat     = "LOAN_FLAT"
	ExpenseTypeLoanInterest = "LOAN_INTEREST"
	ExpenseTypePercentGross = "PERCENTAGE_OF_GROSS"
	ExpenseTypePercentNet   = "PERCENTAGE_OF_NET"
)

// ============================================================================
// DATABASE ENTITIES (TABLES)
// ============================================================================

// 1. Table: users
type User struct {
	ID        uint            `gorm:"primaryKey" json:"id"`
	Name      string          `gorm:"type:varchar(100)" json:"name"`
	Email     string          `gorm:"uniqueIndex;type:varchar(100)" json:"email"`
	Password  string          `json:"password,omitempty"`
	Expenses  []Expense       `json:"expenses,omitempty"`
	Plans     []FinancialPlan `json:"plans,omitempty"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
}

// 2. Table: expenses (Master Data)
type Expense struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	UserID        uint      `gorm:"index" json:"user_id"`
	Name          string    `gorm:"type:varchar(100)" json:"name"`
	Type          string    `gorm:"type:varchar(50)" json:"type"`
	DefaultAmount float64   `json:"default_amount"`
	DefaultRate   float64   `json:"default_rate"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// 3. Table: financial_plans (HEADER)
type FinancialPlan struct {
	ID     uint   `gorm:"primaryKey" json:"id"`
	UserID uint   `gorm:"index" json:"user_id"`
	Name   string `gorm:"type:varchar(150)" json:"name"`

	// Input Parameters (Disimpan sebagai JSON agar history input terjaga)
	StartDate      string         `json:"start_date"`
	InitialCapital float64        `json:"initial_capital"`
	TargetSavings  float64        `json:"target_savings"`
	IncomeStreams  []IncomeStream `gorm:"serializer:json" json:"income_streams"`
	ExpenseRules   []ExpenseRule  `gorm:"serializer:json" json:"expense_rules"`

	// RELASI RELATIONAL (Output Hasil Kalkulasi)
	// OnDelete:CASCADE -> Hapus Plan, otomatis hapus bulan & detailnya
	MonthlyProjections []MonthlyProjection `gorm:"foreignKey:PlanID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"monthly_projections"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// 4. Table: plan_monthly_projections (CHILD 1)
type MonthlyProjection struct {
	ID     uint `gorm:"primaryKey" json:"id"`
	PlanID uint `gorm:"index" json:"plan_id"` // FK

	MonthIndex int       `json:"month_index"`
	PeriodDate time.Time `json:"period_date"`

	// Summary Numbers
	MainIncome       float64 `json:"main_income"`
	IncomeNote       string  `gorm:"type:text" json:"income_note"`
	TotalExpense     float64 `json:"total_expense"`
	DisposableIncome float64 `json:"disposable_income"`
	SavingsBalance   float64 `json:"savings_balance"`

	// RELASI RELATIONAL (Detail Pengeluaran)
	ExpenseDetails []PlanExpenseDetail `gorm:"foreignKey:MonthlyProjectionID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"expense_details"`
}

// 5. Table: plan_expense_details (CHILD 2)
type PlanExpenseDetail struct {
	ID                  uint `gorm:"primaryKey" json:"id"`
	MonthlyProjectionID uint `gorm:"index" json:"monthly_projection_id"` // FK

	Name           string  `json:"name"`
	Amount         float64 `json:"amount"`
	CategoryType   string  `json:"category_type"`
	PercentageRate float64 `json:"percentage_rate"`
	Note           string  `gorm:"type:text" json:"note"`
}

// ============================================================================
// HELPER STRUCTS (Non-Table, for JSON Parsing & Logic)
// ============================================================================

type IncomeStream struct {
	StartMonth int     `json:"start_month"`
	EndMonth   int     `json:"end_month"`
	Amount     float64 `json:"amount"`
	Note       string  `json:"note"`
}

type ExpenseRule struct {
	Name           string  `json:"name"`
	Type           string  `json:"type"`
	Amount         float64 `json:"amount"`
	Rate           float64 `json:"rate"`
	InitialBalance float64 `json:"initial_balance"`
	StartMonth     int     `json:"start_month"`
	EndMonth       int     `json:"end_month"`
	Note           string  `json:"note"`
}

type ProjectionRequest struct {
	StartDate      string         `json:"start_date"`
	InitialCapital float64        `json:"initial_capital"`
	TargetSavings  float64        `json:"target_savings"`
	IncomeStreams  []IncomeStream `json:"income_streams"`
	ExpenseRules   []ExpenseRule  `json:"expense_rules"`
}

// Struct Logic Sederhana (dipakai saat kalkulasi sebelum di-convert ke DB Struct)
type ExpenseItem struct {
	Name           string  `json:"name"`
	Amount         float64 `json:"amount"`
	CategoryType   string  `json:"category_type"`
	PercentageRate float64 `json:"percentage_rate,omitempty"`
	Note           string  `json:"note"`
}
