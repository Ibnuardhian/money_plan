package repository

import (
	"context"
	"encoding/json"
	"money_plan/internal/model"

	"github.com/jackc/pgx/v5/pgxpool"
)

type PlanRepository interface {
	Save(ctx context.Context, plan *model.FinancialPlan) error
	// Method lain seperti FindByID, Update, dll.
}

type planRepository struct {
	pool *pgxpool.Pool
}

func NewPlanRepository(pool *pgxpool.Pool) PlanRepository {
	return &planRepository{pool: pool}
}

func (r *planRepository) Save(ctx context.Context, plan *model.FinancialPlan) error {
	// Simpan seluruh dokumen sebagai JSONB agar logic tetap sama
	payload, err := json.Marshal(plan)
	if err != nil {
		return err
	}

	// Tabel: plans(id, user_id, name, created_at, data)
	const q = `
		INSERT INTO plans (id, user_id, name, created_at, data)
		VALUES ($1, $2, $3, $4, $5)
	`
	_, err = r.pool.Exec(ctx, q, plan.ID, plan.UserID, plan.Name, plan.CreatedAt, payload)
	return err
}
