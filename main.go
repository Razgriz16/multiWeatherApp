package main

import (
	"fmt"
	"multiweather/database"
	"multiweather/fetch"
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

	for _, loc := range locations {
		fetch.FetchForecast(loc, fetch.Models)
	}
}
