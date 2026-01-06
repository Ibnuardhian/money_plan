package usecase

import (
	"context"
	"errors"
	"time"

	"money_plan/internal/model"
	"money_plan/internal/repository"
)

type ExpenseUsecase interface {
	CreateExpense(ctx context.Context, e *model.Expense) (*model.Expense, error)
	GetExpensesByUser(ctx context.Context, userID uint) ([]model.Expense, error)
	GetExpenseByID(ctx context.Context, id uint) (*model.Expense, error)
	UpdateExpense(ctx context.Context, e *model.Expense) (*model.Expense, error)
	DeleteExpense(ctx context.Context, id uint) error
}

type expenseUsecase struct {
	expRepo        repository.ExpenseRepository
	planRepo       repository.PlanRepository // added plan repo
	contextTimeout time.Duration
}

func NewExpenseUsecase(r repository.ExpenseRepository, p repository.PlanRepository) ExpenseUsecase {
	return &expenseUsecase{
		expRepo:        r,
		planRepo:       p,
		contextTimeout: time.Second * 5,
	}
}

func (u *expenseUsecase) CreateExpense(ctx context.Context, e *model.Expense) (*model.Expense, error) {
	if e.UserID == 0 || e.Name == "" || e.Type == "" {
		return nil, errors.New("user_id, name and type are required")
	}
	e.CreatedAt = time.Now()
	e.UpdatedAt = time.Now()

	// use timeout context for repo ops
	c, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()

	if err := u.expRepo.Save(c, e); err != nil {
		return nil, err
	}

	// After creating expense master, apply effect to existing plans of the user
	if err := u.applyExpenseToPlans(c, e); err != nil {
		// propagate error so caller knows something failed updating plans
		return e, err
	}

	return e, nil
}

// applyExpenseToPlans will append the new expense as recurring expense detail to every monthly projection
// and adjust totals/savings accordingly.
func (u *expenseUsecase) applyExpenseToPlans(ctx context.Context, e *model.Expense) error {
	// fetch all plans for the user
	plans, err := u.planRepo.FindAll(ctx, e.UserID)
	if err != nil {
		return err
	}

	for _, p := range plans {
		// ensure we have projections loaded; if repository returns plans without relations,
		// FindByID per plan could be used instead. Here we assume FindAll returns MonthlyProjections.
		changed := false
		for i := range p.MonthlyProjections {
			mp := &p.MonthlyProjections[i]

			// Create a new expense detail for this master expense (recurring)
			newDetail := model.PlanExpenseDetail{
				Name:         e.Name,
				Amount:       e.DefaultAmount,
				CategoryType: e.Type,
				Note:         "Applied from master expense",
			}

			// Append detail and adjust totals
			mp.ExpenseDetails = append(mp.ExpenseDetails, newDetail)
			mp.TotalExpense += newDetail.Amount
			mp.DisposableIncome -= newDetail.Amount

			// Decrease savings balance for this month and all subsequent months (cumulative effect)
			for j := i; j < len(p.MonthlyProjections); j++ {
				p.MonthlyProjections[j].SavingsBalance -= newDetail.Amount
			}

			changed = true
		}

		if changed {
			// persist changes for the plan; use Save to persist associations (assumes repo supports it)
			if err := u.planRepo.Save(ctx, &p); err != nil {
				return err
			}
		}
	}

	return nil
}

func (u *expenseUsecase) GetExpensesByUser(ctx context.Context, userID uint) ([]model.Expense, error) {
	c, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()
	return u.expRepo.FindAllByUser(c, userID)
}

func (u *expenseUsecase) GetExpenseByID(ctx context.Context, id uint) (*model.Expense, error) {
	c, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()
	return u.expRepo.FindByID(c, id)
}

func (u *expenseUsecase) UpdateExpense(ctx context.Context, e *model.Expense) (*model.Expense, error) {
	c, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()
	existing, err := u.expRepo.FindByID(c, e.ID)
	if err != nil {
		return nil, err
	}
	existing.Name = e.Name
	existing.Type = e.Type
	existing.DefaultAmount = e.DefaultAmount
	existing.DefaultRate = e.DefaultRate
	existing.UpdatedAt = time.Now()
	if err := u.expRepo.Update(c, existing); err != nil {
		return nil, err
	}
	return existing, nil
}

func (u *expenseUsecase) DeleteExpense(ctx context.Context, id uint) error {
	c, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()
	return u.expRepo.Delete(c, id)
}
