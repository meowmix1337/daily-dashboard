@AGENTS.md

# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

### Backend (run from `backend/`)
```bash
go run ./cmd/server          # Start dev server on :8080
go build ./...               # Compile all packages
go vet ./...                 # Lint
go test -race ./...          # Run tests
go build -o bin/server ./cmd/server  # Build binary
```

### Frontend (run from `frontend/`)
```bash
npm run dev      # Vite dev server on :5173 (proxies /api → :8080)
npm run build    # TypeScript check + production build → dist/
npm run lint     # ESLint
```

### Make shortcuts (from repo root)
```bash
make dev-backend     # go run ./cmd/server
make dev-frontend    # npm run dev
make docker-up       # docker compose up -d (production)
make docker-dev      # docker compose with hot-reload overrides
make lint            # go vet + npm run lint
make test            # go test -race ./...
```

## Architecture

### Request Flow
```
Browser → Vite proxy (:5173/api) → Go API (:8080)   [dev]
Browser → Nginx (:3000/api)      → Go API (:8080)   [prod]
```

The frontend calls `GET /api/dashboard` which fans out to all backend services concurrently via `errgroup` and returns a single `model.DashboardResponse`. Individual services fail gracefully — a failed service returns a zero value, not an error response.

### Backend Structure

**Wiring**: All dependency injection happens in `internal/server/server.go:setupRoutes()`. Services and handlers are constructed there with a shared `*http.Client` (30s timeout) and a shared `*service.CacheService`.

**Service pattern**: Every service follows the same shape:
```go
type XxxService struct { httpClient *http.Client; apiKey string; cache *CacheService }
func NewXxxService(...) *XxxService
func (s *XxxService) Fetch(ctx context.Context) (model.XxxData, error)
```
Each `Fetch` checks the cache first, calls the external API on miss, and returns an error (card shows unavailable state) if the API key is absent or the call fails. Cache TTLs: weather 15m, stocks 10s, news 3h, calendar 15m, sunrise 6h, quotes 24h.

**Adding a new widget**: create `internal/service/foo.go` → `internal/handler/foo.go` → register route in `server.go` → add field to `model.DashboardResponse` → add goroutine in `handler/dashboard.go`.

### Frontend Structure

**Data flow**: `App.tsx` wraps everything in `QueryClientProvider`. `Dashboard.tsx` calls `useDashboard()` (60s stale, 30s refetch) and passes typed slices/structs down to each card component as props. Cards never fetch data themselves.

**State**: Only `TasksCard` has mutations — `useTasks.ts` wraps `PATCH /api/tasks/{id}` with optimistic updates against the `['dashboard']` query cache.

**Styling**: Components use inline styles exclusively (no Tailwind classes in JSX). Tailwind is imported in `index.css` via `@import "tailwindcss"` but the design system is all inline to match the exact pixel values from the mock.

**UI primitives**: `components/ui/Card.tsx` handles the glass-morphism card shell and staggered fade-in animation. `CardHeader.tsx` and `MiniStat.tsx` are the other shared primitives — import these for any new card.

### External APIs
| Service | API | Key env var | Fallback |
|---------|-----|-------------|---------|
| Weather | Open-Meteo + AQI | none | unavailable state |
| News | GNews (9 categories, sequential 1 req/s) | `GNEWS_API_KEY` | unavailable state |
| Stocks | Finnhub (equities) + CoinGecko (BTC) | `FINNHUB_API_KEY` | unavailable state |
| Calendar | ICS feed (parsed with golang-ical) | `CALENDAR_ICS_URL` | unavailable state |
| Sunrise | sunrise-sunset.org | none | unavailable state |
| Quote | api.api-ninjas.com/v2/quotes | `API_NINJAS_API_KEY` | unavailable state |

### Config (env vars → `internal/config/config.go`)
`PORT` (default 8080), `GNEWS_API_KEY`, `FINNHUB_API_KEY`, `API_NINJAS_API_KEY`, `CALENDAR_ICS_URL`, `LATITUDE`/`LONGITUDE` (default SF 37.7749/-122.4194), `TIMEZONE` (IANA tz name, e.g. `America/New_York`; defaults to server local time — required for correct calendar event filtering).

## Known Limitations

- **Tasks are in-memory only** — lost on server restart. No database yet.
- **Stocks watchlist is in-memory only** — symbol additions/removals are lost on restart.
- **No authentication** — all endpoints are public; intended for personal/LAN use.
- **No tests** — zero test files currently exist in backend or frontend.
- **News uses sequential fetching** — GNews free tier requires ~1 req/s; 9 categories × 3h cache means full refresh takes ~9s on cache miss.

## Documentation Rule

**Always update `README.md` and `.env.example` when:**
- A new env var is added or removed (update the Environment Variables table and `.env.example`)
- A new external API or service is added or removed (update the Architecture section and Obtaining API Keys)
- The behavior of an existing service changes in a user-visible way

Keep `README.md` as the single source of truth for setup instructions.
