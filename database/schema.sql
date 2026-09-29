CREATE TABLE IF NOT EXISTS sources (
    id        INTEGER PRIMARY KEY,
    provider  TEXT NOT NULL,        -- e.g. 'open-meteo'
    name      TEXT NOT NULL UNIQUE, -- e.g. 'ecmwf', 'gfs', 'icon'
    api_model TEXT NOT NULL UNIQUE  -- provider's own model id, e.g. 'ecmwf_ifs025'.
                                     -- forecastResponse (fetch/types.go) has one
                                     -- hardcoded field per model — adding a row here
                                     -- doesn't make it decode on its own.
);
-- One row per Open-Meteo WMO weather_code — the actual values a model
-- can return, not a collapsed category. How an *aggregate* row picks
-- one of these across 3 sources is still an open question
-- (docs/PROJECT-RESEARCH.md §7); nothing here needs to pre-decide that.
CREATE TABLE IF NOT EXISTS conditions (
    id          INTEGER PRIMARY KEY,
    code        INTEGER NOT NULL UNIQUE, -- Open-Meteo's WMO weather_code
    description TEXT NOT NULL            -- Open-Meteo's own text, e.g. 'Heavy rain'
);
INSERT OR IGNORE INTO sources (provider, name, api_model) VALUES
    ('open-meteo', 'ecmwf', 'ecmwf_ifs025'),
    ('open-meteo', 'gfs',   'gfs_seamless'),
    ('open-meteo', 'icon',  'icon_seamless');

INSERT OR IGNORE INTO conditions (code, description) VALUES
    (0,  'Clear sky'),
    (1,  'Mainly clear'),
    (2,  'Partly cloudy'),
    (3,  'Overcast'),
    (45, 'Fog'),
    (48, 'Depositing rime fog'),
    (51, 'Light drizzle'),
    (53, 'Moderate drizzle'),
    (55, 'Dense drizzle'),
    (56, 'Light freezing drizzle'),
    (57, 'Dense freezing drizzle'),
    (61, 'Slight rain'),
    (63, 'Moderate rain'),
    (65, 'Heavy rain'),
    (66, 'Light freezing rain'),
    (67, 'Heavy freezing rain'),
    (71, 'Slight snowfall'),
    (73, 'Moderate snowfall'),
    (75, 'Heavy snowfall'),
    (77, 'Snow grains'),
    (80, 'Slight rain showers'),
    (81, 'Moderate rain showers'),
    (82, 'Violent rain showers'),
    (85, 'Slight snow showers'),
    (86, 'Heavy snow showers'),
    (95, 'Thunderstorm'),
    (96, 'Thunderstorm, slight hail'),
    (97, 'Heavy thunderstorm'),
    (99, 'Thunderstorm, heavy hail');
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
