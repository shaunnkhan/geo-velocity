package geo

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	geopoint "github.com/kellydunn/golang-geo"
)

type Resposne struct {
	Speed            float64 `json:"speed"`
	Unit             string  `json:"unit"`
	ImpossibleTravel bool    `json:"impossible_travel"`
}

type GeoHandler struct {
	repo           Repository
	maxTravelSpeed float64
	unit           string
}

func NewGeoHandler(maxTravelSpeed float64, unit string, repo Repository) GeoHandler {
	return GeoHandler{
		repo:           repo,
		maxTravelSpeed: maxTravelSpeed,
		unit:           unit,
	}
}

func (g *GeoHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /geo-speed", g.GetGeoSpeed)
}

func (g GeoHandler) GetGeoSpeed(w http.ResponseWriter, r *http.Request) {
	// var req Request
	// if err := json.NewDecoder(r.Body).Decode(&req); err != nil {

	// }

	// get query params
	sessID := r.URL.Query().Get("session_id")
	unit := r.URL.Query().Get("unit")

	maxSpeed, err := strconv.ParseFloat(r.URL.Query().Get("max_speed"), 64)
	if err != nil {
		// todo
	}
	lat, err := strconv.ParseFloat(r.URL.Query().Get("latitude"), 64)
	if err != nil {
		// todo
	}
	lng, err := strconv.ParseFloat(r.URL.Query().Get("longitude"), 64)
	if err != nil {
		// todo
	}

	// maybe validate parameters. return 400 if missing must haves. assign defaults for others

	// retrieve session with last known location
	sess := g.repo.GetSession(sessID)

	// calculate distance
	dist := calculateDistance(lat, lng, sess.LasKnownLocation.Latitude, sess.LasKnownLocation.Longitude, unit)

	// calculate speed
	speed := calculateSpeed(dist, sess.AuthenticatedAt, time.Now())

	// determine if speed is impossible
	impossible := speed > maxSpeed

	// write response
	res := Resposne{
		Speed:            speed,
		Unit:             unit,
		ImpossibleTravel: impossible,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(res)
}

func calculateDistance(currLat, currLng, prevLat, prevLng float64, unit string) float64 {

	p1 := geopoint.NewPoint(currLat, currLng)
	p2 := geopoint.NewPoint(prevLat, prevLng)

	dist := p1.GreatCircleDistance(p2)

	if unit == "mph" {
		dist = dist / 1.609344
	}

	return dist
}

func calculateSpeed(distance float64, begin time.Time, end time.Time) float64 {
	return distance / begin.Sub(end).Abs().Hours()
}
