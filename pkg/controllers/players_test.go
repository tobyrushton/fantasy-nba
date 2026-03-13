package controllers

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/suite"
	"github.com/tobyrushton/fantasy-nba/pkg/db/models"
	"github.com/tobyrushton/fantasy-nba/pkg/fakes"
)

type PlayersControllerSuite struct {
	suite.Suite
}

func TestPlayersControllerSuite(t *testing.T) {
	suite.Run(t, new(PlayersControllerSuite))
}

func (s *PlayersControllerSuite) TestGetPlayersReturnsPlayers() {
	repo := &fakes.FakeRepo{}
	repo.GetPlayersReturns([]models.Player{
		{ID: 1, FirstName: "Nikola", LastName: "Jokic", Position: "C"},
		{ID: 2, FirstName: "Stephen", LastName: "Curry", Position: "PG"},
	}, nil)

	res := s.performRequest(s.newApp(repo), http.MethodGet, "/players")

	s.Equal(fiber.StatusOK, res.StatusCode)
	s.Equal(1, repo.GetPlayersCallCount())

	_, gotSearch, gotPosition := repo.GetPlayersArgsForCall(0)
	s.Equal("", gotSearch)
	s.Equal("", gotPosition)

	players := s.decodePlayersResponse(res)
	s.Len(players, 2)
	s.Equal(int64(1), players[0].ID)
	s.Equal("Nikola", players[0].FirstName)
	s.Equal("Jokic", players[0].LastName)
	s.Equal("C", players[0].Position)
	s.Equal(int64(2), players[1].ID)
	s.Equal("Stephen", players[1].FirstName)
	s.Equal("Curry", players[1].LastName)
	s.Equal("PG", players[1].Position)
}

func (s *PlayersControllerSuite) TestGetPlayersPassesSearchAndPositionQueryToRepo() {
	repo := &fakes.FakeRepo{}
	repo.GetPlayersReturns([]models.Player{}, nil)

	res := s.performRequest(s.newApp(repo), http.MethodGet, "/players?search=steph&position=PG")

	s.Equal(fiber.StatusOK, res.StatusCode)
	s.Equal(1, repo.GetPlayersCallCount())

	_, gotSearch, gotPosition := repo.GetPlayersArgsForCall(0)
	s.Equal("steph", gotSearch)
	s.Equal("PG", gotPosition)

	players := s.decodePlayersResponse(res)
	s.Len(players, 0)
}

func (s *PlayersControllerSuite) TestGetPlayersReturns500WhenRepoFails() {
	repo := &fakes.FakeRepo{}
	repo.GetPlayersReturns(nil, errors.New("db down"))

	res := s.performRequest(s.newApp(repo), http.MethodGet, "/players?search=steph&position=PG")

	s.Equal(fiber.StatusInternalServerError, res.StatusCode)
	s.Equal(1, repo.GetPlayersCallCount())

	_, gotSearch, gotPosition := repo.GetPlayersArgsForCall(0)
	s.Equal("steph", gotSearch)
	s.Equal("PG", gotPosition)

	s.Equal(map[string]string{"error": "Failed to fetch players"}, s.decodeErrorResponse(res))
}

func (s *PlayersControllerSuite) TestGetPlayerReturns400WhenIDIsInvalid() {
	repo := &fakes.FakeRepo{}

	res := s.performRequest(s.newApp(repo), http.MethodGet, "/players/not-a-number")

	s.Equal(fiber.StatusBadRequest, res.StatusCode)
	s.Equal(0, repo.GetPlayerByIDCallCount())
	s.Equal(map[string]string{"error": "Invalid player ID"}, s.decodeErrorResponse(res))
}

func (s *PlayersControllerSuite) TestGetPlayerReturns500WhenRepoFails() {
	repo := &fakes.FakeRepo{}
	repo.GetPlayerByIDReturns(nil, errors.New("db down"))

	res := s.performRequest(s.newApp(repo), http.MethodGet, "/players/1")

	s.Equal(fiber.StatusInternalServerError, res.StatusCode)
	s.Equal(1, repo.GetPlayerByIDCallCount())
	s.Equal(0, repo.GetPlayerStatsByIDCallCount())

	_, gotID := repo.GetPlayerByIDArgsForCall(0)
	s.Equal(int64(1), gotID)
	s.Equal(map[string]string{"error": "Failed to fetch player"}, s.decodeErrorResponse(res))
}

func (s *PlayersControllerSuite) TestGetPlayerReturns404WhenPlayerNotFound() {
	repo := &fakes.FakeRepo{}
	repo.GetPlayerByIDReturns(nil, nil)

	res := s.performRequest(s.newApp(repo), http.MethodGet, "/players/1")

	s.Equal(fiber.StatusNotFound, res.StatusCode)
	s.Equal(1, repo.GetPlayerByIDCallCount())
	s.Equal(0, repo.GetPlayerStatsByIDCallCount())
	s.Equal(map[string]string{"error": "Player not found"}, s.decodeErrorResponse(res))
}

func (s *PlayersControllerSuite) TestGetPlayerReturnsPlayerWithStats() {
	repo := &fakes.FakeRepo{}
	repo.GetPlayerByIDReturns(&models.Player{
		ID: 1, FirstName: "Nikola", LastName: "Jokic", Position: "C",
	}, nil)
	repo.GetPlayerStatsByIDReturns([]*models.PlayerGameStats{
		{PlayerID: 1, Points: 30, Rebounds: 12, Assists: 8, Steals: 1, Blocks: 2, Turnovers: 3, ThreePointersMade: 1, FreeThrowsMade: 5},
		{PlayerID: 1, Points: 25, Rebounds: 10, Assists: 6, Steals: 2, Blocks: 1, Turnovers: 2, ThreePointersMade: 2, FreeThrowsMade: 4},
	}, nil)

	res := s.performRequest(s.newApp(repo), http.MethodGet, "/players/1")

	s.Equal(fiber.StatusOK, res.StatusCode)
	s.Equal(1, repo.GetPlayerByIDCallCount())
	s.Equal(1, repo.GetPlayerStatsByIDCallCount())

	_, gotID := repo.GetPlayerByIDArgsForCall(0)
	s.Equal(int64(1), gotID)
	_, gotStatsID := repo.GetPlayerStatsByIDArgsForCall(0)
	s.Equal(int64(1), gotStatsID)

	var resp playerStatsResponse
	s.Require().NoError(json.NewDecoder(res.Body).Decode(&resp))
	s.NoError(res.Body.Close())

	s.Equal(int64(1), resp.Player.ID)
	s.Equal("Nikola", resp.Player.FirstName)
	s.Equal("Jokic", resp.Player.LastName)
	s.Equal("C", resp.Player.Position)
	s.Len(resp.Games, 2)
	s.Equal(int64(1), resp.Games[0].PlayerID)
	s.Equal(30, resp.Games[0].Points)
	s.Equal(12, resp.Games[0].Rebounds)
	s.Equal(8, resp.Games[0].Assists)
	s.Equal(1, resp.Games[0].Steals)
	s.Equal(2, resp.Games[0].Blocks)
	s.Equal(3, resp.Games[0].Turnovers)
	s.Equal(1, resp.Games[0].MadeThreePointers)
	s.Equal(5, resp.Games[0].MadeFreeThrows)
}

func (s *PlayersControllerSuite) TestGetPlayerReturnsPlayerWithNoStats() {
	repo := &fakes.FakeRepo{}
	repo.GetPlayerByIDReturns(&models.Player{
		ID: 1, FirstName: "Nikola", LastName: "Jokic", Position: "C",
	}, nil)
	repo.GetPlayerStatsByIDReturns(nil, nil)

	res := s.performRequest(s.newApp(repo), http.MethodGet, "/players/1")

	s.Equal(fiber.StatusOK, res.StatusCode)

	var resp playerStatsResponse
	s.Require().NoError(json.NewDecoder(res.Body).Decode(&resp))
	s.NoError(res.Body.Close())

	s.Equal(int64(1), resp.Player.ID)
	s.Empty(resp.Games)
}

func (s *PlayersControllerSuite) newApp(repo *fakes.FakeRepo) *fiber.App {
	controller := NewPlayersController(repo)
	app := fiber.New()
	app.Get("/players", controller.GetPlayers)
	app.Get("/players/:id", controller.GetPlayer)
	return app
}

func (s *PlayersControllerSuite) performRequest(app *fiber.App, method, path string) *http.Response {
	req := httptest.NewRequest(method, path, nil)

	res, err := app.Test(req)
	s.Require().NoError(err)
	return res
}

func (s *PlayersControllerSuite) decodePlayersResponse(res *http.Response) []playerResponse {
	s.T().Cleanup(func() {
		s.NoError(res.Body.Close())
	})

	var players []playerResponse
	err := json.NewDecoder(res.Body).Decode(&players)
	s.Require().NoError(err)
	return players
}

func (s *PlayersControllerSuite) decodeErrorResponse(res *http.Response) map[string]string {
	s.T().Cleanup(func() {
		s.NoError(res.Body.Close())
	})

	var resp map[string]string
	err := json.NewDecoder(res.Body).Decode(&resp)
	s.Require().NoError(err)
	return resp
}
