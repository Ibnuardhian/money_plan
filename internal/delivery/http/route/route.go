package route

import (
	deliveryhttp "money_plan/internal/delivery/http"

	"github.com/gofiber/fiber/v2"
)

// SetupRoutes menerima fiber.App dan Handler/Controller
// Tambah parameter userHandler & expenseHandler agar endpoint tersedia
func SetupRoutes(app *fiber.App, planHandler *deliveryhttp.PlanHandler, userHandler *deliveryhttp.UserHandler, expenseHandler *deliveryhttp.ExpenseHandler) {
	// 1. Route Global / Utility
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status": "ok",
			"app":    "money_plan",
		})
	})

	// Tambahkan route register & login di root path supaya /users/register tidak 404
	app.Post("/users/register", userHandler.Register)
	app.Post("/users/login", userHandler.Login)

	// 2. API Grouping (Best Practice)
	// Agar URL rapi, misal: /api/v1/plans
	api := app.Group("/api")
	v1 := api.Group("/v1")
	plans := v1.Group("/plans")
	users := v1.Group("/users")
	expenses := v1.Group("/expenses")

	// Dummy route
	v1.Get("/test", func(c *fiber.Ctx) error {
		return c.SendString("API v1 is working")
	})

	// Endpoint kalkulasi proyeksi
	plans.Post("/calculate", planHandler.CalculateProjection)
	plans.Get("/", planHandler.GetUserPlans)
	plans.Get("/:id", planHandler.GetPlanDetail)
	plans.Delete("/:id", planHandler.DeletePlan)

	// Expense endpoints (use expenseHandler instance)
	expenses.Post("/", expenseHandler.Create)
	expenses.Get("/", expenseHandler.List)
	expenses.Get("/:id", expenseHandler.Get)
	expenses.Put("/:id", expenseHandler.Update)
	expenses.Delete("/:id", expenseHandler.Delete)

	// Endpoint user register & login di versi API
	users.Post("/register", userHandler.Register)
	users.Post("/login", userHandler.Login)
}
