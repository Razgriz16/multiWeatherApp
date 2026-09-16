# Schema Analysis

Column-by-column detail and the reasoning behind every table, constraint,
and column in the multiWeatherApp database. This is the expansion of
[PROJECT-RESEARCH.md](PROJECT-RESEARCH.md) §4 — read that first for how
the schema fits the overall architecture (§3) and technology choices
(§5); this file exists so §4 can stay a summary instead of a wall of SQL.

**v1 assumes a fresh database** — no schema versioning, no migration
path. The schema below is created from scratch; a migration system is a
problem for if/when it's ever needed, not designed for now.

Five tables: two lookups (`sources`, `conditions`) so repeating text
doesn't get duplicated per row, one dimension table (`locations`), one
fact table (`readings`) for raw per-source hourly data, one derived table
(`aggregates`) for the cross-source median.

## `sources`

```sql
CREATE TABLE sources (
    id       INTEGER PRIMARY KEY,
    provider TEXT NOT NULL,        -- e.g. 'open-meteo'
    name     TEXT NOT NULL UNIQUE  -- e.g. 'ecmwf', 'gfs', 'icon'
);
```

- v1's source set is fixed to 3 models (ECMWF, NOAA GFS, DWD ICON) for
  every active location — a lookup table anyway, so widening the set
  later doesn't touch the schema.
- Would also be where a future per-source accuracy weight (Phase 2
  accuracy-weighted blending) lives as an added column — not included
  now, kept out until that's actually built.

## `conditions`

```sql
CREATE TABLE conditions (
    id   INTEGER PRIMARY KEY,
    code TEXT NOT NULL UNIQUE      -- 'clear', 'rain', 'overcast', 'snow', 'fog', 'thunderstorm'
);
```

Wind speed/direction are plain numeric columns on `readings`/`aggregates`,
not a lookup table — they aren't categorical, only `condition` is.

## `locations`

```sql
CREATE TABLE locations (
    id         INTEGER PRIMARY KEY,
    name       TEXT NOT NULL,
    latitude   REAL NOT NULL,                  -- rounded to 4 decimals before insert
    longitude  REAL NOT NULL,                  -- rounded to 4 decimals before insert
    timezone   TEXT NOT NULL,                  -- IANA name, e.g. 'America/Santiago'
    active     INTEGER NOT NULL DEFAULT 1,      -- ingest iterates over these
    notify     INTEGER NOT NULL DEFAULT 0,      -- Phase 2 notify reads these
    created_at TEXT NOT NULL,
    UNIQUE (latitude, longitude)                -- compared post-rounding
);
```

Replaces inline lat/lon on `readings`/`aggregates`. The join cost was
never the real issue — the original `UNIQUE (latitude, longitude,
valid_date)` compared REAL floats for equality, so a config writing
-33.45 vs -33.450000001 would silently produce duplicate aggregate rows.

- **Append-only.** Added, never edited. Mutating a row's coordinates
  would retroactively relabel every historical reading pointing at it,
  silently mixing two climates in one series. Travelling to a new place
  creates a new row; returning reuses the existing one and the series
  continues.
- **Round coordinates to 4 decimals before insert**, and apply the
  UNIQUE constraint to the rounded values — otherwise this reintroduces
  the float-equality bug the table exists to solve. 4 decimals is ~11 m,
  far finer than any global model's grid, which is kilometres wide.
  Consider also storing the grid-cell coordinates Open-Meteo actually
  returns, alongside the requested ones, to make explicit which point is
  really being measured.
- **`timezone` is a per-location IANA name**, never a fixed offset —
  Chile shifts twice a year, and a hardcoded offset would displace every
  `valid_time` for half the year. The local-calendar-date rule (below)
  is per-location because of this, not a global constant.
- **`active` and `notify` are separate flags.** The ingest job iterates
  over all active locations; the Phase 2 notify job reads only those
  with `notify` set. A location can be ingested without being notified —
  e.g. activating a destination a week before travelling so it arrives
  with accumulated history rather than an empty table. `notify` is added
  now, unused in v1, because it's one column and avoids a later
  migration — which v1 otherwise has none of.
- Practical ceiling of 2–3 active locations at a time is hygiene, not a
  technical limit.

## `readings`

```sql
CREATE TABLE readings (
    id                        INTEGER PRIMARY KEY,
    source_id                 INTEGER NOT NULL REFERENCES sources(id),
    location_id               INTEGER NOT NULL REFERENCES locations(id),
    valid_time                TEXT NOT NULL,  -- ISO-8601 timestamp this forecast is FOR
    init_time                 TEXT,           -- nullable: not exposed by the Forecast API in v1, see note
    issued_at                 TEXT NOT NULL,  -- cron run timestamp for this ingest cycle
    fetched_at                TEXT NOT NULL,  -- ISO-8601 timestamp of the API call (informational)
    temperature               REAL,  -- critical
    humidity                  REAL,  -- critical
    dew_point                 REAL,
    precipitation_mm          REAL,
    precipitation_probability REAL,
    wind_speed                REAL,
    wind_direction            REAL,
    condition_id              INTEGER REFERENCES conditions(id),
    UNIQUE (source_id, location_id, valid_time, issued_at)
);
```

- **Hourly, not daily.** `valid_time` stores a full ISO-8601 timestamp
  (renamed from `valid_date`). The old schema stored one `temperature`
  column per day without ever defining whether that was max, min or
  mean, and daily mean relative humidity is meaningless given its strong
  diurnal cycle. Open-Meteo returns hourly data — store hourly, derive
  daily rollups on read.
- **Two separate keys:**
  - Primary key `id INTEGER PRIMARY KEY` — aliases the SQLite rowid,
    costs nothing, gives a short stable handle.
  - Unique constraint (business identity): `UNIQUE (source_id,
    location_id, valid_time, issued_at)`. The old `UNIQUE (source_id,
    valid_date, fetched_at)` was wrong twice over — it omitted
    `location_id`, mandatory now that multiple locations exist, and it
    keyed on download time rather than forecast identity.
- **Why model-run identity matters — it's measurement accuracy, not just
  deduplication.** If cron runs at 08:00 but the model initialised at
  00:00, the forecast is already 8 hours old on download, so a lead time
  computed from `fetched_at` understates it systematically. The three
  models publish with different latencies (lagged ensemble, below), so
  comparing their skill "at 24 hours" without per-model init times would
  compare different actual lead times with a different bias per model —
  exactly what Phase 2 measures. Having a stable unique constraint also
  makes ingestion idempotent: `INSERT ... ON CONFLICT DO NOTHING` needs
  one to collide against.
  - **Checked 2026-09-15: Open-Meteo's standard Forecast API response
    does not expose `init_time`.** The documented JSON return fields are
    `latitude`, `longitude`, `elevation`, `generationtime_ms`,
    `utc_offset_seconds`, `timezone`, `timezone_abbreviation`, `hourly`,
    `daily` and their `_units` siblings — no run-initialisation
    timestamp. Open-Meteo does publish `last_run_initialisation_time`
    and related fields, but on a separate model-status/metadata endpoint,
    not embedded per reading in the query this project actually calls.
    **v1 therefore keys `readings` on `issued_at`** (the cron run stamp)
    instead of `init_time` — a retry within the same cycle still
    collides. `init_time` stays as a nullable column and `fetched_at` as
    an informational one. The lead-time approximation this causes, and
    its per-model bias, is a known limitation — tracked in PROJECT-RESEARCH.md §7.
    Correlating the model-status endpoint's data back onto readings is a
    possible future refinement, not attempted in v1.
- **Lagged ensemble.** Aggregation does not require model synchrony.
  Global models run on the same synchronised 00/06/12/18 UTC cycle; what
  differs between them is publication latency, not schedule (open-data
  ECMWF dissemination in particular carries a further delay versus its
  real-time feed on the order of a couple of hours — exact current
  per-model figures should be checked against Open-Meteo's live
  model-status page before relying on a number; see §7). Aggregation is
  keyed on `valid_time`, not on freshness: three models describing the
  same future hour are comparable regardless of which run each came
  from. Combining runs of differing initialisation times is a standard
  meteorological technique — a lagged ensemble — not a workaround. The
  operating rule: for each `valid_time`, take the most recent available
  run of each source, i.e. `MAX(init_time)` grouped by source.

## `aggregates`

```sql
CREATE TABLE aggregates (
    id                        INTEGER PRIMARY KEY,
    location_id               INTEGER NOT NULL REFERENCES locations(id),
    valid_time                TEXT NOT NULL,
    issued_at                 TEXT NOT NULL,  -- cron run timestamp
    lead_hours                INTEGER GENERATED ALWAYS AS
                                   (CAST((julianday(valid_time) - julianday(issued_at)) * 24 AS INTEGER))
                                   VIRTUAL,   -- approximate, see note
    temperature               REAL,
    humidity                  REAL,
    dew_point                 REAL,
    precipitation_mm          REAL,
    precipitation_probability REAL,
    condition_id              INTEGER REFERENCES conditions(id),
    source_count              INTEGER NOT NULL,  -- critical fields only, see note
    computed_at               TEXT NOT NULL,
    UNIQUE (location_id, valid_time, issued_at)
);
```

- **Preserves the forecast trajectory** — `UNIQUE (location_id,
  valid_time, issued_at)` instead of collapsing on `valid_time` alone.
  Collapsing would throw away how the forecast for a given hour evolved
  across successive runs, which is the most interesting data the project
  produces (PROJECT-RESEARCH.md §3.1). "Current forecast" is a
  `MAX(issued_at)` query, not a stored flag.
- **`lead_hours` is approximate.** It's a generated column derived from
  `issued_at`, but an aggregate row combines sources whose `init_time`s
  may differ (`readings`, above) — no single exact lead time exists at
  the aggregate level. Exact per-source lead-time analysis belongs at
  the `readings` level, where each row has its own `init_time`.
- **Missing/null source handling (v1 default, easy to tune later)**:
  exclude a source's value from a field's median if that field is null;
  skip writing an aggregate entirely for an hour if fewer than 2 of the 3
  sources have a valid value for a critical field.
- **`source_count` refers to the critical fields (temperature, humidity)
  only**, not one number covering every field — null-handling is applied
  per field, so e.g. temperature can have 3 contributing sources in the
  same row where humidity has 2. Documented here explicitly rather than
  adding a per-field count column, to keep the row shape simple.
- **Precipitation is aggregated differently from the other continuous
  fields.** Median is right for temperature, dew point, pressure, wind —
  continuous variables where an outlier source should be outvoted, not
  averaged in. It's wrong for precipitation amount: 3 models at 0mm and
  2 at 15mm yields a median of 0, erasing the extreme-event signal that
  matters most for precipitation specifically.
  `precipitation_probability` must NOT be aggregated by median either —
  averaging other models' stated uncertainty is conceptually wrong. Both
  are independent of source independence; better sources don't fix this.
  Instead: derive probability from the ensemble itself — the fraction of
  models exceeding a threshold (e.g. 2 of 3 over 1mm). The threshold
  itself is a location-specific constant, not invented here — see the
  calibration study, PROJECT-RESEARCH.md §6.2.
- **`condition_id` is derived, not voted, by default.** Median is
  undefined for a categorical variable. Preferred approach: derive the
  aggregate condition from the already-aggregated numeric fields
  (precipitation_mm, cloud cover, wind) via thresholds — keeps median on
  continuous variables only, and reduces `conditions` to an output
  dictionary rather than something itself aggregated. If majority vote
  across sources is used instead, a tie-break rule is needed now: with 3
  sources a three-way tie is common. Options: most-severe-wins, or write
  null and display "no consensus" — left open in §7 if
  threshold-derivation turns out not to cover every case.

## Deferred: `observations`

An `observations` table (ground truth, for Phase 2 accuracy tracking)
stays out of v1 — deliberately. Adding a new table later is an additive
migration (nothing existing changes); the daily-to-hourly change above
was structural (it invalidated already-ingested data). Only structural
changes were urgent enough to do now.

## Cross-cutting rules

- Dates/timestamps stored as ISO-8601 `TEXT` — SQLite has no native date
  type, and TEXT ISO-8601 sorts correctly.
- **Timestamps (`issued_at`, `fetched_at`, `computed_at`) are UTC.** A
  daily rollup, when needed, uses the *local calendar date at that
  reading's location* — resolved per `locations.timezone`, not a global
  constant, since a UTC-day boundary can land on the wrong local date.
- **Units are metric only** (Celsius, km/h, mm) — Open-Meteo's default;
  never pass its imperial unit parameters.
- **Retention: 5 years, then deleted.** Recalculated for hourly data: 3
  models × 168 hourly rows per 7-day fetch window ≈ 504 `readings`
  rows/day per location, ~184,000 rows/year per location, roughly 18
  MB/year per location. Scales linearly with active locations and cron
  frequency — e.g. 3 active locations on a 6-hourly cron would be
  roughly 220 MB/year, still trivial for SQLite on a phone, but that's an
  illustrative figure, not a fixed one. Implementation: a
  `DELETE WHERE valid_time < 5 years ago` on both `readings` and
  `aggregates`, tacked onto the end of the same `ingest` run — no
  separate scheduled process needed.
