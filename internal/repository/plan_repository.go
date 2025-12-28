package repository

import (
	"context"
	"money_plan/internal/model"
	"gorm.io/gorm"
)

type PlanRepository interface {
	Save(ctx context.Context, plan *model.FinancialPlan) error
	FindByID(ctx context.Context, id uint) (*model.FinancialPlan, error)
}

type planRepository struct {
	db *gorm.DB
}

func NewPlanRepository(db *gorm.DB) PlanRepository {
	return &planRepository{db}
}

// Save menyimpan plan baru ke Postgres
func (r *planRepository) Save(ctx context.Context, plan *model.FinancialPlan) error {
	// GORM support context untuk timeout cancellation
	// .Create() otomatis generate SQL INSERT
	result := r.db.WithContext(ctx).Create(plan)
	return result.Error
}

// Contoh fungsi Find (Opsional)
func (r *planRepository) FindByID(ctx context.Context, id uint) (*model.FinancialPlan, error) {
	var plan model.FinancialPlan
	// GORM otomatis convert kolom JSONB kembali ke struct Go
	err := r.db.WithContext(ctx).First(&plan, id).Error
	return &plan, err
}