# Testing Strategy RFC — Daily Dashboard

## Status

Draft — pending review and approval

---

## Summary

This RFC recommends a **hybrid/layered approach** anchored on Go unit tests for the backend's
genuinely complex pure logic (weather parsing, calendar filtering, AQI categorisation, duration
formatting, stocks watchlist mutations) plus HTTP-level handler tests using `net/http/httptest`,
and Vitest + Testing Library component/hook tests on the frontend for the two areas with real
behavioural complexity: the `useTasks` optimistic-update mutations and the `useSymbolSearch`
debounce logic. E2E and contract testing are explicitly deferred — they add disproportionate
infrastructure overhead relative to the value they provide for a single-developer personal
dashboard that already has strong type safety between frontend and backend.

---

## Current State

### Backend (Go 1.26, chi router, sqlx + modernc/sqlite)

| Layer | Files | Complexity verdict |
|---|---|---|
| `service/weather.go` | 232 lines | **Medium** — WMO code lookup table, hourly strip filtering via string-compare time trick, AQI categorisation, UV index extraction from hourly array |
| `service/calendar.go` | 207 lines | **High** — ICS parsing, all-day vs timed event detection, TZID/UTC handling, today-filter, sort (all-day first then chronological), duration formatting |
| `service/stocks.go` | 289 lines | **Medium** — concurrent Finnhub + CoinGecko fan-out with errgroup, watchlist mutex, add/remove with sentinel errors, symbol search capping at 10 results |
| `service/news.go` | 152 lines | **Low-medium** — sequential fetch with 1 s sleep, `relativeTime` string logic |
| `service/tasks.go` | 78 lines | **Low** — thin in-memory CRUD wrapper over a slice with mutex; being replaced by SQLite |
| `service/cache.go` | 66 lines | **Low** — `sync.Map` TTL cache with background eviction; straightforward but concurrent |
| `handler/dashboard.go` | 121 lines | **Low** — errgroup fan-out to services; always returns nil errors; serialises to JSON |
| `handler/stocks.go` | 120 lines | **Low-medium** — request decoding, sentinel error → HTTP status mapping, rate limiting middleware |
| `handler/tasks.go` | 79 lines | **Low** — thin REST wrapper |

**No test files exist anywhere in the backend.**

No test helper libraries are present in `go.mod` beyond the standard library — `testing`,
`net/http/httptest`, and `encoding/json` are all that is needed to start.

### Frontend (React 19, TypeScript, Vite 7, TanStack Query v5)

| File | Complexity verdict |
|---|---|
| `hooks/useTasks.ts` | **High** — three mutations (toggle, create, delete) each with full optimistic-update / rollback pattern against the `['dashboard']` query cache |
| `hooks/useSymbolSearch.ts` | **Medium** — manual debounce via `useRef` + `setTimeout`, wires into TanStack Query with `enabled` guard |
| `hooks/useDashboard.ts` | **Low** — thin `useQuery` wrapper, gated on `isAuthenticated` |
| `hooks/useAuth.ts` | **Low** — session state, no client-side logic |
| `api/client.ts` | **Low** — `apiFetch` wrapper, typed call functions |
| `components/TasksCard.tsx` | **Medium** — optimistic UI, form state, delete confirmation |
| `components/StocksCard.tsx` | **Medium** — symbol search dropdown, add/remove watchlist, debounced query display |
| Other cards | **Low** — mostly presentational, render props passed from Dashboard |

**No test files, no test runner, no test dependencies installed.**

---

## Problem Statement

The project has zero automated tests. The risk profile is:

1. **Calendar date filtering is the highest-risk code path.** The all-day vs timed distinction,
   TZID handling, and the `filterToday` logic has multiple branches that are easy to break
   silently — a wrong event appears or today's events disappear.

2. **Optimistic updates in `useTasks` are hard to reason about.** The three-way cancel/setQueryData/
   rollback pattern across three mutations is the most complex frontend code in the project. A
   regression here means tasks appear to toggle then snap back, or a phantom temp task stays
   visible after a server error.

3. **Watchlist mutation logic has subtle concurrency semantics.** `AddSymbol`/`RemoveSymbol` use
   `sync.RWMutex` + cache invalidation; sentinel errors map to specific HTTP status codes. Easy to
   break silently when the service is refactored.

4. **SQLite persistence is being added in parallel PRs** (tasks and stocks). The new DB-backed
   service methods will need tests more than the in-memory versions they replace.

5. **The `wmoToCondition` mapping and `aqiCategory` are pure functions** — the simplest tests in
   the codebase to write and the most stable to maintain.

---

## Options Considered

### Option A: E2E Only (e.g. Playwright)

**Description**: Browser-driven tests that spin up the full stack (Go backend + Vite frontend or
built bundle) and interact through the UI.

**Pros**
- Highest confidence: tests the exact path a user takes
- Catches integration issues across the nginx proxy, auth cookies, and TanStack Query refetch logic
- Framework is well-established and TypeScript-native

**Cons**
- Requires a running backend with valid API keys (or extensive mocking of external APIs)
- Slow feedback loop: a full suite can take minutes
- Dashboard content is time-sensitive (weather, calendar, stocks) — assertions on data are fragile
- Session/auth cookie setup adds significant test-environment complexity
- External API rate limits (GNews 1 req/s, Finnhub) make real calls impractical in CI
- Enormous overkill for a single-developer personal tool with no SLA

**Fit for this project**: Poor. The project is personal/LAN-use with no team or CI requirements
that would justify this overhead. Most bugs in this codebase are in pure logic and data transforms,
not in click-flows.

---

### Option B: Component Testing (e.g. Vitest + Testing Library)

**Description**: Mount individual React components in a simulated DOM (jsdom), render them with
fake props or a mocked QueryClient, and assert on output.

**Pros**
- Fast (< 1s per test typically)
- Catches rendering regressions on cards
- Good for verifying conditional states (loading, error, unavailable)

**Cons**
- Most cards are purely presentational — they receive typed props and render them. Testing "does
  `<WeatherCard temp={72} />` render '72°F'" is a tautology.
- The interesting frontend logic lives in *hooks*, not in JSX render trees
- Requires jsdom, `@testing-library/react`, and careful QueryClient setup

**Fit for this project**: Partial. Component tests are the right tool for *hooks* testing, not for
card rendering. Cards are so thin that component-level tests add noise without signal.

---

### Option C: Contract Testing (e.g. Pact)

**Description**: Consumer-driven contract tests where the frontend defines expected API shapes,
and the backend verifies it can satisfy them.

**Pros**
- Catches API shape mismatches between frontend `types/dashboard.ts` and Go model structs
- Good for teams where frontend and backend are developed independently

**Cons**
- The project already has TypeScript types (`frontend/src/types/dashboard.ts`) hand-maintained
  to match Go structs, plus a shared `GET /api/dashboard` aggregator — a drift is caught at
  compile time on the TS side and at JSON unmarshal time on the Go side
- Pact adds a Pact Broker or local `.pact` file management overhead
- Single developer — there is no team coordination problem to solve here
- `go build ./...` + `npm run build` already provide a lightweight contract check

**Fit for this project**: Not recommended. The type-safety story is already reasonably strong
without a separate contract testing layer.

---

### Option D: Go Unit Tests Only

**Description**: Standard `go test` targeting service-layer pure functions and handler behaviour
via `net/http/httptest`.

**Pros**
- Zero new dependencies — `testing` and `net/http/httptest` are in the standard library
- Fast execution (typically < 1s for the entire backend suite)
- Direct: test the exact functions most likely to have bugs
- `go test -race ./...` gives concurrency safety coverage for free, exercising the watchlist mutex
  and cache `sync.Map`

**Cons**
- Does not cover the frontend at all
- Does not test the full HTTP routing stack (middleware, CORS, auth)

**Fit for this project**: Good for backend. Insufficient alone because the frontend has genuine
complexity in `useTasks` that pure backend tests cannot catch.

---

### Option E: Hybrid/Layered Approach (Recommended)

**Description**: Backend pure-logic unit tests + httptest handler integration tests, plus frontend
hook tests with Vitest + React Testing Library. No E2E, no contract testing.

**Pros**
- Tests the highest-risk code in each layer directly, with minimal overhead
- Backend tests: zero new dependencies
- Frontend tests: Vitest is already compatible with the Vite 7 build toolchain; the delta is
  three devDependencies (`vitest`, `@testing-library/react`, `@testing-library/user-event`,
  `@vitejs/plugin-react` is already installed) and a one-line `vite.config.ts` addition
- `go test -race` validates watchlist and cache concurrency
- Fast: entire suite should run in under 10 seconds

**Cons**
- Two test frameworks to maintain (Go + Vitest)
- Hook tests require wrapping components in `QueryClientProvider` — minor boilerplate

**Fit for this project**: Best. Covers the actual risk surface without infrastructure overhead.

---

## Recommendation

**Adopt Option E: Hybrid/Layered testing.**

The codebase has two clusters of real complexity:

1. **Backend pure logic** — WMO lookup, AQI categorisation, calendar `filterToday` (all-day
   detection, TZID handling, sort), `formatDuration`, `relativeTime`, and watchlist
   add/remove sentinel error handling. These are pure or near-pure functions testable with
   `go test` and no additional libraries.

2. **Frontend optimistic mutations** — `useTasks` has three mutations each with a full
   cancel/snapshot/rollback pattern. This is the most complex frontend code and the most likely
   to regress silently during refactors (e.g. the in-progress SQLite persistence work).

Everything else — card rendering, the dashboard fan-out handler, the news sequential fetcher —
is either too simple to fail in interesting ways or too tightly coupled to external APIs to test
cheaply.

**Do not add E2E tests now.** The cost-benefit ratio is wrong for a single-developer personal
tool. Re-evaluate if the project gains a second developer or a CI pipeline with staging.

---

## Implementation Plan

### Phase 1: Backend unit tests (no new dependencies)

Target files: `service/weather_test.go`, `service/calendar_test.go`, `service/news_test.go`,
`service/stocks_test.go`, `service/cache_test.go`

**Pattern — pure function tests:**

```go
// backend/internal/service/weather_test.go
package service

import "testing"

func TestWmoToCondition(t *testing.T) {
    cases := []struct {
        code      int
        condition string
        icon      string
    }{
        {0, "Clear Sky", "☀️"},
        {63, "Rain", "🌧️"},
        {999, "Unknown", "🌡️"}, // unknown code
    }
    for _, tc := range cases {
        cond, icon := wmoToCondition(tc.code)
        if cond != tc.condition || icon != tc.icon {
            t.Errorf("wmoToCondition(%d) = %q %q, want %q %q",
                tc.code, cond, icon, tc.condition, tc.icon)
        }
    }
}

func TestAqiCategory(t *testing.T) {
    cases := []struct{ aqi int; want string }{
        {0, "Good"}, {50, "Good"},
        {51, "Moderate"}, {100, "Moderate"},
        {101, "Unhealthy for Sensitive"},
        {201, "Unhealthy"},
        {301, "Hazardous"},
    }
    for _, tc := range cases {
        if got := aqiCategory(tc.aqi); got != tc.want {
            t.Errorf("aqiCategory(%d) = %q, want %q", tc.aqi, got, tc.want)
        }
    }
}
```

**Pattern — calendar filterToday with table-driven fixtures:**

```go
// backend/internal/service/calendar_test.go
package service

import (
    "strings"
    "testing"
    "time"

    ics "github.com/arran4/golang-ical"
)

func TestFormatDuration(t *testing.T) {
    cases := []struct {
        d    time.Duration
        want string
    }{
        {45 * time.Minute, "45m"},
        {2 * time.Hour, "2h"},
        {90 * time.Minute, "1h 30m"},
        {-1 * time.Minute, "?"},
    }
    for _, tc := range cases {
        if got := formatDuration(tc.d); got != tc.want {
            t.Errorf("formatDuration(%v) = %q, want %q", tc.d, got, tc.want)
        }
    }
}

func TestFilterTodayAllDay(t *testing.T) {
    // Construct a minimal ICS with an all-day event for today
    today := time.Now().Format("20060102")
    icsBody := "BEGIN:VCALENDAR\nBEGIN:VEVENT\nSUMMARY:Team Standup\n" +
        "DTSTART;VALUE=DATE:" + today + "\nEND:VEVENT\nEND:VCALENDAR\n"

    cal, err := ics.ParseCalendar(strings.NewReader(icsBody))
    if err != nil {
        t.Fatal(err)
    }
    svc := &CalendarService{loc: time.Local}
    events := svc.filterToday(cal)
    if len(events) != 1 {
        t.Fatalf("expected 1 event, got %d", len(events))
    }
    if events[0].Time != "All Day" {
        t.Errorf("expected Time='All Day', got %q", events[0].Time)
    }
}
```

**Pattern — stocks watchlist concurrency with `-race`:**

```go
// backend/internal/service/stocks_test.go
package service

import (
    "net/http"
    "testing"
)

func TestAddRemoveSymbol(t *testing.T) {
    svc := NewStocksService(&http.Client{}, "", nil) // nil cache is fine for watchlist ops
    // Note: cache.Delete is a no-op when cache is nil — adjust if needed

    if err := svc.AddSymbol("tsla"); err != nil {
        t.Fatal(err)
    }
    syms := svc.GetSymbols()
    found := false
    for _, s := range syms {
        if s == "TSLA" { found = true }
    }
    if !found {
        t.Error("TSLA not in watchlist after add")
    }

    if err := svc.AddSymbol("TSLA"); err != ErrSymbolExists {
        t.Errorf("expected ErrSymbolExists, got %v", err)
    }

    if err := svc.RemoveSymbol("TSLA"); err != nil {
        t.Fatal(err)
    }
    if err := svc.RemoveSymbol("TSLA"); err != ErrSymbolNotFound {
        t.Errorf("expected ErrSymbolNotFound, got %v", err)
    }
}
```

### Phase 2: Backend handler tests via httptest

Target: `handler/stocks_test.go`, `handler/tasks_test.go`

**Pattern — handler with httptest:**

```go
// backend/internal/handler/stocks_test.go
package handler_test

import (
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "strings"
    "testing"

    "github.com/go-chi/chi/v5"

    "github.com/daily-dashboard/backend/internal/handler"
    "github.com/daily-dashboard/backend/internal/service"
)

func newStocksRouter(t *testing.T) *chi.Mux {
    t.Helper()
    svc := service.NewStocksService(&http.Client{}, "fake-key", nil)
    h := handler.NewStocksHandler(svc)
    r := chi.NewRouter()
    h.AddRoutes(r)
    return r
}

func TestGetWatchlist(t *testing.T) {
    r := newStocksRouter(t)
    req := httptest.NewRequest(http.MethodGet, "/api/stocks/watchlist", nil)
    w := httptest.NewRecorder()
    r.ServeHTTP(w, req)

    if w.Code != http.StatusOK {
        t.Fatalf("expected 200, got %d", w.Code)
    }
    var body struct{ Symbols []string `json:"symbols"` }
    if err := json.NewDecoder(w.Body).Decode(&body); err != nil {
        t.Fatal(err)
    }
    if len(body.Symbols) == 0 {
        t.Error("expected non-empty default watchlist")
    }
}

func TestAddSymbol_Conflict(t *testing.T) {
    r := newStocksRouter(t)
    body := strings.NewReader(`{"symbol":"AAPL"}`) // AAPL is in default list
    req := httptest.NewRequest(http.MethodPost, "/api/stocks/watchlist", body)
    req.Header.Set("Content-Type", "application/json")
    w := httptest.NewRecorder()
    r.ServeHTTP(w, req)

    if w.Code != http.StatusConflict {
        t.Errorf("expected 409, got %d", w.Code)
    }
}
```

### Phase 3: Frontend hook tests with Vitest

**New devDependencies to add to `frontend/package.json`:**

```json
"vitest": "^2.x",
"@testing-library/react": "^16.x",
"@testing-library/user-event": "^14.x",
"jsdom": "^25.x",
"@vitejs/plugin-react": "already installed"
```

**`vite.config.ts` addition:**

```ts
/// <reference types="vitest" />
export default defineConfig({
  // ... existing config ...
  test: {
    environment: 'jsdom',
    globals: true,
    setupFiles: ['./src/test/setup.ts'],
  },
})
```

**`package.json` script:**

```json
"test": "vitest run",
"test:watch": "vitest"
```

**Pattern — useTasks optimistic update:**

```ts
// frontend/src/hooks/useTasks.test.ts
import { renderHook, act } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { useTasks } from './useTasks';
import { vi } from 'vitest';
import * as client from '../api/client';

function wrapper(queryClient: QueryClient) {
  return ({ children }: { children: React.ReactNode }) => (
    <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
  );
}

it('optimistically toggles a task before server response', async () => {
  const qc = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  const initialTask = { id: '1', text: 'Test', done: false, priority: 'medium' };
  qc.setQueryData(['dashboard'], { tasks: [initialTask], weather: null, stocks: [], calendar: [] });

  vi.spyOn(client, 'toggleTask').mockResolvedValue({ ...initialTask, done: true });

  const { result } = renderHook(() => useTasks(), { wrapper: wrapper(qc) });

  act(() => {
    result.current.toggle.mutate({ id: '1', done: true });
  });

  // Optimistic update should be visible immediately
  const optimistic = qc.getQueryData<any>(['dashboard']);
  expect(optimistic?.tasks[0].done).toBe(true);
});
```

**Pattern — useSymbolSearch debounce:**

```ts
// frontend/src/hooks/useSymbolSearch.test.ts
import { renderHook, act } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { useSymbolSearch } from './useSymbolSearch';
import { vi } from 'vitest';
import * as client from '../api/client';

it('does not fire query until debounce settles', async () => {
  vi.useFakeTimers();
  const spy = vi.spyOn(client, 'searchSymbols').mockResolvedValue({ results: [] });
  const qc = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  const W = ({ children }: any) => <QueryClientProvider client={qc}>{children}</QueryClientProvider>;

  const { result } = renderHook(() => useSymbolSearch(), { wrapper: W });

  act(() => result.current.setSearchQuery('TS'));
  expect(spy).not.toHaveBeenCalled(); // before debounce fires

  act(() => vi.advanceTimersByTime(400));
  await act(async () => {}); // flush promises

  expect(spy).toHaveBeenCalledWith('TS');
  vi.useRealTimers();
});
```

### Phase 4: Run in CI (optional, future)

If a GitHub Actions pipeline is ever added:

```yaml
- name: Backend tests
  run: cd backend && go test -race ./...

- name: Frontend tests
  run: cd frontend && npm test
```

---

## Testing Guidelines

### When to write a test

| Scenario | Write a test? |
|---|---|
| Pure function with branching logic (wmoToCondition, aqiCategory, formatDuration, relativeTime) | Yes — table-driven unit test |
| Data transformation with edge cases (filterToday all-day, TZID, cancelled events) | Yes — fixture-driven unit test |
| Mutex + sentinel error logic (AddSymbol, RemoveSymbol) | Yes — unit test + run with `-race` |
| HTTP handler status code mapping (409 conflict, 404 not found) | Yes — httptest integration test |
| Optimistic update / rollback mutation hook | Yes — Vitest hook test with mocked API |
| Debounced query hook | Yes — Vitest hook test with fake timers |
| Presentational card rendering (WeatherCard, CalendarCard, etc.) | No — too thin, changes too often |
| External API integration (Finnhub, Open-Meteo, GNews) | No — use real API or skip; mocking these adds fragility |
| In-memory CRUD that is being replaced by DB-backed implementation | No — test the final version |
| Auth/session middleware | Defer — depends on Google OAuth callback; integration test burden is high |

### What NOT to test

- **Getters and setters with no logic** — `GetSymbols()` returning a copy is obvious
- **JSON serialisation** — Go's `encoding/json` is not your code
- **Framework wiring** — chi routing, TanStack Query's internal fetch lifecycle
- **Dashboard fan-out handler** — it always returns nil; the individual services are what matter
- **Cards that receive props and render them** — unless a non-trivial conditional is introduced

### Coverage philosophy

No coverage threshold. Coverage metrics on a codebase this size create incentives to write
trivial tests. The goal is: **if a future refactor silently breaks the most risky logic, a test
fails**. That means:

- All branches of `filterToday` are covered (all-day, timed, cancelled, wrong-day)
- All branches of `aqiCategory` are covered
- Both sentinel errors from `AddSymbol`/`RemoveSymbol` are covered
- The optimistic-rollback path in `useTasks` is covered
- The debounce guard in `useSymbolSearch` is covered

A suite of ~25-35 targeted tests achieves this. Chasing 80% line coverage would require mocking
every external HTTP call, which inverts the cost-benefit ratio.

---

## Open Questions

The following decisions must be made before implementation begins:

1. **CacheService nil-safety in tests**: `StocksService.AddSymbol` calls `s.cache.Delete("stocks")`
   — if `cache` is nil, this panics. Should `CacheService` be an interface so tests can pass a
   no-op, or should `cache.Delete` be nil-guarded in the service? **Decision needed before writing
   stocks tests.**

2. **Vitest version pin**: Vitest 2.x is the current stable release. The project uses Vite 7 —
   confirm Vitest 2.x compatibility before adding. (Vitest 3 was in beta as of early 2025.)
   **Check before installing.**

3. **Test file location convention**: Backend tests follow Go convention (same package or
   `package foo_test` for black-box). For frontend, should test files live alongside source
   (`hooks/useTasks.test.ts`) or in a `__tests__/` directory? **Recommend co-location
   (alongside source) — simpler to find and import.**

4. **SQLite-backed service tests**: The in-progress PRs (tasks and stocks) are replacing
   in-memory services with SQLite. Should we wait for those PRs to land before writing service
   tests, or write tests now for the in-memory version and rewrite them? **Recommend: wait for
   the SQLite PRs to land, then write tests against the new DB-backed implementation using an
   in-memory SQLite instance (`:memory:` DSN) — this is the most useful version to have.**

5. **Auth handler testing**: The auth service depends on Google OAuth callbacks and a real DB.
   Testing it meaningfully requires either a mock OAuth server or integration test fixtures.
   **Out of scope for this RFC — file a separate bd issue if desired.**
