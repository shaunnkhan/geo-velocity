package geo

import "time"

type Session struct {
	ID               string
	Active           bool
	ExpiresAt        time.Time
	AuthenticatedAt  time.Time
	IssuedAt         time.Time
	LasKnownLocation GeoPoint
}

type GeoPoint struct {
	Latitude  float64
	Longitude float64
}

type Repository interface {
	GetSession(id string) Session
}

type MockRepository struct {
	data map[string]Session
}

func NewMockRepository() MockRepository {
	d := map[string]Session{
		"65dea6f4-5d15-4e61-9eb7-f30190c0b2e2": {
			ID:              "65dea6f4-5d15-4e61-9eb7-f30190c0b2e2",
			Active:          true,
			ExpiresAt:       time.Now().Add(8 * time.Hour),
			AuthenticatedAt: time.Now().Add(-2 * time.Hour),
			IssuedAt:        time.Now().Add(-2 * time.Hour),
			LasKnownLocation: GeoPoint{ // Los Angeles
				Latitude:  34.049914,
				Longitude: -118.236213,
			},
		},
	}

	return MockRepository{data: d}
}

func (m MockRepository) GetSession(id string) Session {
	return m.data[id]
}
