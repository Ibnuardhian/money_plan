package route

import (
	deliveryhttp "money_plan/internal/delivery/http"

	"github.com/gofiber/fiber/v2"
)

// SetupRoutes menerima fiber.App dan Handler/Controller
func SetupRoutes(app *fiber.App, planHandler *deliveryhttp.PlanHandler) {
	// 1. Route Global / Utility
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status": "ok",
			"app":    "money_plan",
		})
	})

	// 2. API Grouping (Best Practice)
	// Agar URL rapi, misal: /api/v1/plans
	api := app.Group("/api")
	v1 := api.Group("/v1")
	plans := v1.Group("/plans")

	// Dummy route
	v1.Get("/test", func(c *fiber.Ctx) error {
		return c.SendString("API v1 is working")
	})

	// Endpoint kalkulasi proyeksi
	plans.Post("/calculate", planHandler.CalculateProjection)
}
