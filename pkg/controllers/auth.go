package controllers

import (
	jwtware "github.com/gofiber/contrib/v3/jwt"
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

type authTokenResponse struct {
	Token string `json:"token"`
}

func NewAuthController(repo models.Repo, secret string) *AuthController {
	return &AuthController{
		repo:   repo,
		secret: secret,
	}
}

// Register godoc
// @Summary Register a new user
// @Description Creates a new user account with username and password.
// @Tags auth
// @Accept json
// @Produce json
// @Param request body authRequest true "Registration request"
// @Success 201 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Router /auth/register [post]
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

// Login godoc
// @Summary Authenticate user
// @Description Authenticates a user and returns a JWT token.
// @Tags auth
// @Accept json
// @Produce json
// @Param request body authRequest true "Login request"
// @Success 200 {object} authTokenResponse
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /auth/login [post]
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

func (c *AuthController) Middleware(ctx fiber.Ctx) error {
	return jwtware.New(jwtware.Config{
		SigningKey: jwtware.SigningKey{Key: c.secret},
		ErrorHandler: func(c fiber.Ctx, err error) error {
			return c.Status(401).JSON(fiber.Map{
				"message": "unauthorized",
			})
		},
	})(ctx)
}

func generatePasswordHash(password string) string {
	hash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(hash)
}

func comparePasswords(hashedPassword, password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password)) == nil
}
