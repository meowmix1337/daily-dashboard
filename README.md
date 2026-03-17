# Daily Dashboard

A personal daily life dashboard with a Go backend and React/TypeScript frontend. Displays weather, calendar, tasks, news, stocks/crypto, and a daily quote — all in a single dark-themed view.

![screenshot placeholder](docs/screenshot.png)

## Tech Stack

| Layer | Technology |
|-------|-----------|
| Backend | Go 1.23 + chi router |
| Frontend | React 19 + TypeScript + Vite 6 |
| Styling | Tailwind CSS 4 |
| State | TanStack Query v5 |
| Containers | Docker + Docker Compose v2 |
| Proxy | Nginx 1.27 |

## Architecture

```
Browser → Nginx (frontend) → Go API (backend)
                                 ├─ Open-Meteo        (weather, no auth)
                                 ├─ GNews             (headlines, API key)
                                 ├─ Finnhub           (stocks, API key)
                                 ├─ CoinGecko         (BTC, no auth)
                                 ├─ Sunrise-Sunset.org (no auth)
                                 ├─ API Ninjas        (daily quote, API key)
                                 └─ ICS Calendar      (today's events, URL)
```

## Quick Start

```bash
cp .env.example .env
# Edit .env and add your API keys
make docker-up
# Open http://localhost:3000
```

## Local Development

**Terminal 1 — backend:**
```bash
make dev-backend
# Starts Go server at http://localhost:8080
```

**Terminal 2 — frontend:**
```bash
cd frontend && npm install
make dev-frontend
# Starts Vite dev server at http://localhost:5173
```

## Obtaining API Keys & Config

Cards that require credentials show an unavailable state when their key/URL is not set — no crashes or mock data.

### GNews (headlines)

1. Go to [gnews.io](https://gnews.io) and click **Get API Key**
2. Sign up for a free account (100 requests/day on the free tier)
3. Copy the key from your dashboard and set `GNEWS_API_KEY` in `.env`

### Finnhub (stocks)

1. Go to [finnhub.io](https://finnhub.io) and click **Get free API key**
2. Sign up — no credit card required (60 requests/minute on the free tier)
3. Copy the key from **Dashboard → API Key** and set `FINNHUB_API_KEY` in `.env`

### API Ninjas (daily quote)

1. Go to [api-ninjas.com](https://api-ninjas.com) and create a free account
2. Copy your API key from the dashboard and set `API_NINJAS_API_KEY` in `.env`

### Calendar (ICS URL)

The calendar card reads any standard ICS/iCal feed — no OAuth required.

**Google Calendar:**
1. Open [Google Calendar](https://calendar.google.com) → Settings (gear icon)
2. Select the calendar you want to display under **Settings for my calendars**
3. Scroll to **Integrate calendar** → copy the **Secret address in iCal format**
4. Set `CALENDAR_ICS_URL` in `.env` to that URL

**Apple iCloud Calendar:**
1. In the Calendar app, right-click (or Ctrl-click) the calendar → **Share Calendar**
2. Enable **Public Calendar** and copy the link
3. Replace `webcal://` with `https://` and set `CALENDAR_ICS_URL` in `.env`

> The ICS URL is private — treat it like a password. Calendar data is cached for 15 minutes.

## Environment Variables

| Variable | Description | Required | Default |
|----------|-------------|----------|---------|
| `GNEWS_API_KEY` | GNews API key for headlines | Optional | — (unavailable state) |
| `FINNHUB_API_KEY` | Finnhub API key for stocks | Optional | — (unavailable state) |
| `API_NINJAS_API_KEY` | API Ninjas key for daily quote | Optional | — (unavailable state) |
| `CALENDAR_ICS_URL` | Private ICS URL for calendar events | Optional | — (unavailable state) |
| `LATITUDE` | Your location latitude | No | 37.7749 (SF) |
| `LONGITUDE` | Your location longitude | No | -122.4194 (SF) |
| `TIMEZONE` | IANA timezone for calendar date filtering (e.g. `America/New_York`) | No | server local (UTC in Docker) |
| `PORT` | Backend server port | No | 8080 |
| `CORS_ORIGIN` | Allowed CORS origin for the API (set to your frontend URL in production) | No | `http://localhost:5173` |

## API Endpoints

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/health` | Health check |
| GET | `/api/dashboard` | All data aggregated |
| GET | `/api/weather` | Current weather + forecast |
| GET | `/api/news` | Top headlines |
| GET | `/api/stocks` | Stock + crypto quotes |
| GET | `/api/calendar` | Today's events |
| GET | `/api/tasks` | Task list |
| POST | `/api/tasks` | Create task |
| PATCH | `/api/tasks/{id}` | Toggle task done |
| DELETE | `/api/tasks/{id}` | Delete task |
| GET | `/api/meta` | Sunrise/sunset + quote |

## Project Structure

```
daily-dashboard/
├── backend/
│   ├── cmd/server/main.go        # Entrypoint
│   └── internal/
│       ├── config/config.go      # Env config
│       ├── handler/              # HTTP handlers
│       ├── middleware/           # CORS + logging
│       ├── model/models.go       # Shared types
│       ├── server/server.go      # Router + wiring
│       └── service/              # Business logic + API clients
├── frontend/
│   └── src/
│       ├── api/client.ts         # API fetch wrapper
│       ├── components/           # React components
│       ├── hooks/                # React Query hooks
│       └── types/dashboard.ts    # TypeScript interfaces
├── deploy/
│   ├── backend.Dockerfile
│   ├── frontend.Dockerfile
│   └── nginx.conf
├── docker-compose.yml
├── .env.example
└── Makefile
```

## Deployment

Deploy on any Linux VPS:

```bash
git clone <your-repo>
cd daily-dashboard
cp .env.example .env
# Fill in API keys
docker compose up -d
```

The stack runs on ports 3000 (frontend) and 8080 (backend). Point a reverse proxy at port 3000.
