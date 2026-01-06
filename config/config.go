package config

import (
	"fmt"
	"log"
	"os"

	migration "money_plan/db/migration"

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

	// Jalankan migration pusat yang mendaftarkan semua model
	log.Println("Running Auto Migrate (all models)...")
	migration.RunMigration(db)

	log.Println("✅ Connected to PostgreSQL & Migrated!")
	return db
}
