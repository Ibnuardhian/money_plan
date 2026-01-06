package usecase

import (
	"context"
	"fmt"
	"math"
	"time"

	"money_plan/internal/model" // Sesuaikan module path
	"money_plan/internal/repository"
)

type PlanUsecase interface {
	GenerateAndSaveProjection(ctx context.Context, req model.ProjectionRequest) (*model.FinancialPlan, error)
	GetUserPlans(ctx context.Context, userID uint) ([]model.FinancialPlan, error)
	GetPlanDetail(ctx context.Context, planID uint) (*model.FinancialPlan, error)
	DeletePlan(ctx context.Context, planID uint) error
}

type planUsecase struct {
	planRepo       repository.PlanRepository
	contextTimeout time.Duration
}

func NewPlanUsecase(repo repository.PlanRepository) PlanUsecase {
	return &planUsecase{
		planRepo:       repo,
		contextTimeout: time.Second * 10,
	}
}

func (u *planUsecase) GetUserPlans(ctx context.Context, userID uint) ([]model.FinancialPlan, error) {
	c, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()
	return u.planRepo.FindAll(c, userID)
}

func (u *planUsecase) GetPlanDetail(ctx context.Context, planID uint) (*model.FinancialPlan, error) {
	c, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()
	return u.planRepo.FindByID(c, planID)
}

func (u *planUsecase) DeletePlan(ctx context.Context, planID uint) error {
	c, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()
	// Optionally check existence first
	_, err := u.planRepo.FindByID(c, planID)
	if err != nil {
		return err
	}
	return u.planRepo.Delete(c, planID)
}

// LOGIC UTAMA
func (u *planUsecase) GenerateAndSaveProjection(ctx context.Context, req model.ProjectionRequest) (*model.FinancialPlan, error) {
	c, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()

	// 1. Setup Initial Variables
	startDate, err := time.Parse("2006-01-02", req.StartDate)
	if err != nil {
		startDate = time.Now()
	}
	currentSavings := req.InitialCapital
	debtBalances := make(map[string]float64)

	// Setup Saldo Awal Hutang
	for _, rule := range req.ExpenseRules {
		if (rule.Type == model.ExpenseTypeLoanFlat || rule.Type == model.ExpenseTypeLoanInterest) && rule.InitialBalance > 0 {
			debtBalances[rule.Name] = rule.InitialBalance
		}
	}

	monthsToProject := 24

	// Variable penampung untuk Database Rows
	var dbMonthlyProjections []model.MonthlyProjection

	// 2. Loop Kalkulasi
	for i := 0; i < monthsToProject; i++ {
		currentDate := startDate.AddDate(0, i, 0)
		currentMonthNum := i + 1

		// --- A. Hitung Income ---
		monthlyIncome := 0.0
		incomeNote := ""
		incomeFound := false
		for _, stream := range req.IncomeStreams {
			if currentMonthNum >= stream.StartMonth && currentMonthNum <= stream.EndMonth {
				monthlyIncome = stream.Amount
				incomeNote = stream.Note
				incomeFound = true
				break
			}
		}
		if !incomeFound && len(req.IncomeStreams) > 0 {
			lastStream := req.IncomeStreams[len(req.IncomeStreams)-1]
			monthlyIncome = lastStream.Amount
			incomeNote = "Extrapolated"
		}
		totalIncome := monthlyIncome

		// --- B. Hitung Expenses (2 Pass) ---
		var logicExpenses []model.ExpenseItem // Struct sementara
		totalFixedAndDebt := 0.0

		// Pass 1: Fixed & Loans
		for _, rule := range req.ExpenseRules {
			if rule.StartMonth > 0 && currentMonthNum < rule.StartMonth {
				continue
			}
			if rule.EndMonth > 0 && currentMonthNum > rule.EndMonth {
				continue
			}
			if rule.Type == model.ExpenseTypePercentGross || rule.Type == model.ExpenseTypePercentNet {
				continue
			}

			expenseAmount := 0.0
			note := rule.Note
			categoryType := rule.Type

			switch rule.Type {
			case model.ExpenseTypeFixed:
				expenseAmount = rule.Amount
			case model.ExpenseTypeLoanFlat:
				remaining := debtBalances[rule.Name]
				if remaining > 0 {
					payment := rule.Amount
					if remaining < payment {
						payment = remaining
					}
					expenseAmount = payment
					debtBalances[rule.Name] = remaining - payment
					note = fmt.Sprintf("Sisa hutang: %.0f", debtBalances[rule.Name])
				}
			case model.ExpenseTypeLoanInterest:
				remaining := debtBalances[rule.Name]
				if remaining > 0 {
					interest := remaining * (rule.Rate / 100 / 12)
					totalPayment := rule.Amount
					if totalPayment <= interest {
						totalPayment = interest
						note = "Interest only warning"
					}
					expenseAmount = totalPayment
					principalPaid := totalPayment - interest
					newBalance := remaining - principalPaid
					if newBalance < 0 {
						newBalance = 0
					}
					debtBalances[rule.Name] = newBalance
					categoryType = "DEBT_REPAYMENT"
					note = fmt.Sprintf("Bunga: %.0f, Pokok: %.0f, Sisa: %.0f", interest, principalPaid, newBalance)
				}
			}

			if expenseAmount > 0 {
				logicExpenses = append(logicExpenses, model.ExpenseItem{
					Name: rule.Name, Amount: math.Round(expenseAmount), CategoryType: categoryType, Note: note,
				})
				totalFixedAndDebt += expenseAmount
			}
		}

		// Pass 2: Percentage
		totalPercentageExpense := 0.0
		for _, rule := range req.ExpenseRules {
			if rule.StartMonth > 0 && currentMonthNum < rule.StartMonth {
				continue
			}
			if rule.EndMonth > 0 && currentMonthNum > rule.EndMonth {
				continue
			}

			expenseAmount := 0.0
			note := rule.Note

			if rule.Type == model.ExpenseTypePercentGross {
				expenseAmount = totalIncome * (rule.Rate / 100)
				note = fmt.Sprintf("%.1f%% dari Gross", rule.Rate)
			} else if rule.Type == model.ExpenseTypePercentNet {
				netBase := totalIncome - totalFixedAndDebt
				if netBase > 0 {
					expenseAmount = netBase * (rule.Rate / 100)
					note = fmt.Sprintf("%.1f%% dari Netto", rule.Rate)
				}
			}

			if expenseAmount > 0 {
				logicExpenses = append(logicExpenses, model.ExpenseItem{
					Name: rule.Name, Amount: math.Round(expenseAmount), CategoryType: rule.Type, PercentageRate: rule.Rate, Note: note,
				})
				totalPercentageExpense += expenseAmount
			}
		}

		// --- C. Summary ---
		totalExpenseThisMonth := totalFixedAndDebt + totalPercentageExpense
		disposableIncome := totalIncome - totalExpenseThisMonth
		currentSavings += disposableIncome

		// 3. MAPPING KE DB STRUCT (RELATIONAL)
		// Convert Logic Item ke DB Table Row
		var dbExpenseRows []model.PlanExpenseDetail
		for _, lexp := range logicExpenses {
			dbExpenseRows = append(dbExpenseRows, model.PlanExpenseDetail{
				Name:           lexp.Name,
				Amount:         lexp.Amount,
				CategoryType:   lexp.CategoryType,
				PercentageRate: lexp.PercentageRate,
				Note:           lexp.Note,
			})
		}

		// Create Monthly Row
		dbMonthlyProjections = append(dbMonthlyProjections, model.MonthlyProjection{
			MonthIndex:       i,
			PeriodDate:       currentDate,
			MainIncome:       math.Round(monthlyIncome),
			IncomeNote:       incomeNote,
			TotalExpense:     math.Round(totalExpenseThisMonth),
			DisposableIncome: math.Round(disposableIncome),
			SavingsBalance:   math.Round(currentSavings),
			ExpenseDetails:   dbExpenseRows, // GORM akan otomatis insert relasi ini
		})
	}

	// 4. Build Parent Object & Save
	finalPlan := &model.FinancialPlan{
		UserID:         1, // Hardcode/TODO from context
		Name:           "Plan " + req.StartDate,
		StartDate:      req.StartDate,
		InitialCapital: req.InitialCapital,
		TargetSavings:  req.TargetSavings,
		IncomeStreams:  req.IncomeStreams, // Simpan Config Input (JSON)
		ExpenseRules:   req.ExpenseRules,  // Simpan Config Input (JSON)

		// Hasil Kalkulasi (Relational Tables)
		MonthlyProjections: dbMonthlyProjections,

		CreatedAt: time.Now(),
	}

	err = u.planRepo.Save(c, finalPlan)
	if err != nil {
		return nil, err
	}

	return finalPlan, nil
}
