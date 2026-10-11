package model

// One row per (location, valid_time, issued_at) — the cross-source
// consensus for one forecast hour from one ingest run. Pointer fields are
// nullable, same convention as Reading.
type Aggregate struct {
	ID                       int64
	LocationID               int64
	ValidTime                string
	IssuedAt                 string
	Temperature              *float64
	Humidity                 *float64
	DewPoint                 *float64
	PrecipitationMM          *float64
	PrecipitationProbability *float64
	ConditionID              *int64
	SourceCount              int64
	ComputedAt               string
}
