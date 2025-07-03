# geo-velocity
Geo-velocity Service

## Overview

The geo-velocity service is an HTTP API that calculates travel speed between geographic locations and determines if the calculated speed represents impossible travel. This service is useful for fraud detection, user authentication verification, and security monitoring by identifying when a user appears to move between locations at an unrealistic speed.

The service:
- Calculates the distance between two geographic points using the Haversine formula
- Computes the travel speed based on the time elapsed between authentication events
- Determines if the calculated speed exceeds a configurable maximum threshold
- Supports both metric (km/h) and imperial (mph) units

## Usage

### Prerequisites

- Go 1.24.4 or higher
- Make

### Installation

1. Clone the repository:
```bash
git clone https://github.com/shaunnkhan/geo-velocity.git
cd geo-velocity
```

2. Install dependencies:
```bash
go mod download
```

### Running the Service

Using Make:
```bash
make run
```

#### Command Line Options

The service accepts the following command-line flags:

- `-addr` (default: `:8080`): Server port
- `-max-speed` (default: `880.0`): Default maximum allowed speed
- `-unit` (default: `km/h`): Unit of speed (mph or km/h)

Example with custom options:
```bash
go run main.go -addr=:3000 -max-speed=500 -unit=mph
```

### Running Tests

Using Make:
```bash
make test
```

### API Endpoints

#### GET /geo-speed

Calculates the travel speed between a user's last known location and their current location.

**Query Parameters:**
- `session_id` (required): User session identifier
- `latitude` (required): Current latitude
- `longitude` (required): Current longitude
- `unit` (optional): Speed unit (mph or km/h) - defaults to service configuration
- `max_speed` (optional): Maximum allowed speed - defaults to service configuration

**Example Request:**
```
GET /geo-speed?session_id=65dea6f4-5d15-4e61-9eb7-f30190c0b2e2&latitude=48.183085&longitude=12.035587&unit=km/h&max_speed=805.00
```

**Response:**
```json
{
  "speed": 1220.9894373210473,
  "unit": "km/h",
  "impossible_travel": true
}
```

**Error Response (400 Bad Request):**
```json
{
  "error_msg": "malformed session_id"
}
```
