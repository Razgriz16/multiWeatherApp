package model

// One row per source per hour. Pointer fields are nullable — nil means
// that source had no value for this hour, distinct from a real 0.
type Reading struct {
	ID                       int64
	SourceID                 int64
	LocationID               int64
	ValidTime                string
	IssuedAt                 string
	FetchedAt                string
	Temperature              *float64
	Humidity                 *float64
	DewPoint                 *float64
	PrecipitationMM          *float64
	PrecipitationProbability *float64
	WindSpeed                *float64
	WindDirection            *float64
	ConditionID              *int64
}
