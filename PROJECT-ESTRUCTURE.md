# Project Structure

Folder layout for the architecture locked in [PROJECT-RESEARCH.md](PROJECT-RESEARCH.md)
and [SCHEMA-ANALYSIS.md](SCHEMA-ANALYSIS.md): one Go binary, three
subcommands (`ingest`, `serve`, `notify`), one SQLite file, no
`node_modules`. This file is the map from that architecture to actual
directories — read the other two for *why*, this one for *where*.

## Tree

```
multiWeatherApp/
├── cmd/
│   └── multiweather/
│       ├── main.go            # arg dispatch: ingest | serve | notify
│       ├── ingest.go           # wiring only: config → fetch → aggregate → store
│       ├── serve.go            # wiring only: store → serve.Handler → http.ListenAndServe
│       └── notify.go           # Phase 2 stub
├── internal/
│   ├── model/                   # shared structs mirroring the schema — no logic
│   ├── config/                   # loads the location config file
│   ├── fetch/
│   │   └── openmeteo.go         # Open-Meteo client: fetch + normalize (§5.1)
│   ├── aggregate/                 # median + threshold aggregation (§5.2)
│   │   ├── median.go              # temperature, humidity, dew point, wind
│   │   ├── precipitation.go       # ensemble-threshold derivation, not median
│   │   └── condition.go           # derived from aggregated numerics
│   ├── store/                      # SQLite persistence only, no business logic (§5.3)
│   │   ├── sqlite.go                # connection + schema init, no migrations (v1)
│   │   ├── locations.go
│   │   ├── readings.go
│   │   └── aggregates.go
│   ├── serve/                        # HTTP handlers, reads store + web.Templates (§5.4)
│   │   └── handlers.go
│   └── notify/                        # Phase 2 stub (§6.1)
├── web/
│   ├── templates.go                    # package web; go:embed templates/*.html
│   ├── templates/
│   │   ├── layout.html                  # Pico.css + Chart.js via CDN <script>/<link>
│   │   ├── index.html
│   │   └── location.html
│   └── static/
│       └── favicon.ico
├── config/
│   └── locations.example.json           # template; real file may be gitignored
├── deploy/
│   ├── multiweather.cron                 # example cron line for `ingest`
│   └── multiweather.service               # example systemd unit for `serve`
├── scripts/
│   └── calibration/                        # optional, off-device only (§6.2)
│       └── calibration.py                   # never imported by Go code
├── data/                                     # SQLite db file at runtime, gitignored
├── PROJECT-RESEARCH.md
├── SCHEMA-ANALYSIS.md
├── PROJECT-ESTRUCTURE.md                      # this file
├── AGENTS.md
├── CLAUDE.md
├── go.mod
└── go.sum
```

## Mapping your four buckets to this layout

| Your term | Where it lives | Why not a top-level folder |
|---|---|---|
| **views** | `web/templates/` (HTML) + `internal/serve/` (handlers) | Go idiom splits *what's rendered* (`web/`) from *what renders it* (`internal/serve/`); combining them into one `views/` would mix markup and Go logic in the same directory |
| **api-fetching** | `internal/fetch/` | `fetch` is the package name; `openmeteo.go` is a file inside it. If a second provider is ever added (§7), it's `internal/fetch/otherprovider.go`, not a new top-level folder |
| **aggregates** | `internal/aggregate/` | This is the one package where median/threshold logic is allowed to exist — CLAUDE.md's "aggregation logic lives in one place only" rule. `internal/store` must never compute anything, only persist |
| **data-analysis** | `scripts/calibration/` | This is the §6.2 calibration study — Python, off-device, optional. Kept out of `internal/` entirely so nothing accidentally imports it into the deploy binary |

## Why `cmd/` + `internal/`, not `views/`, `api-fetching/`, etc. at the root

This is standard Go project layout, not a stylistic preference:

- **`cmd/multiweather/`** holds only `main()` and CLI wiring — parsing
  `os.Args[1]` into `ingest`/`serve`/`notify` and calling into `internal/`.
  No business logic here, so the three subcommands of the *one binary*
  (§3) stay thin and symmetric.
- **`internal/`** is a Go-enforced boundary: nothing outside this module
  can import from it. Appropriate here because this isn't a reusable
  library — it's the guts of one binary.
- **`web/`** sits outside `internal/` on purpose: it's data (`.html`
  templates, a favicon), not Go logic, and it's imported by
  `internal/serve` via a Go package (`web`) that wraps a `go:embed`
  directive — so templates compile into the binary and ship as part of
  the single static executable (§5.2's whole point), not as loose files
  that have to be copied alongside it on deploy.

## Naming conventions

- **Go package names**: short, lowercase, one word, no
  underscores/hyphens/mixedCaps — `fetch`, `aggregate`, `store`, `serve`,
  not `apiFetching` or `data_analysis`.
- **No stutter**: callers write `fetch.OpenMeteo(...)`, not
  `fetch.FetchOpenMeteo(...)`. The package name is part of the call site.
- **One package, one responsibility**: `aggregate` computes, `store`
  persists, `fetch` retrieves, `serve` renders. A function that needs to
  do two of those needs two calls from `cmd/`, not a package that does
  both.
- **Non-Go directories** (`web/`, `config/`, `deploy/`, `scripts/`,
  `data/`) use plain lowercase names since they're not Go packages and
  carry no compiler constraints — matches the filesystem convention Go
  tooling itself expects (`go.mod`, `go.sum`).

## Deliberately not here

- No `views/`, `api-fetching/`, `data-analysis/` as literal top-level
  folder names — they're your mental model for the pipeline, mapped
  above onto Go's actual package boundaries.
- No `migrations/` — v1 assumes a fresh database (§4/SCHEMA-ANALYSIS.md).
- No separate `cmd/ingest/`, `cmd/serve/`, `cmd/notify/` binaries — it's
  one binary, one `cmd/multiweather/`, three subcommands (§3).
- No `node_modules/`, no `package.json`, no JS build step anywhere —
  Pico.css and Chart.js load via CDN `<script>`/`<link>` tags directly in
  `web/templates/layout.html`.
