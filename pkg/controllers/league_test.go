package controllers

import (
	"bytes"
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

type LeagueControllerSuite struct {
	suite.Suite
}

func TestLeagueControllerSuite(t *testing.T) {
	suite.Run(t, new(LeagueControllerSuite))
}

func (s *LeagueControllerSuite) TestCreateLeagueReturns201AndLeague() {
	league := &models.League{
		ID:        7,
		Name:      "Champions",
		CreatorID: 42,
	}
	repo := &fakes.FakeRepo{}
	repo.CreateLeagueReturns(league, nil)

	res := s.performJSONRequest(s.newApp(repo), http.MethodPost, "/leagues", `{"name":"Champions","user_id":42}`)

	s.Equal(fiber.StatusCreated, res.StatusCode)
	s.Equal(1, repo.CreateLeagueCallCount())

	_, gotName, gotUserID := repo.CreateLeagueArgsForCall(0)
	s.Equal("Champions", gotName)
	s.Equal(int64(42), gotUserID)

	var resp leagueResponse
	s.Require().NoError(json.NewDecoder(res.Body).Decode(&resp))
	s.Equal(int64(7), resp.ID)
	s.Equal("Champions", resp.Name)
	s.Equal(int64(42), resp.CreatorID)
	s.NoError(res.Body.Close())
}

func (s *LeagueControllerSuite) TestCreateLeagueReturns400WhenBodyIsInvalid() {
	repo := &fakes.FakeRepo{}

	res := s.performJSONRequest(s.newApp(repo), http.MethodPost, "/leagues", "{bad-json")

	s.Equal(fiber.StatusBadRequest, res.StatusCode)
	s.Equal(0, repo.CreateLeagueCallCount())
	s.Equal(map[string]string{"error": "invalid request body"}, s.decodeErrorResponse(res))
}

func (s *LeagueControllerSuite) TestCreateLeagueReturns500WhenRepoFails() {
	repo := &fakes.FakeRepo{}
	repo.CreateLeagueReturns(nil, errors.New("db down"))

	res := s.performJSONRequest(s.newApp(repo), http.MethodPost, "/leagues", `{"name":"Champions","user_id":42}`)

	s.Equal(fiber.StatusInternalServerError, res.StatusCode)
	s.Equal(1, repo.CreateLeagueCallCount())
	s.Equal(map[string]string{"error": "failed to create league"}, s.decodeErrorResponse(res))
}

func (s *LeagueControllerSuite) TestGetLeaguesReturnsLeagueList() {
	repo := &fakes.FakeRepo{}
	repo.GetLeaguesReturns([]*models.League{
		{ID: 1, Name: "Champions", CreatorID: 42},
		{ID: 2, Name: "Dynasty", CreatorID: 84},
	}, nil)

	res := s.performJSONRequest(s.newApp(repo), http.MethodGet, "/leagues", "")

	s.Equal(fiber.StatusOK, res.StatusCode)
	s.Equal(1, repo.GetLeaguesCallCount())

	var resp []leagueResponse
	s.Require().NoError(json.NewDecoder(res.Body).Decode(&resp))
	s.Len(resp, 2)
	s.Equal(int64(1), resp[0].ID)
	s.Equal("Champions", resp[0].Name)
	s.Equal(int64(42), resp[0].CreatorID)
	s.Equal(int64(2), resp[1].ID)
	s.Equal("Dynasty", resp[1].Name)
	s.Equal(int64(84), resp[1].CreatorID)
	s.NoError(res.Body.Close())
}

func (s *LeagueControllerSuite) TestGetLeaguesReturns500WhenRepoFails() {
	repo := &fakes.FakeRepo{}
	repo.GetLeaguesReturns(nil, errors.New("db down"))

	res := s.performJSONRequest(s.newApp(repo), http.MethodGet, "/leagues", "")

	s.Equal(fiber.StatusInternalServerError, res.StatusCode)
	s.Equal(1, repo.GetLeaguesCallCount())
	s.Equal(map[string]string{"error": "failed to get leagues"}, s.decodeErrorResponse(res))
}

func (s *LeagueControllerSuite) TestGetLeagueByIDReturns400WhenIDIsInvalid() {
	repo := &fakes.FakeRepo{}

	res := s.performJSONRequest(s.newApp(repo), http.MethodGet, "/leagues/not-a-number", "")

	s.Equal(fiber.StatusBadRequest, res.StatusCode)
	s.Equal(0, repo.GetLeagueByIDCallCount())
	s.Equal(map[string]string{"error": "invalid league ID"}, s.decodeErrorResponse(res))
}

func (s *LeagueControllerSuite) TestGetLeagueByIDReturns500WhenRepoFails() {
	repo := &fakes.FakeRepo{}
	repo.GetLeagueByIDReturns(nil, errors.New("db down"))

	res := s.performJSONRequest(s.newApp(repo), http.MethodGet, "/leagues/7", "")

	s.Equal(fiber.StatusInternalServerError, res.StatusCode)
	s.Equal(1, repo.GetLeagueByIDCallCount())

	_, gotID := repo.GetLeagueByIDArgsForCall(0)
	s.Equal(7, gotID)
	s.Equal(map[string]string{"error": "failed to get league"}, s.decodeErrorResponse(res))
}

func (s *LeagueControllerSuite) TestGetLeagueByIDReturns404WhenLeagueNotFound() {
	repo := &fakes.FakeRepo{}
	repo.GetLeagueByIDReturns(nil, nil)

	res := s.performJSONRequest(s.newApp(repo), http.MethodGet, "/leagues/7", "")

	s.Equal(fiber.StatusNotFound, res.StatusCode)
	s.Equal(1, repo.GetLeagueByIDCallCount())
	s.Equal(map[string]string{"error": "league not found"}, s.decodeErrorResponse(res))
}

func (s *LeagueControllerSuite) TestGetLeagueByIDReturnsLeague() {
	repo := &fakes.FakeRepo{}
	repo.GetLeagueByIDReturns(&models.League{ID: 7, Name: "Champions", CreatorID: 42}, nil)

	res := s.performJSONRequest(s.newApp(repo), http.MethodGet, "/leagues/7", "")

	s.Equal(fiber.StatusOK, res.StatusCode)
	s.Equal(1, repo.GetLeagueByIDCallCount())

	_, gotID := repo.GetLeagueByIDArgsForCall(0)
	s.Equal(7, gotID)

	var resp leagueResponse
	s.Require().NoError(json.NewDecoder(res.Body).Decode(&resp))
	s.Equal(int64(7), resp.ID)
	s.Equal("Champions", resp.Name)
	s.Equal(int64(42), resp.CreatorID)
	s.NoError(res.Body.Close())
}

func (s *LeagueControllerSuite) TestDeleteLeagueReturns400WhenBodyIsInvalid() {
	repo := &fakes.FakeRepo{}

	res := s.performJSONRequest(s.newApp(repo), http.MethodDelete, "/leagues", "{bad-json")

	s.Equal(fiber.StatusBadRequest, res.StatusCode)
	s.Equal(0, repo.DeleteLeagueCallCount())
	s.Equal(map[string]string{"error": "invalid request body"}, s.decodeErrorResponse(res))
}

func (s *LeagueControllerSuite) TestDeleteLeagueReturns500WhenRepoFails() {
	repo := &fakes.FakeRepo{}
	repo.DeleteLeagueReturns(errors.New("db down"))

	res := s.performJSONRequest(s.newApp(repo), http.MethodDelete, "/leagues", `{"id":7,"user_id":42}`)

	s.Equal(fiber.StatusInternalServerError, res.StatusCode)
	s.Equal(1, repo.DeleteLeagueCallCount())

	_, gotID, gotUserID := repo.DeleteLeagueArgsForCall(0)
	s.Equal(7, gotID)
	s.Equal(int64(42), gotUserID)
	s.Equal(map[string]string{"error": "failed to delete league"}, s.decodeErrorResponse(res))
}

func (s *LeagueControllerSuite) TestDeleteLeagueReturnsSuccessMessage() {
	repo := &fakes.FakeRepo{}
	repo.DeleteLeagueReturns(nil)

	res := s.performJSONRequest(s.newApp(repo), http.MethodDelete, "/leagues", `{"id":7,"user_id":42}`)

	s.Equal(fiber.StatusOK, res.StatusCode)
	s.Equal(1, repo.DeleteLeagueCallCount())

	_, gotID, gotUserID := repo.DeleteLeagueArgsForCall(0)
	s.Equal(7, gotID)
	s.Equal(int64(42), gotUserID)

	var resp map[string]string
	s.Require().NoError(json.NewDecoder(res.Body).Decode(&resp))
	s.Equal("league deleted successfully", resp["message"])
	s.NoError(res.Body.Close())
}

func (s *LeagueControllerSuite) TestJoinLeagueReturns400WhenBodyIsInvalid() {
	repo := &fakes.FakeRepo{}

	res := s.performJSONRequest(s.newApp(repo), http.MethodPost, "/leagues/join", "{bad-json")

	s.Equal(fiber.StatusBadRequest, res.StatusCode)
	s.Equal(0, repo.JoinLeagueCallCount())
	s.Equal(map[string]string{"error": "invalid request body"}, s.decodeErrorResponse(res))
}

func (s *LeagueControllerSuite) TestJoinLeagueReturns500WhenRepoFails() {
	repo := &fakes.FakeRepo{}
	repo.JoinLeagueReturns(errors.New("db down"))

	res := s.performJSONRequest(s.newApp(repo), http.MethodPost, "/leagues/join", `{"league_id":7,"user_id":42}`)

	s.Equal(fiber.StatusInternalServerError, res.StatusCode)
	s.Equal(1, repo.JoinLeagueCallCount())

	_, gotLeagueID, gotUserID := repo.JoinLeagueArgsForCall(0)
	s.Equal(7, gotLeagueID)
	s.Equal(int64(42), gotUserID)
	s.Equal(map[string]string{"error": "failed to join league"}, s.decodeErrorResponse(res))
}

func (s *LeagueControllerSuite) TestJoinLeagueReturnsSuccessMessage() {
	repo := &fakes.FakeRepo{}
	repo.JoinLeagueReturns(nil)

	res := s.performJSONRequest(s.newApp(repo), http.MethodPost, "/leagues/join", `{"league_id":7,"user_id":42}`)

	s.Equal(fiber.StatusOK, res.StatusCode)
	s.Equal(1, repo.JoinLeagueCallCount())

	_, gotLeagueID, gotUserID := repo.JoinLeagueArgsForCall(0)
	s.Equal(7, gotLeagueID)
	s.Equal(int64(42), gotUserID)

	var resp map[string]string
	s.Require().NoError(json.NewDecoder(res.Body).Decode(&resp))
	s.Equal("successfully joined league", resp["message"])
	s.NoError(res.Body.Close())
}

func (s *LeagueControllerSuite) TestCreateRosterReturns400WhenBodyIsInvalid() {
	repo := &fakes.FakeRepo{}

	res := s.performJSONRequest(s.newApp(repo), http.MethodPost, "/rosters", "{bad-json")

	s.Equal(fiber.StatusBadRequest, res.StatusCode)
	s.Equal(0, repo.CreateRosterCallCount())
	s.Equal(map[string]string{"error": "invalid request body"}, s.decodeErrorResponse(res))
}

func (s *LeagueControllerSuite) TestCreateRosterReturns400WhenPlayerCountIsNotTen() {
	repo := &fakes.FakeRepo{}

	res := s.performJSONRequest(s.newApp(repo), http.MethodPost, "/rosters", `{"league_id":7,"user_id":42,"player_ids":[1,2,3]}`)

	s.Equal(fiber.StatusBadRequest, res.StatusCode)
	s.Equal(0, repo.CreateRosterCallCount())
	s.Equal(map[string]string{"error": "roster must contain exactly 10 players"}, s.decodeErrorResponse(res))
}

func (s *LeagueControllerSuite) TestCreateRosterReturns400WhenPlayerIDsAreDuplicated() {
	repo := &fakes.FakeRepo{}

	res := s.performJSONRequest(s.newApp(repo), http.MethodPost, "/rosters", `{"league_id":7,"user_id":42,"player_ids":[1,2,3,4,5,6,7,8,9,1]}`)

	s.Equal(fiber.StatusBadRequest, res.StatusCode)
	s.Equal(0, repo.CreateRosterCallCount())
	s.Equal(map[string]string{"error": "duplicate player IDs are not allowed"}, s.decodeErrorResponse(res))
}

func (s *LeagueControllerSuite) TestCreateRosterReturns500WhenRepoFails() {
	repo := &fakes.FakeRepo{}
	repo.CreateRosterReturns(errors.New("db down"))
	playerIDs := []int64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}

	res := s.performJSONRequest(s.newApp(repo), http.MethodPost, "/rosters", `{"league_id":7,"user_id":42,"player_ids":[1,2,3,4,5,6,7,8,9,10]}`)

	s.Equal(fiber.StatusInternalServerError, res.StatusCode)
	s.Equal(1, repo.CreateRosterCallCount())

	_, gotLeagueID, gotUserID, gotPlayerIDs := repo.CreateRosterArgsForCall(0)
	s.Equal(int64(7), gotLeagueID)
	s.Equal(int64(42), gotUserID)
	s.Equal(playerIDs, gotPlayerIDs)
	s.Equal(map[string]string{"error": "failed to create roster"}, s.decodeErrorResponse(res))
}

func (s *LeagueControllerSuite) TestCreateRosterReturnsSuccessMessage() {
	repo := &fakes.FakeRepo{}
	repo.CreateRosterReturns(nil)
	playerIDs := []int64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}

	res := s.performJSONRequest(s.newApp(repo), http.MethodPost, "/rosters", `{"league_id":7,"user_id":42,"player_ids":[1,2,3,4,5,6,7,8,9,10]}`)

	s.Equal(fiber.StatusOK, res.StatusCode)
	s.Equal(1, repo.CreateRosterCallCount())

	_, gotLeagueID, gotUserID, gotPlayerIDs := repo.CreateRosterArgsForCall(0)
	s.Equal(int64(7), gotLeagueID)
	s.Equal(int64(42), gotUserID)
	s.Equal(playerIDs, gotPlayerIDs)

	var resp map[string]string
	s.Require().NoError(json.NewDecoder(res.Body).Decode(&resp))
	s.Equal("roster created successfully", resp["message"])
	s.NoError(res.Body.Close())
}

func (s *LeagueControllerSuite) TestGetRostersByLeagueIDReturns400WhenIDIsInvalid() {
	repo := &fakes.FakeRepo{}

	res := s.performJSONRequest(s.newApp(repo), http.MethodGet, "/leagues/not-a-number/rosters", "")

	s.Equal(fiber.StatusBadRequest, res.StatusCode)
	s.Equal(0, repo.GetRostersByLeagueIDCallCount())
	s.Equal(map[string]string{"error": "invalid league ID"}, s.decodeErrorResponse(res))
}

func (s *LeagueControllerSuite) TestGetRostersByLeagueIDReturns500WhenRepoFails() {
	repo := &fakes.FakeRepo{}
	repo.GetRostersByLeagueIDReturns(nil, errors.New("db down"))

	res := s.performJSONRequest(s.newApp(repo), http.MethodGet, "/leagues/7/rosters", "")

	s.Equal(fiber.StatusInternalServerError, res.StatusCode)
	s.Equal(1, repo.GetRostersByLeagueIDCallCount())
	s.Equal(0, repo.GetUsersInLeagueCallCount())

	_, gotLeagueID := repo.GetRostersByLeagueIDArgsForCall(0)
	s.Equal(7, gotLeagueID)
	s.Equal(map[string]string{"error": "failed to get rosters"}, s.decodeErrorResponse(res))
}

func (s *LeagueControllerSuite) TestGetRostersByLeagueIDReturnsPlayersGroupedByUser() {
	repo := &fakes.FakeRepo{}
	repo.GetRostersByLeagueIDReturns([]*models.TeamRoster{
		{UserID: 42, Player: &models.Player{ID: 1, FirstName: "Stephen", LastName: "Curry"}},
		{UserID: 42, Player: &models.Player{ID: 2, FirstName: "Klay", LastName: "Thompson"}},
		{UserID: 84, Player: &models.Player{ID: 3, FirstName: "Nikola", LastName: "Jokic"}},
	}, nil)
	repo.GetUsersInLeagueReturns([]*models.User{
		{ID: 42, Username: "toby"},
		{ID: 84, Username: "alex"},
	}, nil)

	res := s.performJSONRequest(s.newApp(repo), http.MethodGet, "/leagues/7/rosters", "")

	s.Equal(fiber.StatusOK, res.StatusCode)
	s.Equal(1, repo.GetRostersByLeagueIDCallCount())
	s.Equal(1, repo.GetUsersInLeagueCallCount())

	_, gotLeagueID := repo.GetRostersByLeagueIDArgsForCall(0)
	s.Equal(7, gotLeagueID)

	_, gotUsersLeagueID := repo.GetUsersInLeagueArgsForCall(0)
	s.Equal(7, gotUsersLeagueID)

	var resp []rosterResponse
	s.Require().NoError(json.NewDecoder(res.Body).Decode(&resp))
	s.Len(resp, 2)
	s.Equal(int64(42), resp[0].User.ID)
	s.Equal("toby", resp[0].User.Username)
	s.Len(resp[0].Players, 2)
	s.Equal("Stephen", resp[0].Players[0].FirstName)
	s.Equal("Klay", resp[0].Players[1].FirstName)
	s.Equal(int64(84), resp[1].User.ID)
	s.Equal("alex", resp[1].User.Username)
	s.Len(resp[1].Players, 1)
	s.Equal("Nikola", resp[1].Players[0].FirstName)
	s.NoError(res.Body.Close())
}

func (s *LeagueControllerSuite) TestUpdateRosterReturns400WhenBodyIsInvalid() {
	repo := &fakes.FakeRepo{}

	res := s.performJSONRequest(s.newApp(repo), http.MethodPut, "/rosters", "{bad-json")

	s.Equal(fiber.StatusBadRequest, res.StatusCode)
	s.Equal(0, repo.UpdateRosterCallCount())
	s.Equal(map[string]string{"error": "invalid request body"}, s.decodeErrorResponse(res))
}

func (s *LeagueControllerSuite) TestUpdateRosterReturns400WhenNoPlayersProvided() {
	repo := &fakes.FakeRepo{}

	res := s.performJSONRequest(s.newApp(repo), http.MethodPut, "/rosters", `{"league_id":7,"user_id":42,"remove_players":[],"add_players":[]}`)

	s.Equal(fiber.StatusBadRequest, res.StatusCode)
	s.Equal(0, repo.UpdateRosterCallCount())
	s.Equal(map[string]string{"error": "must add or remove at least one player"}, s.decodeErrorResponse(res))
}

func (s *LeagueControllerSuite) TestUpdateRosterReturns400WhenAddRemoveCountsDiffer() {
	repo := &fakes.FakeRepo{}

	res := s.performJSONRequest(s.newApp(repo), http.MethodPut, "/rosters", `{"league_id":7,"user_id":42,"remove_players":[1],"add_players":[2,3]}`)

	s.Equal(fiber.StatusBadRequest, res.StatusCode)
	s.Equal(0, repo.UpdateRosterCallCount())
	s.Equal(map[string]string{"error": "number of players to add must equal number of players to remove"}, s.decodeErrorResponse(res))
}

func (s *LeagueControllerSuite) TestUpdateRosterReturns500WhenRepoFails() {
	repo := &fakes.FakeRepo{}
	repo.UpdateRosterReturns(errors.New("db down"))
	removePlayers := []int64{1, 2}
	addPlayers := []int64{3, 4}

	res := s.performJSONRequest(s.newApp(repo), http.MethodPut, "/rosters", `{"league_id":7,"user_id":42,"remove_players":[1,2],"add_players":[3,4]}`)

	s.Equal(fiber.StatusInternalServerError, res.StatusCode)
	s.Equal(1, repo.UpdateRosterCallCount())

	_, gotLeagueID, gotUserID, gotRemovePlayers, gotAddPlayers := repo.UpdateRosterArgsForCall(0)
	s.Equal(int64(7), gotLeagueID)
	s.Equal(int64(42), gotUserID)
	s.Equal(removePlayers, gotRemovePlayers)
	s.Equal(addPlayers, gotAddPlayers)
	s.Equal(map[string]string{"error": "failed to update roster"}, s.decodeErrorResponse(res))
}

func (s *LeagueControllerSuite) TestUpdateRosterReturnsSuccessMessage() {
	repo := &fakes.FakeRepo{}
	repo.UpdateRosterReturns(nil)
	removePlayers := []int64{1, 2}
	addPlayers := []int64{3, 4}

	res := s.performJSONRequest(s.newApp(repo), http.MethodPut, "/rosters", `{"league_id":7,"user_id":42,"remove_players":[1,2],"add_players":[3,4]}`)

	s.Equal(fiber.StatusOK, res.StatusCode)
	s.Equal(1, repo.UpdateRosterCallCount())

	_, gotLeagueID, gotUserID, gotRemovePlayers, gotAddPlayers := repo.UpdateRosterArgsForCall(0)
	s.Equal(int64(7), gotLeagueID)
	s.Equal(int64(42), gotUserID)
	s.Equal(removePlayers, gotRemovePlayers)
	s.Equal(addPlayers, gotAddPlayers)

	var resp map[string]string
	s.Require().NoError(json.NewDecoder(res.Body).Decode(&resp))
	s.Equal("roster updated successfully", resp["message"])
	s.NoError(res.Body.Close())
}

func (s *LeagueControllerSuite) newApp(repo *fakes.FakeRepo) *fiber.App {
	controller := NewLeagueController(repo)
	app := fiber.New()
	app.Post("/leagues", controller.CreateLeague)
	app.Post("/leagues/join", controller.JoinLeague)
	app.Post("/rosters", controller.CreateRoster)
	app.Put("/rosters", controller.UpdateRoster)
	app.Get("/leagues", controller.GetLeagues)
	app.Get("/leagues/:id", controller.GetLeagueByID)
	app.Get("/leagues/:id/rosters", controller.GetRostersByLeagueID)
	app.Delete("/leagues", controller.DeleteLeague)
	return app
}

func (s *LeagueControllerSuite) performJSONRequest(app *fiber.App, method, path, body string) *http.Response {
	req := httptest.NewRequest(method, path, bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")

	res, err := app.Test(req)
	s.Require().NoError(err)
	return res
}

func (s *LeagueControllerSuite) decodeErrorResponse(res *http.Response) map[string]string {
	defer func() {
		s.NoError(res.Body.Close())
	}()

	var resp map[string]string
	s.Require().NoError(json.NewDecoder(res.Body).Decode(&resp))
	return resp
}
