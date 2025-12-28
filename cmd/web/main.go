package main

import (
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/joho/godotenv"

	"money_plan/config"
	myHttp "money_plan/internal/delivery/http"
	route "money_plan/internal/delivery/http/route"
	"money_plan/internal/repository"
	"money_plan/internal/usecase"
)

func main() {
	_ = godotenv.Load()

	// 1. SETUP DATABASE (POSTGRESQL)
	// Function ini akan return *gorm.DB
	db := config.NewPostgresDatabase()

	// 2. SETUP LAYERS
	planRepo := repository.NewPlanRepository(db)
	planUsecase := usecase.NewPlanUsecase(planRepo)
	planHandler := myHttp.NewPlanHandler(planUsecase)

	// 3. SERVER
	app := fiber.New()
	app.Use(logger.New())
	app.Use(cors.New())

	route.SetupRoutes(app, planHandler)

	// 4. START
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}
	log.Printf("Starting server on :%s", port)
	if err := app.Listen(":" + port); err != nil {
		log.Fatal(err)
	}
}
