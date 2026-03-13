package controllers

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/tobyrushton/fantasy-nba/pkg/fakes"
	"golang.org/x/crypto/bcrypt"
)

func TestAuth_Register(t *testing.T) {
	t.Run("creates user and returns 201", func(t *testing.T) {
		repo := &fakes.FakeRepo{}
		controller := NewAuthController(repo)

		app := fiber.New()
		app.Post("/register", controller.Register)

		body := []byte(`{"username":"toby","password":"secret"}`)
		req := httptest.NewRequest("POST", "/register", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		res, err := app.Test(req)
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}

		if res.StatusCode != fiber.StatusCreated {
			t.Fatalf("expected status %d, got %d", fiber.StatusCreated, res.StatusCode)
		}

		if repo.CreateUserCallCount() != 1 {
			t.Fatalf("expected CreateUser to be called once, got %d", repo.CreateUserCallCount())
		}

		_, gotUsername, gotHashedPassword := repo.CreateUserArgsForCall(0)
		if gotUsername != "toby" {
			t.Fatalf("expected username %q, got %q", "toby", gotUsername)
		}
		if gotHashedPassword == "secret" {
			t.Fatalf("expected password to be hashed")
		}
		if err := bcrypt.CompareHashAndPassword([]byte(gotHashedPassword), []byte("secret")); err != nil {
			t.Fatalf("expected valid bcrypt hash, compare failed: %v", err)
		}

		var resp map[string]string
		if err := json.NewDecoder(res.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}
		if resp["message"] != "user created" {
			t.Fatalf("expected message %q, got %q", "user created", resp["message"])
		}
	})

	t.Run("returns 400 when body is invalid", func(t *testing.T) {
		repo := &fakes.FakeRepo{}
		controller := NewAuthController(repo)

		app := fiber.New()
		app.Post("/register", controller.Register)

		req := httptest.NewRequest("POST", "/register", bytes.NewReader([]byte("{bad-json")))
		req.Header.Set("Content-Type", "application/json")

		res, err := app.Test(req)
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}

		if res.StatusCode != fiber.StatusBadRequest {
			t.Fatalf("expected status %d, got %d", fiber.StatusBadRequest, res.StatusCode)
		}

		if repo.CreateUserCallCount() != 0 {
			t.Fatalf("expected CreateUser to not be called, got %d", repo.CreateUserCallCount())
		}
	})

	t.Run("returns 400 when repo create fails", func(t *testing.T) {
		repo := &fakes.FakeRepo{}
		repo.CreateUserReturns(nil, errors.New("username already exists"))
		controller := NewAuthController(repo)

		app := fiber.New()
		app.Post("/register", controller.Register)

		body := []byte(`{"username":"toby","password":"secret"}`)
		req := httptest.NewRequest("POST", "/register", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		res, err := app.Test(req)
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}

		if res.StatusCode != fiber.StatusBadRequest {
			t.Fatalf("expected status %d, got %d", fiber.StatusBadRequest, res.StatusCode)
		}

		if repo.CreateUserCallCount() != 1 {
			t.Fatalf("expected CreateUser to be called once, got %d", repo.CreateUserCallCount())
		}

		var resp map[string]string
		if err := json.NewDecoder(res.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}
		if resp["message"] != "username already exists" {
			t.Fatalf("expected message %q, got %q", "username already exists", resp["message"])
		}
	})
}
