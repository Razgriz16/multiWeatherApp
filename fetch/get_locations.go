package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"

	"multiweather/model"
)

type geonamesResponse struct {
	Status struct {
		Message string `json:"message"`
	} `json:"status"`
	Geonames []struct {
		Name        string  `json:"name"`
		CountryCode string  `json:"countryCode"`
		Lat         float64 `json:"lat,string"`
		Lng         float64 `json:"lng,string"`
		Timezone    struct {
			ID string `json:"timeZoneId"`
		} `json:"timezone"`
	} `json:"geonames"`
}

func geocode(name, country string) (model.Location, error) {
	username := os.Getenv("GEONAMES_USERNAME")
	if username == "" {
		return model.Location{}, errors.New("GEONAMES_USERNAME is not set")
	}

	q := url.Values{}
	q.Set("q", name)
	q.Set("featureClass", "P")
	q.Set("maxRows", "5")
	q.Set("style", "FULL")
	q.Set("username", username)
	if country != "" {
		q.Set("country", country)
	}

	resp, err := http.Get("http://api.geonames.org/searchJSON?" + q.Encode())
	if err != nil {
		return model.Location{}, err
	}
	defer resp.Body.Close()

	var result geonamesResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return model.Location{}, fmt.Errorf("geonames returned status %d: %w", resp.StatusCode, err)
	}
	if result.Status.Message != "" {
		return model.Location{}, fmt.Errorf("geonames error: %s", result.Status.Message)
	}
	if resp.StatusCode != http.StatusOK {
		return model.Location{}, fmt.Errorf("geonames returned status %d", resp.StatusCode)
	}
	if len(result.Geonames) == 0 {
		return model.Location{}, fmt.Errorf("no match for %q", name)
	}

	g := result.Geonames[0]
	if g.Timezone.ID == "" {
		return model.Location{}, fmt.Errorf("no timezone returned for %q", name)
	}
	return model.Location{
		Name:      g.Name,
		Country:   g.CountryCode,
		Latitude:  g.Lat,
		Longitude: g.Lng,
		Timezone:  g.Timezone.ID,
	}, nil
}
