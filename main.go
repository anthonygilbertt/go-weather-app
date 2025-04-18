// Weather Service Assignment
// Simple HTTP server in Go that returns today's short forecast and temperature classification

package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"time"
)

// PointsResponse represents the response from the NWS points API
type PointsResponse struct {
	Properties struct {
		Forecast string `json:"forecast"`
	} `json:"properties"`
}

// ForecastResponse represents the response from the NWS forecast API
type ForecastResponse struct {
	Properties struct {
		Periods []struct {
			Name            string `json:"name"`
			StartTime       string `json:"startTime"`
			Temperature     int    `json:"temperature"`
			TemperatureUnit string `json:"temperatureUnit"`
			ShortForecast   string `json:"shortForecast"`
			IsDaytime       bool   `json:"isDaytime"`
		} `json:"periods"`
	} `json:"properties"`
}

// WeatherResult is the JSON structure returned by our endpoint
type WeatherResult struct {
	Forecast       string `json:"forecast"`
	Temperature    int    `json:"temperature"`
	Classification string `json:"classification"`
}

func main() {
	http.HandleFunc("/weather", weatherHandler)
	log.Println("Starting server on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func weatherHandler(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	latStr := r.URL.Query().Get("lat")
	lonStr := r.URL.Query().Get("lon")
	if latStr == "" || lonStr == "" {
		http.Error(w, "Missing lat or lon parameter", http.StatusBadRequest)
		return
	}
	lat, err := strconv.ParseFloat(latStr, 64)
	if err != nil {
		http.Error(w, "Invalid latitude", http.StatusBadRequest)
		return
	}
	lon, err := strconv.ParseFloat(lonStr, 64)
	if err != nil {
		http.Error(w, "Invalid longitude", http.StatusBadRequest)
		return
	}

	forecast, temp, classification, err := getForecast(lat, lon)
	if err != nil {
		log.Println("Error fetching forecast:", err)
		http.Error(w, "Failed to fetch forecast", http.StatusInternalServerError)
		return
	}

	result := WeatherResult{
		Forecast:       forecast,
		Temperature:    temp,
		Classification: classification,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func getForecast(lat, lon float64) (string, int, string, error) {
	// Step 1: Call points API
	pointsURL := fmt.Sprintf("https://api.weather.gov/points/%.4f,%.4f", lat, lon)
	pointsReq, _ := http.NewRequest("GET", pointsURL, nil)
	pointsReq.Header.Set("User-Agent", "weather-service-example")
	resp, err := http.DefaultClient.Do(pointsReq)
	if err != nil {
		return "", 0, "", err
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	var pr PointsResponse
	if err := json.Unmarshal(body, &pr); err != nil {
		return "", 0, "", err
	}

	// Step 2: Call forecast API
	forecastURL := pr.Properties.Forecast
	foreReq, _ := http.NewRequest("GET", forecastURL, nil)
	foreReq.Header.Set("User-Agent", "weather-service-example")
	resp2, err := http.DefaultClient.Do(foreReq)
	if err != nil {
		return "", 0, "", err
	}
	defer resp2.Body.Close()

	body2, _ := ioutil.ReadAll(resp2.Body)
	var fr ForecastResponse
	if err := json.Unmarshal(body2, &fr); err != nil {
		return "", 0, "", err
	}

	// Step 3: Find today's daytime period
	today := time.Now().Format("2006-01-02")
	for _, p := range fr.Properties.Periods {
		if p.IsDaytime && len(p.StartTime) >= 10 && p.StartTime[:10] == today {
			return p.ShortForecast, p.Temperature, classify(p.Temperature), nil
		}
	}

	// Fallback to first period
	p := fr.Properties.Periods[0]
	return p.ShortForecast, p.Temperature, classify(p.Temperature), nil
}

func classify(temp int) string {
	switch {
	case temp >= 80:
		return "hot"
	case temp <= 50:
		return "cold"
	default:
		return "moderate"
	}
}

/*
Build & Run Instructions:

1. Run: go build -o weather-service
2. Start: ./weather-service
3. Query: http://localhost:8080/weather?lat=38.8977&lon=-77.0365

*/
