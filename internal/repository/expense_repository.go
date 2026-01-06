package repository

import (
	"context"
	"money_plan/internal/model"

	"gorm.io/gorm"
)

type ExpenseRepository interface {
	Save(ctx context.Context, e *model.Expense) error
	FindAllByUser(ctx context.Context, userID uint) ([]model.Expense, error)
	FindByID(ctx context.Context, id uint) (*model.Expense, error)
	Update(ctx context.Context, e *model.Expense) error
	Delete(ctx context.Context, id uint) error
}

type expenseRepository struct {
	db *gorm.DB
}

func NewExpenseRepository(db *gorm.DB) ExpenseRepository {
	return &expenseRepository{db}
}

func (r *expenseRepository) Save(ctx context.Context, e *model.Expense) error {
	return r.db.WithContext(ctx).Create(e).Error
}

func (r *expenseRepository) FindAllByUser(ctx context.Context, userID uint) ([]model.Expense, error) {
	var items []model.Expense
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).Order("created_at desc").Find(&items).Error
	return items, err
}

func (r *expenseRepository) FindByID(ctx context.Context, id uint) (*model.Expense, error) {
	var item model.Expense
	err := r.db.WithContext(ctx).First(&item, id).Error
	return &item, err
}

func (r *expenseRepository) Update(ctx context.Context, e *model.Expense) error {
	return r.db.WithContext(ctx).Save(e).Error
}

func (r *expenseRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&model.Expense{}, id).Error
}
