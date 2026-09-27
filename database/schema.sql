CREATE TABLE IF NOT EXISTS sources (
    id       INTEGER PRIMARY KEY,
    provider TEXT NOT NULL,        -- e.g. 'open-meteo'
    name     TEXT NOT NULL UNIQUE  -- e.g. 'ecmwf', 'gfs', 'icon'
);
CREATE TABLE IF NOT EXISTS conditions (
    id   INTEGER PRIMARY KEY,
    code TEXT NOT NULL UNIQUE      -- 'clear', 'rain', 'overcast', 'snow', 'fog', 'thunderstorm'
);
CREATE TABLE IF NOT EXISTS readings (
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
CREATE TABLE IF NOT EXISTS aggregates (
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
CREATE TABLE IF NOT EXISTS locations (
    id         INTEGER PRIMARY KEY,
    name       TEXT NOT NULL,
    latitude   REAL NOT NULL,
    longitude  REAL NOT NULL,
    timezone   TEXT NOT NULL,
    active     INTEGER NOT NULL DEFAULT 1,
    notify     INTEGER NOT NULL DEFAULT 0,
    created_at TEXT NOT NULL,
    UNIQUE (latitude, longitude)
);
