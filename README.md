# PokeDB

Full-stack Pokémon encyclopedia built with Go/Gin, HTMX, and TailwindCSS. Features complete CRUD operations, live search, pagination, and automatic seeding from the PokéAPI.

![Stack](https://img.shields.io/badge/Go-1.22+-00ADD8?logo=go)
![Gin](https://img.shields.io/badge/Gin-1.12-00897B?logo=go)
![HTMX](https://img.shields.io/badge/HTMX-2.0-4F46E5)
![TailwindCSS](https://img.shields.io/badge/TailwindCSS-3.0-38B2AC)
![SQLite](https://img.shields.io/badge/SQLite-3-003B57)

---

## Features

- Full CRUD (Create, Read, Update, Delete) for Pokémon
- Real-time search with HTMX (no page reload)
- Pagination (12 per page, HTMX-powered)
- Auto-seed from PokéAPI (concurrent, ~1000 Pokémon in 1-2 min)
- SQLite database with auto-migration
- Dark themed, responsive UI with TailwindCSS
- Single binary deployment

---

## Quick Start

```bash
# 1. Navigate to project directory


# 2. Install dependencies
go mod tidy

# 3. Seed database (Gen 1 + more, ~30s)
go run main.go -seed -limit 151

# 4. Start server
go run main.go

# 5. Open http://localhost:8080
```

---

## CLI Usage

```bash
# Seed only (no server)
go run main.go -seed -limit 1000

# Seed all Pokémon (1351, ~10+ min)
go run main.go -seed -limit 0

# Run server normally
go run main.go

# Custom port
PORT=3000 go run main.go

# Build binary
go build -o pokedex-server .
./pokedex-server

# Re-seed (wipes old data)
rm pokedex.db && go run main.go -seed -limit 1000
```

---

## Tech Stack

| Layer | Technology |
|-------|------------|
| Language | Go 1.22+ |
| Web Framework | Gin v1.12 |
| Frontend | HTMX 2.0 (no JS build) |
| CSS | TailwindCSS v3.0 |
| Database | SQLite (go-sqlite3) |
| API | PokéAPI v2 (pokeapi.co) |

---

## API Endpoints

| Method | Route | Description |
|--------|-------|-------------|
| GET | `/` | Main page (Pokédex grid) |
| GET | `/pokemon?page=N` | Paginated list (12 per page) |
| POST | `/pokemon/search` | Search by name |
| GET | `/pokemon/create` | Create form |
| POST | `/pokemon/create` | Store new Pokémon |
| GET | `/pokemon/:id` | Detail page |
| GET | `/pokemon/:id/edit` | Edit form |
| POST | `/pokemon/:id/update` | Update Pokémon |
| DELETE | `/pokemon/:id` | Delete Pokémon |
| GET | `/stats` | Total count (HTMX target) |

---

## Database Schema

```sql
CREATE TABLE pokemon (
  id              INTEGER PRIMARY KEY AUTOINCREMENT,
  name            TEXT NOT NULL UNIQUE,
  type1           TEXT NOT NULL,
  type2           TEXT,
  height          TEXT NOT NULL,
  weight          TEXT NOT NULL,
  base_experience INTEGER NOT NULL DEFAULT 0,
  image_url       TEXT,
  description     TEXT,
  evolves_from    TEXT,
  created_at      DATETIME DEFAULT CURRENT_TIMESTAMP,
  updated_at      DATETIME DEFAULT CURRENT_TIMESTAMP
);
```

---

## Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | `8080` | Server port |

---

## Development

```bash
# Live reload (install air first)
go install github.com/cosmtrek/air@latest
air

# Or just run directly
go run main.go

# Test build
go build ./...
go vet ./...
```

---

## Production Deployment

```bash
# 1. Build binary
go build -o pokedex-server .

# 2. Run
./pokedex-server -port 8080
```

### systemd service (`/etc/systemd/system/pokedex.service`)

```ini
[Unit]
Description=PokeDB Server
After=network.target

[Service]
Type=simple
User=pokedex
WorkingDirectory=/opt/pokedex
ExecStart=/opt/pokedex/pokedex-server
Restart=always
Environment=PORT=8080

[Install]
WantedBy=multi-user.target
```

### Docker

```dockerfile
FROM golang:1.22-alpine AS build
WORKDIR /app
COPY . .
RUN go mod tidy && go build -o pokedex-server .

FROM alpine:latest
WORKDIR /app
COPY --from=build /app/pokedex-server .
COPY --from=build /app/templates ./templates
COPY --from=build /app/pokedex.db .
EXPOSE 8080
CMD ["./pokedex-server"]
```

---

## Project Structure

```
pokedex-go/
├── main.go              # Entry point + routing
├── database/
│   ├── db.go            # Connection + migration
│   └── crud.go          # CRUD + pagination + bulk insert
├── handlers/
│   └── handlers.go      # Gin handlers + HTML rendering
├── models/
│   └── pokemon.go       # Data model
├── seed/
│   └── seed.go          # PokéAPI seeder (concurrent)
├── templates/
│   ├── index.html       # Main grid page
│   ├── pokemon_show.html # Detail page
│   ├── pokemon_form.html # Create/Edit form
│   ├── list_empty.html  # Empty state
│   ├── partials.html    # Reusable partials
│   └── stats.html       # Stats endpoint
├── go.mod
├── go.sum
└── pokedex.db           # SQLite database (auto-created)
```

---

## Troubleshooting

| Issue | Solution |
|-------|----------|
| Server exits silently with old data | Remove old DB: `rm pokedex.db && go run main.go -seed -limit 1000` |
| Port already in use | Kill process: `pkill -f pokedex-server` or change `PORT` |
| CGO errors (SQLite) | Install C compiler: `gcc --version` / `xcode-select --install` |
| Seeding slow | Increase workers in `seed/seed.go` (default 20) |

---

## License

MIT
