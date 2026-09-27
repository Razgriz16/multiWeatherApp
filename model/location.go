package model

// Country is filled by geocoding but not stored: locations has no such column.
// ID and CreatedAt are zero until the row exists in the database.
type Location struct {
	ID        int64
	Name      string
	Country   string
	Latitude  float64
	Longitude float64
	Timezone  string
	Active    bool
	Notify    bool
	CreatedAt string
}
