package geo

import (
	"encoding/json"
	"net/http"
)

type Request struct {
	UserID         string  `json:"userId"`
	SessionID      string  `json:"sessionId"`
	MaxTravelSpeed float64 `json:"maxTravelSpeed"`
	Unit           string  `json:"unit"`
	Latitude       float64 `json:"latitude"`
	Longitude      float64 `json:"longitude"`
}

type Resposne struct {
	Velocity         float64 `json:"velocity"`
	Unit             string  `json:"unit"`
	ImpossibleTravel bool    `json:"impossibleTravel"`
}

type GeoHandler struct {
}

func NewGeoHandler() GeoHandler {
	return GeoHandler{}
}

func (g *GeoHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /geo-velocity", g.GetGeoVelocity)
}

func (g GeoHandler) GetGeoVelocity(w http.ResponseWriter, r *http.Request) {
	var req Request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {

	}

	res := Resposne{
		Velocity:         500.00,
		Unit:             "MPH",
		ImpossibleTravel: true,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(res)
}
