package main

import (
	"fmt"
	"log"

	swagger "github.com/gofiber/contrib/v3/swaggerui"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/tobyrushton/fantasy-nba/pkg/config"
	"github.com/tobyrushton/fantasy-nba/pkg/controllers"
	"github.com/tobyrushton/fantasy-nba/pkg/db"
	"github.com/tobyrushton/fantasy-nba/pkg/db/models"
)

// @title Fantasy NBA API
// @version 1.0
// @description API for authentication, players, leagues, and rosters.
// @BasePath /
// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization

func main() {
	cfg := config.MustLoadConfig()

	db, err := db.NewDb(fmt.Sprintf("postgres://admin:%s@localhost:5432/postgres?sslmode=disable", cfg.DB_PASSWORD))
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

	repo := models.NewPostgresRepo(db)

	app := fiber.New()
	app.Use(cors.New(cors.Config{
		AllowOrigins: []string{"http://localhost:4321", "http://127.0.0.1:4321", "http://localhost:3000", "http://127.0.0.1:3000"},
		AllowHeaders: []string{"Origin", "Content-Type", "Accept", "Authorization"},
		AllowMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
	}))

	// set up swagger
	swaggerCfg := swagger.Config{
		FilePath: "./docs/swagger.json",
		Path:     "docs",
	}

	app.Use(swagger.New(swaggerCfg))

	// Set up routes

	// Auth routes
	ac := controllers.NewAuthController(repo, cfg.JWT_SECRET)
	auth := app.Group("/auth")
	auth.Post("/register", ac.Register)
	auth.Post("/login", ac.Login)

	// Player routes
	pc := controllers.NewPlayersController(repo)
	players := app.Group("/players")
	players.Get("/", pc.GetPlayers)
	players.Get("/:id", pc.GetPlayer)

	// league routes
	lc := controllers.NewLeagueController(repo)
	league := app.Group("/league")
	league.Use(ac.Middleware)
	league.Post("/", lc.CreateLeague)
	league.Get("/", lc.GetLeagues)
	league.Post("/join", lc.JoinLeague)
	league.Post("/roster", lc.CreateRoster)
	league.Put("/roster", lc.UpdateRoster)
	league.Delete("/", lc.DeleteLeague)
	league.Get("/:id/rosters", lc.GetRostersByLeagueID)
	league.Get("/:id", lc.GetLeagueByID)

	log.Fatal(app.Listen(":8080"))
}
