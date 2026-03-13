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

func (s *LeagueControllerSuite) newApp(repo *fakes.FakeRepo) *fiber.App {
	controller := NewLeagueController(repo)
	app := fiber.New()
	app.Post("/leagues", controller.CreateLeague)
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