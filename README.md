# multiWeatherApp

A self-hosted weather dashboard that combines several forecast models into one forecast, and shows how much they agree.

## What it solves

Each weather model (ECMWF, NOAA GFS, DWD ICON) predicts something slightly different, and any one of them can be badly off. multiWeatherApp pulls all three hour by hour from [Open-Meteo](https://open-meteo.com), stores them, and shows the **median** across models. That gives you a forecast that holds up when one model is off, and you can still compare each model against it.

Every daily run is kept rather than overwritten. This means you can see whether a future day's forecast is settling or still changing.

## Design thinking

Before writing code I looked at existing aggregators. The full reasoning is in [docs/PROJECT-RESEARCH.md](docs/PROJECT-RESEARCH.md). In short:

- **Median, not mean.** Three separate projects found that a plain average breaks when one source misbehaves. The median handles outliers by construction and needs no past data. Weighting models by their accuracy is planned for v2.
- **Open-Meteo only, no scraping.** One keyless API serves several independent national models, which is already a real multi-model ensemble.
- **One Go binary + one SQLite file.** It has to run on low-power hardware (an old PC or a phone). Go builds a single static binary, and the pure-Go SQLite driver means no cgo. Nothing else needs to be installed on the device.
- **No JS build step.** Pages are rendered on the server with `html/template`, and Pico.css and Chart.js load from a CDN. There is no `node_modules`.
- **Fetching and serving are separate.** `ingest` runs on a schedule and writes. `serve` only reads.

Schema details: [docs/SCHEMA-ANALYSIS.md](docs/SCHEMA-ANALYSIS.md).

## Running it

Requires Go 1.27+.

```sh
go build -o multiweather .
```

**1. Add a location.** There's no command for this yet. The database is created at `data/weather.db` on first run, so run the binary once and then insert a location with any SQLite client:

```sql
INSERT INTO locations (name, latitude, longitude, timezone, created_at)
VALUES ('Santiago', -33.4489, -70.6693, 'America/Santiago', datetime('now'));
```

**2. Fetch forecasts.**

```sh
./multiweather          # ingest: fetch, store, aggregate
```

This is safe to re-run: a second run on the same day inserts nothing new. Schedule it once a day, for example with cron:

```
0 6 * * * cd /path/to/multiWeatherApp && ./multiweather
```

On Windows, use Task Scheduler.

**3. Serve the dashboard.**

```sh
./multiweather serve    # http://localhost:8080
```

**Tests:**

```sh
go test ./...
```

## Credits

Weather data by [Open-Meteo.com](https://open-meteo.com) (CC BY 4.0).
