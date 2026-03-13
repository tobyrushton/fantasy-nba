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

	var resp models.League
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

	var resp []models.League
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

	var resp models.League
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

func (s *LeagueControllerSuite) newApp(repo *fakes.FakeRepo) *fiber.App {
	controller := NewLeagueController(repo)
	app := fiber.New()
	app.Post("/leagues", controller.CreateLeague)
	app.Get("/leagues", controller.GetLeagues)
	app.Get("/leagues/:id", controller.GetLeagueByID)
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
