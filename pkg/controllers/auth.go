package controllers

import (
	"github.com/gofiber/fiber/v3"
	"github.com/tobyrushton/fantasy-nba/pkg/db/models"
	"github.com/tobyrushton/fantasy-nba/pkg/token"
	"golang.org/x/crypto/bcrypt"
)

type AuthController struct {
	repo models.Repo

	secret string
}

type authRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func NewAuthController(repo models.Repo, secret string) *AuthController {
	return &AuthController{
		repo:   repo,
		secret: secret,
	}
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

func (c *AuthController) Login(ctx fiber.Ctx) error {
	var req authRequest
	if err := ctx.Bind().Body(&req); err != nil {
		return ctx.Status(400).JSON(fiber.Map{
			"message": err.Error(),
		})
	}

	user, err := c.repo.GetUserByUsername(ctx.Context(), req.Username)
	if err != nil {
		return ctx.Status(400).JSON(fiber.Map{
			"message": "invalid username or password",
		})
	}

	if !comparePasswords(user.PasswordHash, req.Password) {
		return ctx.Status(400).JSON(fiber.Map{
			"message": "invalid username or password",
		})
	}

	token, err := token.GenerateToken(c.secret, user.ID)
	if err != nil {
		return ctx.Status(500).JSON(fiber.Map{
			"message": "failed to generate token",
		})
	}

	return ctx.Status(200).JSON(fiber.Map{
		"token": token,
	})
}

func generatePasswordHash(password string) string {
	hash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(hash)
}

func comparePasswords(hashedPassword, password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password)) == nil
}
