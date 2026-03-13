package controllers

import "github.com/tobyrushton/fantasy-nba/pkg/db/models"

type leagueResponse struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	CreatorID int64  `json:"creator_id"`
}

type playerResponse struct {
	ID        int64  `json:"id"`
	NBAID     string `json:"nba_id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Position  string `json:"position"`
	TeamID    int64  `json:"team_id"`
}

type rosterResponse struct {
	Players []playerResponse `json:"players"`
	User    userResponse     `json:"user"`
}

type userResponse struct {
	ID       int64  `json:"id"`
	Username string `json:"username"`
}

func newLeagueResponse(league *models.League) leagueResponse {
	return leagueResponse{
		ID:        league.ID,
		Name:      league.Name,
		CreatorID: league.CreatorID,
	}
}

func newPlayerResponse(player models.Player) playerResponse {
	return playerResponse{
		ID:        player.ID,
		NBAID:     player.NBAID,
		FirstName: player.FirstName,
		LastName:  player.LastName,
		Position:  player.Position,
		TeamID:    player.TeamID,
	}
}

func newLeagueResponses(leagues []*models.League) []leagueResponse {
	resp := make([]leagueResponse, 0, len(leagues))
	for _, league := range leagues {
		resp = append(resp, newLeagueResponse(league))
	}

	return resp
}

func newPlayerResponses(players []models.Player) []playerResponse {
	resp := make([]playerResponse, 0, len(players))
	for _, player := range players {
		resp = append(resp, newPlayerResponse(player))
	}

	return resp
}

func newUserResponse(user models.User) userResponse {
	return userResponse{
		ID:       user.ID,
		Username: user.Username,
	}
}
