package controllers

import (
	"github.com/gofiber/fiber/v3"
	"github.com/tobyrushton/fantasy-nba/pkg/db/models"
)

type PlayersController struct {
	repo models.Repo
}

func NewPlayersController(repo models.Repo) *PlayersController {
	return &PlayersController{repo: repo}
}

func (c *PlayersController) GetPlayers(ctx fiber.Ctx) error {
	search := ctx.Query("search")
	position := ctx.Query("position")

	players, err := c.repo.GetPlayers(ctx.Context(), search, position)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch players"})
	}

	return ctx.JSON(players)
}
