package geo

import "time"

// Session represents a user's authenticated session.
type Session struct {
	ID                string
	Active            bool
	ExpiresAt         time.Time
	AuthenticatedAt   time.Time
	IssuedAt          time.Time
	LastKnownLocation GeoPoint
}

// GeoPoint provides the latitude and longitude for
// a geographical location.
type GeoPoint struct {
	Latitude  float64
	Longitude float64
}

// Repository is an interface for a persistence
// layer. It defines GetSession for retrieving a
// user's session by ID.
type Repository interface {
	GetSession(id string) Session
}

// MockRepository implements Repository. It simply
// keeps a map object in memory. Useful for testing.
type MockRepository struct {
	data map[string]Session
}

// NewMockRepository returns an instance of MockRepository.
func NewMockRepository() MockRepository {
	d := map[string]Session{
		"65dea6f4-5d15-4e61-9eb7-f30190c0b2e2": {
			ID:              "65dea6f4-5d15-4e61-9eb7-f30190c0b2e2",
			Active:          true,
			ExpiresAt:       time.Now().Add(8 * time.Hour),
			AuthenticatedAt: time.Now().Add(-2 * time.Hour),
			IssuedAt:        time.Now().Add(-2 * time.Hour),
			LastKnownLocation: GeoPoint{ // Los Angeles
				Latitude:  34.026437,
				Longitude: -118.26109,
			},
		},
	}

	return MockRepository{data: d}
}

// GetSession implements Repository. It takes a session ID and
// returns a Session struct.
func (m MockRepository) GetSession(id string) Session {
	return m.data[id]
}
