package http

import (
	"strconv"
	"strings"

	"money_plan/internal/model"
	"money_plan/internal/usecase"

	"github.com/gofiber/fiber/v2"
)

type ExpenseHandler struct {
	ExpenseUsecase usecase.ExpenseUsecase
}

func NewExpenseHandler(u usecase.ExpenseUsecase) *ExpenseHandler {
	return &ExpenseHandler{ExpenseUsecase: u}
}

// Create POST /api/v1/expenses
func (h *ExpenseHandler) Create(c *fiber.Ctx) error {
	var req model.Expense
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Invalid body", "error": err.Error()})
	}
	req.Name = strings.TrimSpace(req.Name)
	created, err := h.ExpenseUsecase.CreateExpense(c.Context(), &req)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Create failed", "error": err.Error()})
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"data": created})
}

// List GET /api/v1/expenses?user_id={id}
func (h *ExpenseHandler) List(c *fiber.Ctx) error {
	userIDStr := c.Query("user_id")
	if userIDStr == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "user_id is required"})
	}
	uid, err := strconv.Atoi(userIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "invalid user_id", "error": err.Error()})
	}
	items, err := h.ExpenseUsecase.GetExpensesByUser(c.Context(), uint(uid))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Failed to fetch", "error": err.Error()})
	}
	return c.JSON(fiber.Map{"data": items})
}

// Get GET /api/v1/expenses/:id
func (h *ExpenseHandler) Get(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "invalid id", "error": err.Error()})
	}
	item, err := h.ExpenseUsecase.GetExpenseByID(c.Context(), uint(id))
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"message": "Not found", "error": err.Error()})
	}
	return c.JSON(fiber.Map{"data": item})
}

// Update PUT /api/v1/expenses/:id
func (h *ExpenseHandler) Update(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "invalid id", "error": err.Error()})
	}
	var req model.Expense
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Invalid body", "error": err.Error()})
	}
	req.ID = uint(id)
	updated, err := h.ExpenseUsecase.UpdateExpense(c.Context(), &req)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Update failed", "error": err.Error()})
	}
	return c.JSON(fiber.Map{"data": updated})
}

// Delete DELETE /api/v1/expenses/:id
func (h *ExpenseHandler) Delete(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "invalid id", "error": err.Error()})
	}
	if err := h.ExpenseUsecase.DeleteExpense(c.Context(), uint(id)); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Delete failed", "error": err.Error()})
	}
	return c.SendStatus(fiber.StatusNoContent)
}
