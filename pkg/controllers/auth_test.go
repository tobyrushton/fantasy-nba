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
	"github.com/tobyrushton/fantasy-nba/pkg/token"
	"golang.org/x/crypto/bcrypt"
)

type AuthControllerSuite struct {
	suite.Suite

	secret string
}

func TestAuthControllerSuite(t *testing.T) {
	suite.Run(t, new(AuthControllerSuite))
}

func (s *AuthControllerSuite) SetupTest() {
	s.secret = "test-secret"
}

func (s *AuthControllerSuite) TestRegisterCreatesUserAndReturns201() {
	repo := &fakes.FakeRepo{}
	res := s.performJSONRequest(s.newApp(repo), http.MethodPost, "/register", `{"username":"toby","password":"secret"}`)

	s.Equal(fiber.StatusCreated, res.StatusCode)
	s.Equal(1, repo.CreateUserCallCount())

	_, gotUsername, gotHashedPassword := repo.CreateUserArgsForCall(0)
	s.Equal("toby", gotUsername)
	s.NotEqual("secret", gotHashedPassword)
	s.NoError(bcrypt.CompareHashAndPassword([]byte(gotHashedPassword), []byte("secret")))

	resp := s.decodeJSON(res)
	s.Equal("user created", resp["message"])
}

func (s *AuthControllerSuite) TestRegisterReturns400WhenBodyIsInvalid() {
	repo := &fakes.FakeRepo{}
	res := s.performJSONRequest(s.newApp(repo), http.MethodPost, "/register", "{bad-json")

	s.Equal(fiber.StatusBadRequest, res.StatusCode)
	s.Equal(0, repo.CreateUserCallCount())
}

func (s *AuthControllerSuite) TestRegisterReturns400WhenRepoCreateFails() {
	repo := &fakes.FakeRepo{}
	repo.CreateUserReturns(nil, errors.New("username already exists"))

	res := s.performJSONRequest(s.newApp(repo), http.MethodPost, "/register", `{"username":"toby","password":"secret"}`)

	s.Equal(fiber.StatusBadRequest, res.StatusCode)
	s.Equal(1, repo.CreateUserCallCount())

	resp := s.decodeJSON(res)
	s.Equal("username already exists", resp["message"])
}

func (s *AuthControllerSuite) TestLoginReturnsTokenForValidCredentials() {
	repo := &fakes.FakeRepo{}
	repo.GetUserByUsernameReturns(&models.User{
		ID:           42,
		Username:     "toby",
		PasswordHash: generatePasswordHash("secret"),
	}, nil)

	res := s.performJSONRequest(s.newApp(repo), http.MethodPost, "/login", `{"username":"toby","password":"secret"}`)

	s.Equal(fiber.StatusOK, res.StatusCode)
	s.Equal(1, repo.GetUserByUsernameCallCount())

	_, gotUsername := repo.GetUserByUsernameArgsForCall(0)
	s.Equal("toby", gotUsername)

	resp := s.decodeJSON(res)
	gotToken := resp["token"]
	s.NotEmpty(gotToken)

	userID, err := token.VerifyToken(s.secret, gotToken)
	s.NoError(err)
	s.Equal(int64(42), userID)
}

func (s *AuthControllerSuite) TestLoginReturns400WhenBodyIsInvalid() {
	repo := &fakes.FakeRepo{}
	res := s.performJSONRequest(s.newApp(repo), http.MethodPost, "/login", "{bad-json")

	s.Equal(fiber.StatusBadRequest, res.StatusCode)
	s.Equal(0, repo.GetUserByUsernameCallCount())
}

func (s *AuthControllerSuite) TestLoginReturns400WhenUserLookupFails() {
	repo := &fakes.FakeRepo{}
	repo.GetUserByUsernameReturns(nil, errors.New("not found"))

	res := s.performJSONRequest(s.newApp(repo), http.MethodPost, "/login", `{"username":"toby","password":"secret"}`)

	s.Equal(fiber.StatusBadRequest, res.StatusCode)

	resp := s.decodeJSON(res)
	s.Equal("invalid username or password", resp["message"])
}

func (s *AuthControllerSuite) TestLoginReturns400WhenPasswordDoesNotMatch() {
	repo := &fakes.FakeRepo{}
	repo.GetUserByUsernameReturns(&models.User{
		ID:           42,
		Username:     "toby",
		PasswordHash: generatePasswordHash("different-secret"),
	}, nil)

	res := s.performJSONRequest(s.newApp(repo), http.MethodPost, "/login", `{"username":"toby","password":"secret"}`)

	s.Equal(fiber.StatusBadRequest, res.StatusCode)

	resp := s.decodeJSON(res)
	s.Equal("invalid username or password", resp["message"])
}

func (s *AuthControllerSuite) newApp(repo *fakes.FakeRepo) *fiber.App {
	controller := NewAuthController(repo, s.secret)
	app := fiber.New()
	app.Post("/register", controller.Register)
	app.Post("/login", controller.Login)
	return app
}

func (s *AuthControllerSuite) performJSONRequest(app *fiber.App, method, path, body string) *http.Response {
	req := httptest.NewRequest(method, path, bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")

	res, err := app.Test(req)
	s.Require().NoError(err)
	return res
}

func (s *AuthControllerSuite) decodeJSON(res *http.Response) map[string]string {
	s.T().Cleanup(func() {
		s.NoError(res.Body.Close())
	})

	var resp map[string]string
	err := json.NewDecoder(res.Body).Decode(&resp)
	s.Require().NoError(err)
	return resp
}
