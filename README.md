# Fantasy NBA

Fantasy NBA is a full-stack fantasy basketball app with a Go + Fiber API and an Astro + React frontend.

The backend handles authentication, player data, leagues, memberships, and roster management. The frontend provides the player explorer, player stat pages, auth screens, protected league pages, and roster creation and editing flows.

## Features

- Browse NBA players from the frontend home page
- Search and filter players by name and position
- Open individual player pages with game-by-game stats
- Display player games by date, newest first
- Exclude DNP games from average stat calculations
- Register and log in with JWT-based authentication
- Create leagues and join existing leagues
- View protected league pages when signed in
- Create a 10-player fantasy roster inside a league
- Update an existing roster by swapping players in and out
- Delete a league if you are the league owner

## Tech Stack

### Backend

- Go
- Fiber v3
- Bun ORM
- PostgreSQL
- Swagger / OpenAPI

### Frontend

- Astro 5
- React 19
- TypeScript
- Tailwind CSS 4
- shadcn/ui

## Project Structure

- API server: [main.go](main.go)
- Migration command: [cmd/migrate/main.go](cmd/migrate/main.go)
- Seed command: [cmd/seed/main.go](cmd/seed/main.go)
- Controllers: [pkg/controllers](pkg/controllers)
- Bun models: [pkg/db/models/models.go](pkg/db/models/models.go)
- Swagger docs: [docs/swagger.json](docs/swagger.json), [docs/swagger.yaml](docs/swagger.yaml)
- Frontend app: [www](www)
- Frontend pages: [www/src/pages](www/src/pages)
- Frontend components: [www/src/components](www/src/components)
- Frontend API client: [www/src/client/api.ts](www/src/client/api.ts)

## Frontend Pages

- `/`
  - Player explorer with client-side search and filtering
  - Auth navigation for login, registration, and leagues access
- `/players/[id]`
  - Player stat detail page
  - Game dates shown instead of game numbers
  - Games sorted from most recent to oldest
  - Average row excludes games marked as DNP
- `/login`
  - Login form for existing users
- `/register`
  - Registration form for new users
- `/leagues`
  - Protected leagues dashboard
  - Create a league, view leagues, and join leagues
- `/leagues/[id]`
  - Protected league detail page
  - View league members and rosters
  - Create and update your roster
  - Owner-only league deletion

## Database Models

The Bun models are defined in [pkg/db/models/models.go](pkg/db/models/models.go).

### Core Basketball Data

- Team
  - Fields: `id`, `nba_id`, `name`, `abbreviation`, `created_at`, `updated_at`
  - Table: `teams`

- Player
  - Fields: `id`, `nba_id`, `first_name`, `last_name`, `position`, `team_id`, `created_at`, `updated_at`
  - Table: `players`
  - Relation: belongs to `Team`

- Game
  - Fields: `id`, `nba_id`, `season`, `game_date`, `home_team_id`, `away_team_id`, `home_score`, `away_score`, `created_at`, `updated_at`
  - Table: `games`
  - Relations: belongs to `HomeTeam` and `AwayTeam`

- PlayerGameStats
  - Fields: `id`, `player_id`, `game_id`, `team_id`, `did_not_play`, `points`, `rebounds`, `assists`, `steals`, `blocks`, `turnovers`, `three_pointers_made`, `free_throws_made`, `created_at`, `updated_at`
  - Table: `player_game_stats`
  - Relations: belongs to `Player`, `Game`, and `Team`

### Fantasy App Data

- User
  - Fields: `id`, `username`, `password_hash`
  - Table: `users`

- League
  - Fields: `id`, `name`, `creator_id`
  - Table: `leagues`
  - Relation: creator (`User`)

- LeagueMembership
  - Fields: `id`, `league_id`, `user_id`
  - Table: `league_memberships`
  - Relations: `League`, `User`

- TeamRoster
  - Fields: `id`, `league_id`, `user_id`, `player_id`
  - Table: `team_rosters`
  - Relations: `League`, `User`, `Player`

## Environment Variables

### Backend

The API loads environment variables from a root `.env` file.

Required values:

- `DB_PASSWORD`
- `JWT_SECRET`

Example:

```env
DB_PASSWORD=your_postgres_password
JWT_SECRET=your_jwt_secret
```

### Frontend

The frontend expects a public API base URL.

Create `www/.env` with:

```env
PUBLIC_API_URL=http://localhost:8080
```

## Running The App

### 1. Start PostgreSQL

Option A: use the provided script

```bash
sh start.sh
```

Option B: run Docker manually

```bash
docker build -t fantasy-nba-db .

docker run -d \
  --name fantasy-nba-db \
  -e POSTGRES_USER=admin \
  -e POSTGRES_PASSWORD=$DB_PASSWORD \
  -e POSTGRES_DB=mydb \
  -p 5432:5432 \
  fantasy-nba-db
```

### 2. Install Backend Dependencies

```bash
go mod download
```

### 3. Run Database Migrations

```bash
go run cmd/migrate/main.go
```

### 4. Optionally Seed NBA Data

```bash
go run cmd/seed/main.go
```

### 5. Start the API Server

```bash
go run main.go
```

The API runs on `http://localhost:8080`.

Swagger UI is available at `http://localhost:8080/docs`.

### 6. Install Frontend Dependencies

```bash
cd www
npm install
```

### 7. Start the Frontend

```bash
cd www
npm run dev
```

The frontend runs on Astro's local dev server, typically `http://localhost:4321`.

## API Routes

### Public

- `POST /auth/register`
- `POST /auth/login`
- `GET /players/`
- `GET /players/:id`

### JWT Protected

- `POST /league/`
- `GET /league/`
- `GET /league/:id`
- `DELETE /league/`
- `POST /league/join`
- `POST /league/roster`
- `PUT /league/roster`
- `GET /league/:id/rosters`

## Frontend Scripts

From [www](www):

- `npm run dev` to start the Astro dev server
- `npm run build` to create a production build
- `npm run preview` to preview the production build locally
- `npm run lint` to run ESLint
- `npm run typecheck` to run Astro type checking

## Testing

Run backend tests from the repo root:

```bash
go test ./...
```

Run frontend lint from `www`:

```bash
npm run lint
```

## Notes

- League pages require a valid JWT obtained from the login flow.
- The frontend stores the auth token in the browser and sends it on protected league requests.
- League creators are automatically added as league members when a league is created.
