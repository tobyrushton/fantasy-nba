package controllers

import (
	"strconv"

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

type deleteLeagueRequest struct {
	ID     int   `json:"id"`
	UserID int64 `json:"user_id"`
}

type joinLeagueRequest struct {
	LeagueID int   `json:"league_id"`
	UserID   int64 `json:"user_id"`
}

type createRosterRequest struct {
	LeagueID  int     `json:"league_id"`
	UserID    int64   `json:"user_id"`
	PlayerIDs []int64 `json:"player_ids"`
}

type updateRosterRequest struct {
	LeagueID      int     `json:"league_id"`
	UserID        int64   `json:"user_id"`
	RemovePlayers []int64 `json:"remove_players"`
	AddPlayers    []int64 `json:"add_players"`
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

func (c *LeagueController) GetLeagues(ctx fiber.Ctx) error {
	leagues, err := c.repo.GetLeagues(ctx.Context())
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to get leagues"})
	}

	return ctx.JSON(leagues)
}

func (c *LeagueController) GetLeagueByID(ctx fiber.Ctx) error {
	id, err := strconv.Atoi(ctx.Params("id"))
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid league ID"})
	}

	league, err := c.repo.GetLeagueByID(ctx.Context(), id)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to get league"})
	}

	if league == nil {
		return ctx.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "league not found"})
	}

	return ctx.JSON(league)
}

func (c *LeagueController) DeleteLeague(ctx fiber.Ctx) error {
	var req deleteLeagueRequest
	if err := ctx.Bind().Body(&req); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	// Delete the league
	if err := c.repo.DeleteLeague(ctx.Context(), req.ID, req.UserID); err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to delete league"})
	}

	return ctx.JSON(fiber.Map{"message": "league deleted successfully"})
}

func (c *LeagueController) JoinLeague(ctx fiber.Ctx) error {
	var req joinLeagueRequest
	if err := ctx.Bind().Body(&req); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	// Join the league
	if err := c.repo.JoinLeague(ctx.Context(), req.LeagueID, req.UserID); err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to join league"})
	}

	return ctx.JSON(fiber.Map{"message": "successfully joined league"})
}

func (c *LeagueController) CreateRoster(ctx fiber.Ctx) error {
	var req createRosterRequest
	if err := ctx.Bind().Body(&req); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	if len(req.PlayerIDs) != 10 {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "roster must contain exactly 10 players"})
	}
	contains := make(map[int64]interface{})
	for _, id := range req.PlayerIDs {
		if _, exists := contains[id]; exists {
			return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "duplicate player IDs are not allowed"})
		}
		contains[id] = struct{}{}
	}

	// Create the roster
	if err := c.repo.CreateRoster(ctx.Context(), int64(req.LeagueID), req.UserID, req.PlayerIDs); err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to create roster"})
	}

	return ctx.JSON(fiber.Map{"message": "roster created successfully"})
}

func (c *LeagueController) GetRostersByLeagueID(ctx fiber.Ctx) error {
	leagueID, err := strconv.Atoi(ctx.Params("id"))
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid league ID"})
	}

	rosters, err := c.repo.GetRostersByLeagueID(ctx.Context(), leagueID)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to get rosters"})
	}

	groupByUser := make(map[int64][]models.Player)
	for _, roster := range rosters {
		groupByUser[roster.UserID] = append(groupByUser[roster.UserID], *roster.Player)
	}

	return ctx.JSON(groupByUser)
}

func (c *LeagueController) UpdateRoster(ctx fiber.Ctx) error {
	var req updateRosterRequest
	if err := ctx.Bind().Body(&req); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	if len(req.AddPlayers)+len(req.RemovePlayers) == 0 {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "must add or remove at least one player"})
	}

	if len(req.AddPlayers) != len(req.RemovePlayers) {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "number of players to add must equal number of players to remove"})
	}

	// Update the roster
	if err := c.repo.UpdateRoster(ctx.Context(), int64(req.LeagueID), req.UserID, req.RemovePlayers, req.AddPlayers); err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to update roster"})
	}

	return ctx.JSON(fiber.Map{"message": "roster updated successfully"})
}
