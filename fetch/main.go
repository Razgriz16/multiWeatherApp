package main

import (
	"fmt"
	"multiweather/database"
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

	p, err := geocode("Talca", "CL")
	if err != nil {
		fmt.Println("geocode failed:", err)
		return
	}

	inserted, err := database.AddLocation(db, p)
	if err != nil {
		fmt.Println("insert failed:", err)
		return
	}
	if !inserted {
		fmt.Println(p.Name, "is already tracked")
		return
	}
	fmt.Println("added", p.Name)
}
