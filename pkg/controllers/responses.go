package controllers

import "github.com/tobyrushton/fantasy-nba/pkg/db/models"

type leagueResponse struct {
	ID              int64  `json:"id"`
	Name            string `json:"name"`
	CreatorID       int64  `json:"creator_id"`
	CreatorUsername string `json:"creator_username"`
	MemberCount     int    `json:"member_count"`
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

type playerGameStatsResponse struct {
	PlayerID          int64  `json:"player_id"`
	GameDate          string `json:"game_date"`
	DidNotPlay        bool   `json:"did_not_play"`
	Points            int    `json:"points"`
	Rebounds          int    `json:"rebounds"`
	Assists           int    `json:"assists"`
	Steals            int    `json:"steals"`
	Blocks            int    `json:"blocks"`
	Turnovers         int    `json:"turnovers"`
	MadeThreePointers int    `json:"made_three_pointers"`
	MadeFreeThrows    int    `json:"made_free_throws"`
}

type playerStatsResponse struct {
	Player playerResponse            `json:"player"`
	Games  []playerGameStatsResponse `json:"games"`
}

func newLeagueResponse(league *models.League, creatorUsername string, memberCount int) leagueResponse {
	return leagueResponse{
		ID:              league.ID,
		Name:            league.Name,
		CreatorID:       league.CreatorID,
		CreatorUsername: creatorUsername,
		MemberCount:     memberCount,
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

func newPlayerGameStatsResponse(stats *models.PlayerGameStats) playerGameStatsResponse {
	gameDate := ""
	if stats.Game != nil {
		gameDate = stats.Game.GameDate.Format("2006-01-02")
	}

	return playerGameStatsResponse{
		PlayerID:          stats.PlayerID,
		GameDate:          gameDate,
		DidNotPlay:        stats.DidNotPlay,
		Points:            stats.Points,
		Rebounds:          stats.Rebounds,
		Assists:           stats.Assists,
		Steals:            stats.Steals,
		Blocks:            stats.Blocks,
		Turnovers:         stats.Turnovers,
		MadeThreePointers: stats.ThreePointersMade,
		MadeFreeThrows:    stats.FreeThrowsMade,
	}
}

func newPlayerStatsResponse(player models.Player, games []*models.PlayerGameStats) playerStatsResponse {
	gameStats := make([]playerGameStatsResponse, 0, len(games))
	for _, game := range games {
		gameStats = append(gameStats, newPlayerGameStatsResponse(game))
	}

	return playerStatsResponse{
		Player: newPlayerResponse(player),
		Games:  gameStats,
	}
}
