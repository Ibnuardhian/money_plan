package http

import (
	"strconv"

	"money_plan/internal/model"   // Ganti sesuai module name
	"money_plan/internal/usecase" // Ganti sesuai module name

	"github.com/gofiber/fiber/v2"
)

// PlanHandler struct menyimpan referensi ke Logic Bisnis (Usecase)
type PlanHandler struct {
	PlanUsecase usecase.PlanUsecase
}

// NewPlanHandler adalah constructor untuk menginisialisasi Handler
func NewPlanHandler(usecase usecase.PlanUsecase) *PlanHandler {
	return &PlanHandler{
		PlanUsecase: usecase,
	}
}

// CalculateProjection menangani POST /api/v1/plans/calculate
func (h *PlanHandler) CalculateProjection(c *fiber.Ctx) error {
	// 1. Siapkan variable untuk menampung input user
	var request model.ProjectionRequest

	// 2. Parse JSON body dari request ke struct
	if err := c.BodyParser(&request); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid request body",
			"error":   err.Error(),
		})
	}

	// 3. Panggil Usecase (Otak aplikasi) untuk menghitung
	// Context diambil dari Fiber ctx agar support timeout/cancellation
	result, err := h.PlanUsecase.GenerateAndSaveProjection(c.Context(), request)

	if err != nil {
		// Jika ada error saat kalkulasi atau save ke DB
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Failed to generate projection",
			"error":   err.Error(),
		})
	}

	// 4. Return hasil sukses dengan HTTP 201 (Created)
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "Projection created successfully",
		"data":    result,
	})
}

// GetUserPlans handles GET /api/v1/plans?user_id={id}
func (h *PlanHandler) GetUserPlans(c *fiber.Ctx) error {
	userIDStr := c.Query("user_id")
	if userIDStr == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "user_id query param is required",
		})
	}
	uid, err := strconv.Atoi(userIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "invalid user_id",
			"error":   err.Error(),
		})
	}

	plans, err := h.PlanUsecase.GetUserPlans(c.Context(), uint(uid))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Failed to get plans",
			"error":   err.Error(),
		})
	}
	return c.JSON(fiber.Map{"data": plans})
}

// GetPlanDetail handles GET /api/v1/plans/:id
func (h *PlanHandler) GetPlanDetail(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "invalid plan id",
			"error":   err.Error(),
		})
	}

	plan, err := h.PlanUsecase.GetPlanDetail(c.Context(), uint(id))
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"message": "Plan not found",
			"error":   err.Error(),
		})
	}
	return c.JSON(fiber.Map{"data": plan})
}

// DeletePlan handles DELETE /api/v1/plans/:id
func (h *PlanHandler) DeletePlan(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "invalid plan id",
			"error":   err.Error(),
		})
	}

	if err := h.PlanUsecase.DeletePlan(c.Context(), uint(id)); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Failed to delete plan",
			"error":   err.Error(),
		})
	}
	return c.SendStatus(fiber.StatusNoContent)
}
