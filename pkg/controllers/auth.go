package controllers

import (
	"github.com/gofiber/fiber/v3"
	"github.com/tobyrushton/fantasy-nba/pkg/db/models"
	"golang.org/x/crypto/bcrypt"
)

type AuthController struct {
	repo models.Repo
}

type authRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func NewAuthController(repo models.Repo) *AuthController {
	return &AuthController{repo: repo}
}

func (c *AuthController) Register(ctx fiber.Ctx) error {
	var req authRequest
	if err := ctx.Bind().Body(&req); err != nil {
		return ctx.Status(400).JSON(fiber.Map{
			"message": err.Error(),
		})
	}

	_, err := c.repo.CreateUser(ctx.Context(), req.Username, generatePasswordHash(req.Password))
	if err != nil {
		return ctx.Status(400).JSON(fiber.Map{
			"message": err.Error(),
		})
	}

	return ctx.Status(201).JSON(fiber.Map{
		"message": "user created",
	})
}

func generatePasswordHash(password string) string {
	hash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(hash)
}
