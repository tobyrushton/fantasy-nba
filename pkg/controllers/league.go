package controllers

import (
	"context"
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

func (c *LeagueController) buildLeagueResponse(ctx context.Context, league *models.League) (leagueResponse, error) {
	users, err := c.repo.GetUsersInLeague(ctx, int(league.ID))
	if err != nil {
		return leagueResponse{}, err
	}

	creatorUsername := "Unknown"
	if league.Creator != nil && league.Creator.Username != "" {
		creatorUsername = league.Creator.Username
	}

	creatorInLeague := false
	memberCount := len(users)
	for _, user := range users {
		if user == nil {
			continue
		}

		if user.ID == league.CreatorID {
			creatorInLeague = true
			if creatorUsername == "Unknown" && user.Username != "" {
				creatorUsername = user.Username
			}
		}
	}

	if !creatorInLeague {
		memberCount++
	}

	return newLeagueResponse(league, creatorUsername, memberCount), nil
}

// CreateLeague godoc
// @Summary Create league
// @Description Creates a league for the provided user.
// @Tags league
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body createLeagueRequest true "Create league request"
// @Success 201 {object} leagueResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /league/ [post]
func (c *LeagueController) CreateLeague(ctx fiber.Ctx) error {
	var req createLeagueRequest
	if err := ctx.Bind().Body(&req); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	league, err := c.repo.CreateLeague(ctx.Context(), req.Name, req.UserID)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to create league"})
	}

	if err := c.repo.JoinLeague(ctx.Context(), int(league.ID), req.UserID); err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to join league"})
	}

	leagueWithCreator, err := c.repo.GetLeagueByID(ctx.Context(), int(league.ID))
	if err == nil && leagueWithCreator != nil {
		league = leagueWithCreator
	}

	resp, err := c.buildLeagueResponse(ctx.Context(), league)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to get leagues"})
	}

	return ctx.Status(fiber.StatusCreated).JSON(resp)
}

// GetLeagues godoc
// @Summary List leagues
// @Description Returns all leagues.
// @Tags league
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {array} leagueResponse
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /league/ [get]
func (c *LeagueController) GetLeagues(ctx fiber.Ctx) error {
	leagues, err := c.repo.GetLeagues(ctx.Context())
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to get leagues"})
	}

	resp := make([]leagueResponse, 0, len(leagues))
	for _, league := range leagues {
		r, err := c.buildLeagueResponse(ctx.Context(), league)
		if err != nil {
			return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to get leagues"})
		}
		resp = append(resp, r)
	}

	return ctx.JSON(resp)
}

// GetLeagueByID godoc
// @Summary Get league by ID
// @Description Returns a single league by ID.
// @Tags league
// @Produce json
// @Security ApiKeyAuth
// @Param id path int true "League ID"
// @Success 200 {object} leagueResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /league/{id} [get]
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

	resp, err := c.buildLeagueResponse(ctx.Context(), league)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to get league"})
	}

	return ctx.JSON(resp)
}

// DeleteLeague godoc
// @Summary Delete league
// @Description Deletes a league created by the provided user.
// @Tags league
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body deleteLeagueRequest true "Delete league request"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /league/ [delete]
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

// JoinLeague godoc
// @Summary Join league
// @Description Adds a user to an existing league.
// @Tags league
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body joinLeagueRequest true "Join league request"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /league/join [post]
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

// CreateRoster godoc
// @Summary Create roster
// @Description Creates a roster of 10 unique players for a user in a league.
// @Tags league
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body createRosterRequest true "Create roster request"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /league/roster [post]
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

	users, err := c.repo.GetUsersInLeague(ctx.Context(), req.LeagueID)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to get users in league"})
	}

	isMember := false
	for _, user := range users {
		if user != nil && user.ID == req.UserID {
			isMember = true
			break
		}
	}

	if !isMember {
		return ctx.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "user is not a member of this league"})
	}

	// Create the roster
	if err := c.repo.CreateRoster(ctx.Context(), int64(req.LeagueID), req.UserID, req.PlayerIDs); err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to create roster"})
	}

	return ctx.JSON(fiber.Map{"message": "roster created successfully"})
}

// GetRostersByLeagueID godoc
// @Summary Get league rosters
// @Description Returns rosters grouped by user for the given league.
// @Tags league
// @Produce json
// @Security ApiKeyAuth
// @Param id path int true "League ID"
// @Success 200 {array} rosterResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /league/{id}/rosters [get]
func (c *LeagueController) GetRostersByLeagueID(ctx fiber.Ctx) error {
	leagueID, err := strconv.Atoi(ctx.Params("id"))
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid league ID"})
	}

	rosters, err := c.repo.GetRostersByLeagueID(ctx.Context(), leagueID)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to get rosters"})
	}

	groupByUser := make(map[int64][]playerResponse)
	for _, roster := range rosters {
		if roster.Player == nil {
			continue
		}

		groupByUser[roster.UserID] = append(groupByUser[roster.UserID], newPlayerResponse(*roster.Player))
	}

	users, err := c.repo.GetUsersInLeague(ctx.Context(), leagueID)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to get users in league"})
	}

	r := make([]rosterResponse, 0, len(users))
	for _, user := range users {
		if user == nil {
			continue
		}

		players := groupByUser[user.ID]
		r = append(r, rosterResponse{
			Players: players,
			User:    newUserResponse(*user),
		})
	}

	return ctx.JSON(r)
}

// UpdateRoster godoc
// @Summary Update roster
// @Description Swaps players in an existing league roster.
// @Tags league
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body updateRosterRequest true "Update roster request"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /league/roster [put]
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
