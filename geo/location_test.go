package geo

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestExtractFromRequest(t *testing.T) {
	tests := []struct {
		name     string
		headers  map[string]string
		wantLoc  *Location
	}{
		{
			name: "US location with all fields",
			headers: map[string]string{
				"X-Appengine-Country":     "US",
				"X-Appengine-Region":      "ca",
				"X-Appengine-City":        "San Francisco",
				"X-Appengine-CityLatLong": "37.7749,-122.4194",
			},
			wantLoc: &Location{
				Country:     "US",
				Region:      "ca",
				City:        "San Francisco",
				CityLatLong: "37.7749,-122.4194",
			},
		},
		{
			name: "International location",
			headers: map[string]string{
				"X-Appengine-Country": "FR",
				"X-Appengine-City":    "Paris",
			},
			wantLoc: &Location{
				Country: "FR",
				City:    "Paris",
			},
		},
		{
			name:    "No geo headers",
			headers: map[string]string{},
			wantLoc: &Location{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			for k, v := range tt.headers {
				req.Header.Set(k, v)
			}

			got := ExtractFromRequest(req)

			if got.Country != tt.wantLoc.Country {
				t.Errorf("Country = %v, want %v", got.Country, tt.wantLoc.Country)
			}
			if got.Region != tt.wantLoc.Region {
				t.Errorf("Region = %v, want %v", got.Region, tt.wantLoc.Region)
			}
			if got.City != tt.wantLoc.City {
				t.Errorf("City = %v, want %v", got.City, tt.wantLoc.City)
			}
			if got.CityLatLong != tt.wantLoc.CityLatLong {
				t.Errorf("CityLatLong = %v, want %v", got.CityLatLong, tt.wantLoc.CityLatLong)
			}

			// Verify DetectedAt is set to a recent time
			if time.Since(got.DetectedAt) > time.Second {
				t.Errorf("DetectedAt not set to current time: %v", got.DetectedAt)
			}

			// Verify DetectedFrom is empty (to be set by caller)
			if got.DetectedFrom != "" {
				t.Errorf("DetectedFrom should be empty, got %v", got.DetectedFrom)
			}
		})
	}
}

func TestLocation_IsUS(t *testing.T) {
	tests := []struct {
		name    string
		country string
		want    bool
	}{
		{"United States", "US", true},
		{"Canada", "CA", false},
		{"France", "FR", false},
		{"Empty", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := &Location{Country: tt.country}
			if got := l.IsUS(); got != tt.want {
				t.Errorf("IsUS() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLocation_IsValid(t *testing.T) {
	tests := []struct {
		name    string
		country string
		want    bool
	}{
		{"Valid US", "US", true},
		{"Valid France", "FR", true},
		{"Empty country", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := &Location{Country: tt.country}
			if got := l.IsValid(); got != tt.want {
				t.Errorf("IsValid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLocation_GetCountryName(t *testing.T) {
	tests := []struct {
		name    string
		country string
		want    string
	}{
		{"United States", "US", "United States"},
		{"Canada", "CA", "Canada"},
		{"United Kingdom", "GB", "United Kingdom"},
		{"France", "FR", "France"},
		{"Germany", "DE", "Germany"},
		{"Japan", "JP", "Japan"},
		{"China", "CN", "China"},
		{"India", "IN", "India"},
		{"Brazil", "BR", "Brazil"},
		{"Mexico", "MX", "Mexico"},
		{"Unknown code", "XX", "XX"},
		{"Empty", "", "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := &Location{Country: tt.country}
			if got := l.GetCountryName(); got != tt.want {
				t.Errorf("GetCountryName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLocation_GetDisplayString(t *testing.T) {
	tests := []struct {
		name string
		loc  *Location
		want string
	}{
		{
			name: "Full location",
			loc: &Location{
				Country: "US",
				Region:  "ca",
				City:    "San Francisco",
			},
			want: "San Francisco, ca, United States",
		},
		{
			name: "City and country only",
			loc: &Location{
				Country: "FR",
				City:    "Paris",
			},
			want: "Paris, France",
		},
		{
			name: "Country only",
			loc: &Location{
				Country: "JP",
			},
			want: "Japan",
		},
		{
			name: "No country (invalid)",
			loc:  &Location{},
			want: "Location Unknown",
		},
		{
			name: "Region and country only",
			loc: &Location{
				Country: "CA",
				Region:  "on",
			},
			want: "on, Canada",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.loc.GetDisplayString(); got != tt.want {
				t.Errorf("GetDisplayString() = %v, want %v", got, tt.want)
			}
		})
	}
}
