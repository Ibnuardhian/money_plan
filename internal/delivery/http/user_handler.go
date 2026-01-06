package http

import (
	"encoding/json"
	"log"
	"strings"

	"money_plan/internal/model"
	"money_plan/internal/usecase"

	"github.com/gofiber/fiber/v2"
)

// UserHandler menyimpan referensi ke usecase user
type UserHandler struct {
	UserUsecase usecase.UserUsecase
}

func NewUserHandler(u usecase.UserUsecase) *UserHandler {
	return &UserHandler{UserUsecase: u}
}

type RegisterRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

// Register menangani POST /users/register dan POST /api/v1/users/register
func (h *UserHandler) Register(c *fiber.Ctx) error {
	var req RegisterRequest

	// Coba BodyParser terlebih dahulu (butuh Content-Type: application/json)
	if err := c.BodyParser(&req); err != nil {
		// Fallback: parse raw body regardless of headers
		body := c.Body()
		log.Println("DEBUG: BodyParser failed:", err)
		log.Println("DEBUG: Raw request body:", string(body)) // debug output

		if len(body) == 0 {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"message": "Invalid request body",
				"error":   "empty body - make sure request has JSON payload",
			})
		}
		if err2 := json.Unmarshal(body, &req); err2 != nil {
			// Return parsing error detail so client can fix payload
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"message": "Invalid request body - JSON parse error",
				"error":   err2.Error(),
			})
		}
	}

	// Trim input
	req.Name = strings.TrimSpace(req.Name)
	req.Email = strings.TrimSpace(req.Email)
	req.Password = strings.TrimSpace(req.Password)

	// Basic validation here to give clearer error messages
	if req.Name == "" || req.Email == "" || req.Password == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Register failed",
			"error":   "name, email and password are required",
		})
	}

	// Map to model.User
	user := &model.User{
		Name:     req.Name,
		Email:    req.Email,
		Password: req.Password,
	}

	// Use request context so timeout/cancel works
	created, err := h.UserUsecase.Register(c.Context(), user)
	if err != nil {
		log.Println("DEBUG: Register usecase error:", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Register failed",
			"error":   err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "User created successfully",
		"data":    created,
	})
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// Login menangani POST /users/login dan POST /api/v1/users/login
func (h *UserHandler) Login(c *fiber.Ctx) error {
	var req LoginRequest
	if err := c.BodyParser(&req); err != nil {
		// fallback parse raw body
		body := c.Body()
		log.Println("DEBUG: BodyParser failed in Login:", err)
		if len(body) == 0 {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"message": "Invalid request body",
				"error":   "empty body - make sure request has JSON payload",
			})
		}
		if err2 := json.Unmarshal(body, &req); err2 != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"message": "Invalid request body - JSON parse error",
				"error":   err2.Error(),
			})
		}
	}

	req.Email = strings.TrimSpace(req.Email)
	req.Password = strings.TrimSpace(req.Password)

	if req.Email == "" || req.Password == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Login failed",
			"error":   "email and password are required",
		})
	}

	token, user, err := h.UserUsecase.Login(c.Context(), req.Email, req.Password)
	if err != nil {
		log.Println("DEBUG: Login usecase error:", err)
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "Login failed",
			"error":   err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Login successful",
		"token":   token,
		"user":    user,
	})
}
