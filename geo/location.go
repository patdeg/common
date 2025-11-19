// Package geo provides geo-location utilities for AppEngine
package geo

import (
	"net/http"
	"time"
)

// Location represents geographic information from AppEngine headers
type Location struct {
	Country      string    `datastore:"country"`        // ISO 3166-1 alpha-2 country code (e.g., "US")
	Region       string    `datastore:"region"`         // Region/state (e.g., "ca" for California)
	City         string    `datastore:"city"`           // City name
	CityLatLong  string    `datastore:"city_latlong"`   // Latitude,longitude
	DetectedAt   time.Time `datastore:"detected_at"`    // When this location was detected
	DetectedFrom string    `datastore:"detected_from"`  // "session" or "signup"
}

// ExtractFromRequest extracts geo-location from AppEngine headers
// AppEngine automatically adds these headers based on the client's IP address
func ExtractFromRequest(r *http.Request) *Location {
	return &Location{
		Country:      r.Header.Get("X-Appengine-Country"),
		Region:       r.Header.Get("X-Appengine-Region"),
		City:         r.Header.Get("X-Appengine-City"),
		CityLatLong:  r.Header.Get("X-Appengine-CityLatLong"),
		DetectedAt:   time.Now().UTC(),
		DetectedFrom: "", // Will be set by caller ("session" or "signup")
	}
}

// IsUS checks if the location is in the United States
func (l *Location) IsUS() bool {
	return l.Country == "US"
}

// IsValid checks if location data was successfully extracted
func (l *Location) IsValid() bool {
	return l.Country != ""
}

// GetCountryName returns a human-readable country name
// This is a simple implementation - for production you may want a full lookup table
func (l *Location) GetCountryName() string {
	switch l.Country {
	case "US":
		return "United States"
	case "CA":
		return "Canada"
	case "GB":
		return "United Kingdom"
	case "FR":
		return "France"
	case "DE":
		return "Germany"
	case "JP":
		return "Japan"
	case "CN":
		return "China"
	case "IN":
		return "India"
	case "BR":
		return "Brazil"
	case "MX":
		return "Mexico"
	default:
		if l.Country != "" {
			return l.Country // Return code if name not found
		}
		return "Unknown"
	}
}

// GetDisplayString returns a formatted location string for display
func (l *Location) GetDisplayString() string {
	if !l.IsValid() {
		return "Location Unknown"
	}

	parts := []string{}
	if l.City != "" {
		parts = append(parts, l.City)
	}
	if l.Region != "" {
		parts = append(parts, l.Region)
	}
	parts = append(parts, l.GetCountryName())

	result := ""
	for i, part := range parts {
		if i > 0 {
			result += ", "
		}
		result += part
	}
	return result
}
