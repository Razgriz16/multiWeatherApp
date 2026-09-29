package main

import (
	"fmt"
	"multiweather/database"
	"multiweather/fetch"
	"time"
)

func main() {
	if err := loadEnv(".env"); err != nil {
		fmt.Println("env failed:", err)
		return
	}

	db, err := database.Open("data/weather.db")
	if err != nil {
		fmt.Println("database failed:", err)
		return
	}
	defer db.Close()

	locations, err := database.GetActiveLocations(db)
	if err != nil {
		fmt.Println("load locations failed:", err)
		return
	}

	sourceIDs, err := database.SourceIDs(db)
	if err != nil {
		fmt.Println("load sources failed:", err)
		return
	}

	models, err := database.APIModels(db)
	if err != nil {
		fmt.Println("load models failed:", err)
		return
	}

	conditionIDs, err := database.ConditionIDs(db)
	if err != nil {
		fmt.Println("load conditions failed:", err)
		return
	}

	issuedAt := time.Now().UTC().Truncate(24 * time.Hour).Format(time.RFC3339)

	for _, loc := range locations {
		forecast, err := fetch.FetchForecast(loc, models)
		if err != nil {
			fmt.Println(loc.Name, "fetch failed:", err)
			continue
		}
		fetchedAt := time.Now().UTC().Format(time.RFC3339)

		readings, err := fetch.NormalizeReadings(loc, sourceIDs, conditionIDs, forecast, issuedAt, fetchedAt)
		if err != nil {
			fmt.Println(loc.Name, "normalize failed:", err)
			continue
		}

		inserted, err := database.InsertReadings(db, readings)
		if err != nil {
			fmt.Println(loc.Name, "insert failed:", err)
			continue
		}
		fmt.Println(loc.Name, "inserted", inserted, "of", len(readings), "readings")
	}
}
