// Package web renders the read-only HTML view over readings/aggregates.
// It never fetches or aggregates — purely reads what ingest already wrote
// (docs/PROJECT-RESEARCH.md §3).
package web

import (
	"database/sql"
	_ "embed"
	"fmt"
	"html/template"
	"net/http"

	"multiweather/database"
	"multiweather/model"
)

//go:embed templates/layout.html
var layoutRaw string

//go:embed templates/index.html
var indexRaw string

//go:embed static/style.css
var styleRaw string

// Template text comes from the go:embed vars above — baked into the
// binary at build time, not read from disk at startup.
var indexTemplate = parsePage(indexRaw)

func parsePage(pageRaw string) *template.Template {
	t := template.Must(template.New("layout").Parse(layoutRaw))
	return template.Must(t.Parse(pageRaw))
}

// NewServer builds the HTTP handler for the serve subcommand.
func NewServer(db *sql.DB) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /", indexHandler(db))
	mux.HandleFunc("GET /static/style.css", staticCSSHandler)
	return mux
}

func staticCSSHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/css")
	fmt.Fprint(w, styleRaw)
}

// locationView is one location's card dashboard: Hero holds the day(s)
// shown in detail (today+tomorrow, or a single selected day), Strip
// holds every upcoming day as a small summary card.
type locationView struct {
	Location model.Location
	Hero     []dayCard
	Strip    []dayCard
}

func indexHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		locations, err := database.GetActiveLocations(db)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		selectedDay := r.URL.Query().Get("day")

		views := make([]locationView, len(locations))
		for i, loc := range locations {
			hours, err := database.LatestForecast(db, loc.ID)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			days := rollupDaily(hours)
			views[i] = locationView{
				Location: loc,
				Hero:     selectHero(days, selectedDay),
				Strip:    markSelected(days, selectedDay),
			}
		}

		data := struct {
			Title       string
			Locations   []locationView
			SelectedDay string
		}{"multiWeatherApp", views, selectedDay}

		if err := indexTemplate.ExecuteTemplate(w, "layout", data); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}
