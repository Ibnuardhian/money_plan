package config

import (
	"fmt"
	"log"
	"os"

	"money_plan/internal/model" // Import model untuk auto-migrate

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func NewPostgresDatabase() *gorm.DB {
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=%s",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_SSLMODE"),
		os.Getenv("DB_TIMEZONE"),
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})

	if err != nil {
		log.Fatal("Failed to connect to PostgreSQL:", err)
	}

	// AUTO MIGRATE: Fitur ajaib GORM untuk membuat tabel otomatis
	// Pastikan model FinancialPlan didaftarkan disini
	log.Println("Running Auto Migrate...")
	err = db.AutoMigrate(&model.FinancialPlan{})
	if err != nil {
		log.Fatal("Migration failed:", err)
	}

	log.Println("✅ Connected to PostgreSQL & Migrated!")
	return db
}