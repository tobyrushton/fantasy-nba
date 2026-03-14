# Fantasy NBA API

Fantasy NBA API is a Go + Fiber backend for a fantasy basketball application.
It provides:
- User authentication with JWT
- Player lookup and player stat retrieval
- League creation and league membership
- Roster creation and roster updates
- Seed tooling to scrape NBA teams, players, schedules, and game stats

## Tech Stack

- Go
- Fiber v3
- Bun ORM
- PostgreSQL
- Swagger/OpenAPI

## Project Structure

- main API server: [main.go](main.go)
- DB migration command: [cmd/migrate/main.go](cmd/migrate/main.go)
- DB seed command: [cmd/seed/main.go](cmd/seed/main.go)
- controllers: [pkg/controllers](pkg/controllers)
- Bun models: [pkg/db/models/models.go](pkg/db/models/models.go)
- generated API spec: [docs/swagger.json](docs/swagger.json), [docs/swagger.yaml](docs/swagger.yaml)

## Database Models

The Bun models are defined in [pkg/db/models/models.go](pkg/db/models/models.go).

### Core Basketball Data

- Team
  - Fields: id, nba_id, name, abbreviation, created_at, updated_at
  - Table: teams

- Player
  - Fields: id, nba_id, first_name, last_name, position, team_id, created_at, updated_at
  - Table: players
  - Relation: belongs to Team

- Game
  - Fields: id, nba_id, season, game_date, home_team_id, away_team_id, home_score, away_score, created_at, updated_at
  - Table: games
  - Relations: belongs to HomeTeam and AwayTeam

- PlayerGameStats
  - Fields: id, player_id, game_id, team_id, did_not_play, points, rebounds, assists, steals, blocks, turnovers, three_pointers_made, free_throws_made, created_at, updated_at
  - Table: player_game_stats
  - Relations: belongs to Player, Game, Team

### Fantasy App Data

- User
  - Fields: id, username, password_hash
  - Table: users

- League
  - Fields: id, name, creator_id
  - Table: leagues
  - Relation: creator (User)

- LeagueMembership
  - Fields: id, league_id, user_id
  - Table: league_memberships
  - Relations: League, User

- TeamRoster
  - Fields: id, league_id, user_id, player_id
  - Table: team_rosters
  - Relations: League, User, Player

## API Documentation

Swagger/OpenAPI docs are generated into the docs directory.

- JSON spec: [docs/swagger.json](docs/swagger.json)
- YAML spec: [docs/swagger.yaml](docs/swagger.yaml)

When the API is running, Swagger UI is available at:

- http://localhost:8080/docs

## Environment Variables

The app loads environment variables from .env.

Required values:

- DB_PASSWORD
- JWT_SECRET

Example .env:

DB_PASSWORD=your_postgres_password
JWT_SECRET=your_jwt_secret

## How To Run

### 1. Start PostgreSQL

Option A: use the provided script

sh start.sh

Option B: run manually

docker build -t fantasy-nba-db .

docker run -d \
  --name fantasy-nba-db \
  -e POSTGRES_USER=admin \
  -e POSTGRES_PASSWORD=$DB_PASSWORD \
  -e POSTGRES_DB=mydb \
  -p 5432:5432 \
  fantasy-nba-db

### 2. Install dependencies

go mod download

### 3. Run database migrations

go run cmd/migrate/main.go

### 4. (Optional) Seed database with scraped data

go run cmd/seed/main.go

### 5. Start the API server

go run main.go

The API runs on:

- http://localhost:8080

## Main API Routes

- Auth
  - POST /auth/register
  - POST /auth/login

- Players
  - GET /players
  - GET /players/:id

- League (JWT protected)
  - POST /league
  - GET /league
  - GET /league/:id
  - DELETE /league
  - POST /league/join
  - POST /league/roster
  - PUT /league/roster
  - GET /league/:id/rosters

## Testing

Run all tests:

go test ./...
