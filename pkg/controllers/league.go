package controllers

import (
	"github.com/gofiber/fiber/v3"
	"github.com/tobyrushton/fantasy-nba/pkg/db/models"
)

type LeagueController struct {
	repo models.Repo
}

type createLeagueRequest struct {
	Name   string `json:"name"`
	UserID int64  `json:"user_id"`
}

func NewLeagueController(repo models.Repo) *LeagueController {
	return &LeagueController{repo: repo}
}

func (c *LeagueController) CreateLeague(ctx fiber.Ctx) error {
	var req createLeagueRequest
	if err := ctx.Bind().Body(&req); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	league, err := c.repo.CreateLeague(ctx.Context(), req.Name, req.UserID)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to create league"})
	}

	return ctx.Status(fiber.StatusCreated).JSON(league)
}
