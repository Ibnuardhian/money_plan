package db

import (
	"log"
	"money_plan/internal/model" // GANTI dengan nama module Anda

	"gorm.io/gorm"
)

// RunMigration menjalankan AutoMigrate GORM
func RunMigration(db *gorm.DB) {
	log.Println("Running Database Migration...")

	// ---------------------------------------------------------
	// OPTIONAL: Reset Database (HATI-HATI: Menghapus semua data)
	// Gunakan ini jika Anda ingin tabel dibuat ulang dari nol
	// karena perubahan struktur drastis dari JSON ke Relational.
	// ---------------------------------------------------------
	// resetDatabase(db) 
	
	// ---------------------------------------------------------
	// UTAMA: Auto Migrate
	// GORM akan membuat tabel, kolom, index, dan foreign key
	// berdasarkan struct di package model.
	// ---------------------------------------------------------
	err := db.AutoMigrate(
		&model.User{},              // Table: users
		&model.Expense{},           // Table: expenses
		&model.FinancialPlan{},     // Table: financial_plans (Header)
		&model.MonthlyProjection{}, // Table: plan_monthly_projections (Child 1)
		&model.PlanExpenseDetail{}, // Table: plan_expense_details (Child 2)
	)

	if err != nil {
		log.Fatal("❌ Migration Failed:", err)
	}

	log.Println("✅ Migration Success! All tables created/updated.")
}

// resetDatabase menghapus tabel lama agar fresh (Dev Mode Only)
func resetDatabase(db *gorm.DB) {
	log.Println("⚠️  WARNING: Dropping all tables for Fresh Migration...")
	
	// Drop urutan dibalik (Child dulu baru Parent) agar tidak kena FK Constraint
	db.Migrator().DropTable(&model.PlanExpenseDetail{})
	db.Migrator().DropTable(&model.MonthlyProjection{})
	db.Migrator().DropTable(&model.FinancialPlan{})
	db.Migrator().DropTable(&model.Expense{})
	// db.Migrator().DropTable(&model.User{}) // User mungkin jangan dihapus kalau males register lagi
}