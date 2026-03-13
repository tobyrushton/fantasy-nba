package controllers

import (
	"strconv"

	"github.com/gofiber/fiber/v3"
	"github.com/tobyrushton/fantasy-nba/pkg/db/models"
)

type PlayersController struct {
	repo models.Repo
}

func NewPlayersController(repo models.Repo) *PlayersController {
	return &PlayersController{repo: repo}
}

// GetPlayers godoc
// @Summary List players
// @Description Returns players, optionally filtered by name and position.
// @Tags players
// @Produce json
// @Param search query string false "Search by player name"
// @Param position query string false "Filter by position"
// @Success 200 {array} playerResponse
// @Failure 500 {object} map[string]string
// @Router /players/ [get]
func (c *PlayersController) GetPlayers(ctx fiber.Ctx) error {
	search := ctx.Query("search")
	position := ctx.Query("position")

	players, err := c.repo.GetPlayers(ctx.Context(), search, position)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch players"})
	}

	return ctx.JSON(newPlayerResponses(players))
}

// GetPlayer godoc
// @Summary Get player by ID
// @Description Returns a player and their game stats.
// @Tags players
// @Produce json
// @Param id path int true "Player ID"
// @Success 200 {object} playerStatsResponse
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /players/{id} [get]
func (c *PlayersController) GetPlayer(ctx fiber.Ctx) error {
	idParam := ctx.Params("id")
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid player ID"})
	}

	player, err := c.repo.GetPlayerByID(ctx.Context(), id)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch player"})
	}
	if player == nil {
		return ctx.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Player not found"})
	}

	playerStats, err := c.repo.GetPlayerStatsByID(ctx.Context(), id)

	return ctx.JSON(newPlayerStatsResponse(*player, playerStats))
}
