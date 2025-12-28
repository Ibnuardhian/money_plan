package usecase

import (
	"context"
	"fmt"
	"math"
	"time"

	"money_plan/internal/model"      // GANTI: Sesuaikan dengan module name di go.mod
	"money_plan/internal/repository" // GANTI: Sesuaikan dengan module name di go.mod
)

type PlanUsecase interface {
	GenerateAndSaveProjection(ctx context.Context, req model.ProjectionRequest) (*model.FinancialPlan, error)
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

// =================================================================================
// CORE LOGIC: CALCULATION ENGINE
// =================================================================================

func (u *planUsecase) GenerateAndSaveProjection(ctx context.Context, req model.ProjectionRequest) (*model.FinancialPlan, error) {
	c, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()

	// 1. SETUP INITIAL VARIABLES
	var projections []model.MonthlyProjection

	// Parsing Tanggal Mulai
	startDate, err := time.Parse("2006-01-02", req.StartDate)
	if err != nil {
		startDate = time.Now()
	}

	// Saldo Tabungan Berjalan
	currentSavings := req.InitialCapital

	// Setup Saldo Awal Hutang (Snapshot status hutang saat start)
	// Key: Nama Rule, Value: Sisa Hutang
	debtBalances := make(map[string]float64)
	for _, rule := range req.ExpenseRules {
		// Hanya ambil rule bertipe pinjaman yang punya saldo awal
		if (rule.Type == "LOAN_FLAT" || rule.Type == "LOAN_INTEREST") && rule.InitialBalance > 0 {
			debtBalances[rule.Name] = rule.InitialBalance
		}
	}

	// Durasi Proyeksi (Default 24 Bulan)
	monthsToProject := 24

	// 2. MAIN LOOP (ITERASI BULAN DEMI BULAN)
	for i := 0; i < monthsToProject; i++ {
		currentDate := startDate.AddDate(0, i, 0)
		monthIndex := i
		currentMonthNum := i + 1 // 1, 2, 3... (Human friendly index)

		// --- A. HITUNG INCOME (BULK / STREAMS LOGIC) ---
		monthlyIncome := 0.0
		incomeNote := ""

		// Cari income yang range bulannya cocok dengan bulan ini
		incomeFound := false
		for _, stream := range req.IncomeStreams {
			if currentMonthNum >= stream.StartMonth && currentMonthNum <= stream.EndMonth {
				monthlyIncome = stream.Amount
				incomeNote = stream.Note
				incomeFound = true
				break
			}
		}
		// Fallback: Jika tidak ada definisi income di bulan tua (misal bulan 25), pakai income terakhir
		if !incomeFound && len(req.IncomeStreams) > 0 {
			lastStream := req.IncomeStreams[len(req.IncomeStreams)-1]
			monthlyIncome = lastStream.Amount
			incomeNote = "Extrapolated from last stream"
		}

		totalIncome := monthlyIncome // + SideIncome (future feature)

		// --- B. HITUNG EXPENSES (DYNAMIC RULES & TIME FILTER) ---
		var monthlyExpenses []model.ExpenseItem
		totalExpenseThisMonth := 0.0

		for _, rule := range req.ExpenseRules {
			// [LOGIC BARU] Filter Waktu: Skip jika bulan ini diluar range Start-End rule
			// Jika StartMonth/EndMonth 0, dianggap berlaku selamanya.
			if rule.StartMonth > 0 && currentMonthNum < rule.StartMonth {
				continue // Belum mulai
			}
			if rule.EndMonth > 0 && currentMonthNum > rule.EndMonth {
				continue // Sudah berakhir
			}

			expenseAmount := 0.0
			note := ""
			categoryType := rule.Type

			switch rule.Type {
			case "FIXED":
				// Pengeluaran Tetap (Makan, Listrik, Kursus)
				expenseAmount = rule.Amount

			case "PERCENTAGE_OF_INCOME":
				// Pengeluaran Persentase (Zakat, Pajak, Sedekah)
				// Rumus: Income * (Rate / 100)
				expenseAmount = totalIncome * (rule.Rate / 100)
				note = fmt.Sprintf("Auto: %.1f%% of Income", rule.Rate)

			case "LOAN_FLAT":
				// Hutang Flat / Bunga 0% (Cicilan Teman, KPR Syariah Fixed)
				remaining := debtBalances[rule.Name]
				if remaining > 0 {
					payment := rule.Amount
					if remaining < payment {
						payment = remaining
					}
					expenseAmount = payment

					// Kurangi Hutang
					debtBalances[rule.Name] = remaining - payment
					note = fmt.Sprintf("Remaining: %.0f", debtBalances[rule.Name])
				} else {
					expenseAmount = 0
					note = "Lunas"
				}

			case "LOAN_INTEREST":
				// Hutang Berbunga Menurun (Paylater, CC)
				remaining := debtBalances[rule.Name]
				if remaining > 0 {
					// 1. Hitung Bunga Bulan Ini (Sisa * RatePertahun / 12)
					interest := remaining * (rule.Rate / 100 / 12)

					// 2. Total Bayar dari User
					totalPayment := rule.Amount

					// Validasi: Bayar minimal bunga
					if totalPayment <= interest {
						totalPayment = interest
						note = "Warning: Payment covers interest only!"
					}

					expenseAmount = totalPayment

					// 3. Kurangi Pokok
					principalPaid := totalPayment - interest

					// 4. Update Saldo
					newBalance := remaining - principalPaid
					if newBalance < 0 {
						newBalance = 0
					}

					debtBalances[rule.Name] = newBalance

					note = fmt.Sprintf("Int: %.0f, Principal: %.0f, Rem: %.0f", interest, principalPaid, newBalance)
					categoryType = "DEBT_REPAYMENT"
				} else {
					expenseAmount = 0
					note = "Lunas"
				}
			}

			// Masukkan ke List Expense jika > 0
			// Kita round agar tampilan di JSON bersih (tanpa koma panjang)
			if expenseAmount > 0 || rule.Type == "FIXED" {
				monthlyExpenses = append(monthlyExpenses, model.ExpenseItem{
					Name:           rule.Name,
					Amount:         math.Round(expenseAmount),
					CategoryType:   categoryType,
					PercentageRate: rule.Rate,
					Note:           note,
				})
				totalExpenseThisMonth += expenseAmount
			}
		}

		// --- C. KALKULASI SUMMARY BULANAN ---

		disposableIncome := totalIncome - totalExpenseThisMonth
		currentSavings += disposableIncome

		// SNAPSHOT HUTANG (Deep Copy Map)
		// Penting: Agar history sisa hutang tercatat per bulan, bukan nilai akhir saja.
		currentDebtsSnapshot := make(map[string]float64)
		for k, v := range debtBalances {
			if v > 100 { // Sembunyikan sisa receh float residue
				currentDebtsSnapshot[k] = math.Round(v)
			}
		}

		// Object Proyeksi Bulan Ini
		proj := model.MonthlyProjection{
			PeriodDate:       currentDate,
			MonthIndex:       monthIndex,
			MainIncome:       math.Round(monthlyIncome),
			SideIncome:       0,
			IncomeNote:       incomeNote, // Pastikan field ini ada di struct model
			Expenses:         monthlyExpenses,
			TotalExpense:     math.Round(totalExpenseThisMonth),
			DisposableIncome: math.Round(disposableIncome),
			SavingsBalance:   math.Round(currentSavings),
			RemainingDebts:   currentDebtsSnapshot,
		}

		projections = append(projections, proj)
	}

	// 3. BUILD & SAVE DOCUMENT
	finalPlan := &model.FinancialPlan{
		// ID: Tidak perlu diisi, Postgres akan auto-increment (1, 2, 3...)
		UserID:      1, // Dummy User ID (karena tipe field uint)
		Name:        "Rencana Keuangan - " + req.StartDate,
		CreatedAt:   time.Now(),
		Projections: projections, // Struct ini akan otomatis jadi JSON oleh GORM
	}

	// Simpan ke Database
	err = u.planRepo.Save(c, finalPlan)
	if err != nil {
		return nil, err
	}

	return finalPlan, nil
}
