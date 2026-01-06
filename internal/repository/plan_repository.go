package repository

import (
	"context"
	"money_plan/internal/model" // Sesuaikan module path

	"gorm.io/gorm"
)

type PlanRepository interface {
	Save(ctx context.Context, plan *model.FinancialPlan) error
	FindAll(ctx context.Context, userID uint) ([]model.FinancialPlan, error)
	FindByID(ctx context.Context, id uint) (*model.FinancialPlan, error)
	Delete(ctx context.Context, id uint) error
}

type planRepository struct {
	db *gorm.DB
}

func NewPlanRepository(db *gorm.DB) PlanRepository {
	return &planRepository{db}
}

func (r *planRepository) Save(ctx context.Context, plan *model.FinancialPlan) error {
	// GORM otomatis insert ke tabel anak (MonthlyProjections & ExpenseDetails)
	// jika struct-nya terisi dengan benar.
	return r.db.WithContext(ctx).Create(plan).Error
}

func (r *planRepository) FindAll(ctx context.Context, userID uint) ([]model.FinancialPlan, error) {
	var plans []model.FinancialPlan
	// Kita tidak load MonthlyProjections (Preload) disini agar ringan (Compact View)
	err := r.db.WithContext(ctx).
		Select("id, user_id, name, start_date, target_savings, created_at").
		Where("user_id = ?", userID).
		Order("created_at desc").
		Find(&plans).Error
	return plans, err
}

func (r *planRepository) FindByID(ctx context.Context, id uint) (*model.FinancialPlan, error) {
	var plan model.FinancialPlan
	// EAGER LOADING: Load Plan -> Load Bulan -> Load Detail Expense
	err := r.db.WithContext(ctx).
		Preload("MonthlyProjections.ExpenseDetails").
		Preload("MonthlyProjections").
		First(&plan, id).Error

	return &plan, err
}

func (r *planRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&model.FinancialPlan{}, id).Error
}
