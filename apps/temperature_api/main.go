package main

import (
	"encoding/json"
	"log"
	"math"
	"math/rand/v2"
	"net/http"
	"os"
	"sync"
	"time"
)

type TemperatureResponse struct {
	Value       float64   `json:"value"`
	Unit        string    `json:"unit"`
	Timestamp   time.Time `json:"timestamp"`
	Location    string    `json:"location"`
	Status      string    `json:"status"`
	SensorID    string    `json:"sensor_id"`
	SensorType  string    `json:"sensor_type"`
	Description string    `json:"description"`
}

type generator struct {
	mu   sync.Mutex
	last map[string]float64
}

func main() {
	g := &generator{last: make(map[string]float64)}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, map[string]string{"status": "ok"})
	})
	mux.HandleFunc("GET /temperature", func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		writeJSON(w, g.reading(q.Get("location"), q.Get("sensorId")))
	})
	mux.HandleFunc("GET /temperature/{sensorId}", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, g.reading("", r.PathValue("sensorId")))
	})

	srv := &http.Server{
		Addr:              ":" + getEnv("PORT", "8081"),
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}
	log.Printf("temperature-api listening on %s", srv.Addr)
	log.Fatal(srv.ListenAndServe())
}

func (g *generator) reading(location, sensorID string) TemperatureResponse {
	location, sensorID = resolve(location, sensorID)
	return TemperatureResponse{
		Value:       g.next(sensorID),
		Unit:        "°C",
		Timestamp:   time.Now().UTC(),
		Location:    location,
		Status:      "active",
		SensorID:    sensorID,
		SensorType:  "temperature",
		Description: "Temperature sensor in " + location,
	}
}

func (g *generator) next(sensorID string) float64 {
	g.mu.Lock()
	defer g.mu.Unlock()

	v := randomTemperature()
	for v == g.last[sensorID] {
		v = randomTemperature()
	}
	g.last[sensorID] = v
	return v
}

func randomTemperature() float64 {
	return math.Round((18+rand.Float64()*8)*10) / 10
}

func resolve(location, sensorID string) (string, string) {
	// If no location is provided, use a default based on sensor ID
	if location == "" {
		switch sensorID {
		case "1":
			location = "Living Room"
		case "2":
			location = "Bedroom"
		case "3":
			location = "Kitchen"
		default:
			location = "Unknown"
		}
	}

	// If no sensor ID is provided, generate one based on location
	if sensorID == "" {
		switch location {
		case "Living Room":
			sensorID = "1"
		case "Bedroom":
			sensorID = "2"
		case "Kitchen":
			sensorID = "3"
		default:
			sensorID = "0"
		}
	}

	return location, sensorID
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("write response: %v", err)
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
