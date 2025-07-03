// Package geo implements the core business logic for
// this service. It contains logic for determining
// impossible travel, registering http routes, and implements
// a repository.
package geo

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
	"time"

	geopoint "github.com/kellydunn/golang-geo"
)

// Response is the API response and contains:
//
//	Speed - The rate at which a user would have to travel
//	Unit - The unit of speed (mph or km/h)
//	ImpossibleTravel - A boolean indicating if speed of travel is impossible
type Response struct {
	Speed            float64 `json:"speed"`
	Unit             string  `json:"unit"`
	ImpossibleTravel bool    `json:"impossible_travel"`
}

// GeoHandler is the main handler of the geo service. It
// registers http routes and implements core logic.
type GeoHandler struct {
	repo           Repository
	maxTravelSpeed float64
	unit           string
	logger         *slog.Logger
}

// NewGeoHandler creates and returns an instance of GeoHandler.
func NewGeoHandler(maxTravelSpeed float64, unit string, repo Repository, logger *slog.Logger) GeoHandler {
	return GeoHandler{
		repo:           repo,
		maxTravelSpeed: maxTravelSpeed,
		unit:           unit,
		logger:         logger,
	}
}

// RegisterRoutes takes a ServerMux and registers routes with respective handlers.
func (g *GeoHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /geo-speed", g.GetGeoSpeed)
}

// GetGeoSpeed handles incoming requests to determine impossible travel speed
// by parsing query string parameters, retreiving session, and calulating
// distance and speed. It returns a Response instance to the caller.
// Example request:
// /geo-speed?session_id=65dea6f4-5d15-4e61-9eb7-f30190c0b2e2&unit=km/h&max_speed=805.00&latitude=48.183085&longitude=12.035587
func (g GeoHandler) GetGeoSpeed(w http.ResponseWriter, r *http.Request) {

	/************************************************
	* STEP 1. Capture and validate input parameters
	************************************************/
	// Validate required query params
	sessID, lat, lng, err := getRequiredParams(r.URL.Query())
	if err != nil {
		g.logger.Warn("bad request", "error", err)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		msg := fmt.Sprintf("{\"error_msg\":\"%s\"}", err)
		w.Write([]byte(msg))
		return
	}

	// Validate optional query params (unit & max_speed)
	// Defaults to service command line arguments
	unit := r.URL.Query().Get("unit")
	if unit == "" {
		unit = g.unit
	}

	maxSpeed, err := strconv.ParseFloat(r.URL.Query().Get("max_speed"), 64)
	if err != nil {
		maxSpeed = g.maxTravelSpeed
	}

	/************************************************
	* STEP 2. Calculate travel speed
	************************************************/
	// Retrieve session with last known location
	sess := g.repo.GetSession(sessID)

	// Calculate distance
	dist := calculateDistance(
		lat,
		lng,
		sess.LastKnownLocation.Latitude,
		sess.LastKnownLocation.Longitude,
		unit,
	)

	// Calculate speed
	speed := calculateSpeed(dist, sess.AuthenticatedAt, time.Now())

	/************************************************
	* STEP 3. Make speed decision and write response
	************************************************/
	// Determine if speed is impossible
	impossible := speed > maxSpeed

	// Write response
	res := Response{
		Speed:            speed,
		Unit:             unit,
		ImpossibleTravel: impossible,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(res)
}

// getRequiredParams takes URL values to retrieve, parse, and validate
// required parameters (session_id, latitude, longitude). An error is
// returned if validation fails.
func getRequiredParams(params url.Values) (string, float64, float64, error) {
	sessID := params.Get("session_id")
	if sessID == "" {
		return "", 0.0, 0.0, errors.New("malformed session_id")
	}

	lat, err := strconv.ParseFloat(params.Get("latitude"), 64)
	if err != nil {
		return "", 0.0, 0.0, errors.New("malformed latitude")
	}

	lng, err := strconv.ParseFloat(params.Get("longitude"), 64)
	if err != nil {
		return "", 0.0, 0.0, errors.New("malformed longitude")
	}

	return sessID, lat, lng, nil
}

// calculateDistance will calculate the distance between two geopoint cooridnates using
// the Haversine method. Default unit is kilometers but can be converted to miles based
// on the unit parameter.
func calculateDistance(currLat, currLng, prevLat, prevLng float64, unit string) float64 {

	p1 := geopoint.NewPoint(currLat, currLng)
	p2 := geopoint.NewPoint(prevLat, prevLng)

	dist := p1.GreatCircleDistance(p2)

	if unit == "mph" {
		dist = dist / 1.609344
	}

	return dist
}

// calculateSpeed will calculate the speed of travel based on distance and begin
// and end times.
func calculateSpeed(distance float64, begin time.Time, end time.Time) float64 {
	return distance / begin.Sub(end).Abs().Hours()
}
