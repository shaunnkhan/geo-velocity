package geo

import (
	"encoding/json"
	"errors"
	"log/slog"
	"math"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"
)

func TestGetRequiredParams(t *testing.T) {
	// setup test cases
	var tests = []struct {
		name      string
		params    url.Values
		sessionID string
		latitude  float64
		longitude float64
		err       error
	}{
		{
			name: "base case",
			params: url.Values{
				"session_id": {"1236"},
				"latitude":   {"34.026437"},
				"longitude":  {"-118.26109"},
			},
			sessionID: "1236",
			latitude:  34.026437,
			longitude: -118.26109,
			err:       nil,
		},
		{
			name: "missing session ID",
			params: url.Values{
				"session_id": {""},
				"latitude":   {"34.026437"},
				"longitude":  {"-118.26109"},
			},
			sessionID: "",
			latitude:  0.0,
			longitude: 0.0,
			err:       errors.New("malformed session_id"),
		},
	}

	// run tests
	for _, test := range tests {
		sessID, lat, lng, err := getRequiredParams(test.params)

		if sessID != test.sessionID || lat != test.latitude || lng != test.longitude || !compareErrors(err, test.err) {
			t.Errorf("%s - expected: %v, %v, %v, %v. got: %v, %v, %v, %v", test.name, test.sessionID, test.latitude, test.longitude, test.err, sessID, lat, lng, err)
		}
	}
}

func TestCalculateDistance(t *testing.T) {
	// setup test cases
	var tests = []struct {
		name    string
		currLat float64
		currLng float64
		prevLat float64
		prevLng float64
		unit    string
		dist    float64
	}{
		{
			name:    "base case",
			currLat: 34.026437,
			currLng: -118.26109,
			prevLat: 40.700583,
			prevLng: -74.004531,
			unit:    "km/h",
			dist:    3938.5890573185893,
		},
		{
			name:    "convert to mph",
			currLat: 34.026437,
			currLng: -118.26109,
			prevLat: 40.700583,
			prevLng: -74.004531,
			unit:    "mph",
			dist:    2447.3257782789688,
		},
	}

	// run tests
	for _, test := range tests {
		dist := calculateDistance(test.currLat, test.currLng, test.prevLat, test.prevLng, test.unit)

		if dist != test.dist {
			t.Errorf("%s - expected: %v. got: %v", test.name, test.dist, dist)
		}
	}
}

func TestCalculateSpeed(t *testing.T) {
	// setup test cases
	var tests = []struct {
		name     string
		distance float64
		begin    time.Time
		end      time.Time
		speed    float64
	}{
		{
			name:     "base case",
			distance: 3938.5890573185893,
			begin:    time.Now().Add(-2 * time.Hour),
			end:      time.Now(),
			speed:    1969.2945286127974,
		},
	}

	// run tests
	for _, test := range tests {
		speed := calculateSpeed(test.distance, test.begin, test.end)

		// need to round here due to time variance
		if math.Round(speed) != math.Round(test.speed) {
			t.Errorf("%s - expected: %v. got: %v", test.name, test.speed, speed)
		}
	}

}

func TestGetGeoSpeed(t *testing.T) {
	// Setup mock repository with test data
	mockRepo := NewMockRepository()

	// Create logger
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	// Create handler
	handler := NewGeoHandler(880.0, "km/h", mockRepo, logger)

	tests := []struct {
		name               string
		queryParams        map[string]string
		expectedStatusCode int
		expectedResponse   *Response
		expectedError      bool
		errorContains      string
	}{
		{
			name: "possible travel speed",
			queryParams: map[string]string{
				"session_id": "65dea6f4-5d15-4e61-9eb7-f30190c0b2e2",
				"latitude":   "33.418222",
				"longitude":  "-112.073945", // Phoenix
				"unit":       "mph",
				"max_speed":  "547",
			},
			expectedStatusCode: http.StatusOK,
			expectedResponse: &Response{
				Unit:             "mph",
				ImpossibleTravel: false,
			},
		},
		{
			name: "impossible travel speed",
			queryParams: map[string]string{
				"session_id": "65dea6f4-5d15-4e61-9eb7-f30190c0b2e2",
				"latitude":   "48.086989",
				"longitude":  "11.567296", // Munich
			},
			expectedStatusCode: http.StatusOK,
			expectedResponse: &Response{
				Unit:             "km/h",
				ImpossibleTravel: true,
			},
		},
		{
			name: "minimal parameters (defaults)",
			queryParams: map[string]string{
				"session_id": "65dea6f4-5d15-4e61-9eb7-f30190c0b2e2",
				"latitude":   "34.026437",
				"longitude":  "-118.26109",
			},
			expectedStatusCode: http.StatusOK,
			expectedResponse: &Response{
				Unit:             "km/h",
				ImpossibleTravel: false,
			},
		},
		{
			name: "missing session_id parameter",
			queryParams: map[string]string{
				"latitude":  "34.026437",
				"longitude": "-118.26109",
			},
			expectedStatusCode: http.StatusBadRequest,
			expectedError:      true,
			errorContains:      "malformed session_id",
		},
		{
			name: "missing latitude parameter",
			queryParams: map[string]string{
				"session_id": "65dea6f4-5d15-4e61-9eb7-f30190c0b2e2",
				"longitude":  "-118.26109",
			},
			expectedStatusCode: http.StatusBadRequest,
			expectedError:      true,
			errorContains:      "malformed latitude",
		},
		{
			name: "missing longitude parameter",
			queryParams: map[string]string{
				"session_id": "65dea6f4-5d15-4e61-9eb7-f30190c0b2e2",
				"latitude":   "34.026437",
			},
			expectedStatusCode: http.StatusBadRequest,
			expectedError:      true,
			errorContains:      "malformed longitude",
		},
		{
			name: "invalid latitude format",
			queryParams: map[string]string{
				"session_id": "65dea6f4-5d15-4e61-9eb7-f30190c0b2e2",
				"latitude":   "invalid",
				"longitude":  "-118.26109",
			},
			expectedStatusCode: http.StatusBadRequest,
			expectedError:      true,
			errorContains:      "malformed latitude",
		},
	}

	for _, test := range tests {
		// Build query string
		queryValues := url.Values{}
		for key, value := range test.queryParams {
			queryValues.Set(key, value)
		}

		// Create request
		req := httptest.NewRequest(http.MethodGet, "/geo-speed?"+queryValues.Encode(), nil)
		w := httptest.NewRecorder()

		// Call handler
		handler.GetGeoSpeed(w, req)

		// Check status code
		if w.Code != test.expectedStatusCode {
			t.Errorf("%s - expected status code: %d. got: %d", test.name, test.expectedStatusCode, w.Code)
		}

		// Check content type
		contentType := w.Header().Get("Content-Type")
		if contentType != "application/json" {
			t.Errorf("%s - expected Content-Type: 'application/json'. got: '%s'", test.name, contentType)
		}

		// Check response body
		if test.expectedError {
			// Check error response
			body := w.Body.String()
			if !strings.Contains(body, test.errorContains) {
				t.Errorf("%s - expected error message: '%s'. got: '%s'", test.name, test.errorContains, body)
			}
		} else {
			// Check successful response
			var response Response
			if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
				t.Fatalf("%s - failed to unmarshal response: %v", test.name, err)
			}

			// Check unit
			if response.Unit != test.expectedResponse.Unit {
				t.Errorf("%s - expected unit '%s'. got: '%s'", test.name, test.expectedResponse.Unit, response.Unit)
			}

			// Check that speed is a positive number
			if response.Speed < 0 {
				t.Errorf("%s - expected speed to be positive. got: %f", test.name, response.Speed)
			}

			// Check impossible travel flag if specified
			if test.expectedResponse.ImpossibleTravel {
				if !response.ImpossibleTravel {
					t.Errorf("%s - expected impossible travel to be true. got: false", test.name)
				}
			}
		}
	}
}

// Not pretty but using for this example
func compareErrors(err1, err2 error) bool {
	if err1 == nil && err2 == nil {
		return true
	} else if err1 == nil && err2 != nil {
		return false
	} else if err1 != nil && err2 == nil {
		return false
	}

	return err1.Error() == err2.Error()
}
