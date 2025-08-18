// Copyright 2025 Patrick Deglon
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package common provides shared helpers used across the application.
// It includes utilities for reading and writing HTTP requests and responses
// such as the JSON/XML helpers found in interfaces.go.
package common

import (
	"net"
	"net/http"
	"strconv"
	"time"
)

type Visitor struct {
	// Key is the datastore key for the visitor and also the cookie value.
	Key string `json:"key,omitempty"`
	// Cookie stores the same ID to easily serialize the cookie value.
	Cookie string `json:"cookie,omitempty"`
	// Host is the host name seen when the cookie was issued.
	Host string `json:"host,omitempty"`
	// CreatedTimestamp records when the cookie was created.
	CreatedTimestamp string `json:"createdTimestamp,omitempty"`
	// CreatedIP contains the client IP address at creation time.
	CreatedIP string `json:"createdIP,omitempty"`
	// CreatedReferer holds the HTTP referer header.
	CreatedReferer string `json:"createdReferer,omitempty"`
	// Geolocation fields derived from App Engine headers.
	CreatedCountry string `json:"createdCountry,omitempty"`
	CreatedRegion  string `json:"createdRegion,omitempty"`
	CreatedCity    string `json:"createdCity,omitempty"`
	// CreatedUserAgent stores the raw user agent string.
	CreatedUserAgent string `json:"createdUserAgent,omitempty"`
	// CreatedIsMobile indicates whether the UA looked like a mobile device.
	CreatedIsMobile bool `json:"createdIsMobile,omitempty"`
	// CreatedIsBot indicates whether the UA was identified as a bot.
	CreatedIsBot bool `json:"createdIsBot,omitempty"`
	// Parsed user agent details.
	CreatedMozillaVersion string `json:"createdMozillaVersion,omitempty"`
	CreatedPlatform       string `json:"createdPlatform,omitempty"`
	CreatedOS             string `json:"createdOS,omitempty"`
	// Rendering engine details extracted from the UA string.
	CreatedEngineName    string `json:"createdEngineName,omitempty"`
	CreatedEngineVersion string `json:"createdEngineVersion,omitempty"`
	// Browser details extracted from the UA string.
	CreatedBrowserName    string `json:"createdBrowserName,omitempty"`
	CreatedBrowserVersion string `json:"createdBrowserVersion,omitempty"`
}

// ClearCookie removes the visitor ID cookie from the client by sending an
// expired cookie back to the browser.
func ClearCookie(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     "ID",
		Value:    "",
		Path:     "/",
		Expires:  time.Now(),
		SameSite: http.SameSiteLaxMode,
	})
}

// DoesCookieExists reports whether the request contains a non-empty visitor ID
// cookie.
func DoesCookieExists(r *http.Request) bool {
	cookie, err := r.Cookie("ID")
	if err != nil || cookie == nil || cookie.Value == "" {
		return false
	}
	return true
}

// GetCookieID retrieves the visitor ID cookie or creates a new one if missing.
//
// The created cookie is secured with the following attributes:
//   - Path is always set to "/" so the ID is sent for all application routes.
//   - Expires is 30 days in the future, giving the cookie a one month lifetime.
//   - HttpOnly is true to prevent JavaScript access.
//   - Secure is true for production (false for localhost/127.0.0.1) to support local development.
//   - SameSite is Lax to protect against CSRF while allowing normal navigation.
//   - Domain is only set when the host is neither localhost nor 127.0.0.1.
func GetCookieID(w http.ResponseWriter, r *http.Request) string {
	Debug(">>>> GetCookieID")

	var id string
	cookie, err := r.Cookie("ID")
	Debug("ID cookie: %v", cookie)
	if err != nil || cookie == nil || cookie.Value == "" {
		Debug("Error: %v", err)
		Debug("New Cookie...")
		ts := strconv.FormatInt(time.Now().UnixNano(), 10)
		id = MD5(ts + r.RemoteAddr)
		host := r.Host
		if h, _, err := net.SplitHostPort(host); err == nil {
			host = h
		}
		
		// Determine if we're on localhost for development
		isLocalhost := host == "localhost" || host == "127.0.0.1"
		
		ck := &http.Cookie{
			Name:     "ID",
			Value:    id,
			Path:     "/",
			Expires:  time.Now().Add(time.Hour * 24 * 30),
			HttpOnly: true,
			Secure:   !isLocalhost, // false for localhost, true for production
			SameSite: http.SameSiteLaxMode,
		}
		if !isLocalhost {
			// Set the domain so the cookie is shared across subdomains
			ck.Domain = host
		}
		http.SetCookie(w, ck)
		Debug("New Cookie = %v", id)
		/*
			key := datastore.NewKey(c, "Visitors", id, 0, nil)
			ua := user_agent.New(r.Header.Get("User-Agent"))
			engineName, engineversion := ua.Engine()
			browserName, browserVersion := ua.Browser()
			newVisitor := Visitor{
				Key:                   id,
				Cookie:                id,
				CreatedTimestamp:      ts,
				Host:                  r.Host,
				CreatedIP:             r.RemoteAddr,
				CreatedReferer:        r.Header.Get("Referer"),
				CreatedCountry:        r.Header.Get("X-AppEngine-Country"),
				CreatedRegion:         r.Header.Get("X-AppEngine-Region"),
				CreatedCity:           r.Header.Get("X-AppEngine-City"),
				CreatedUserAgent:      r.Header.Get("User-Agent"),
				CreatedIsMobile:       ua.Mobile(),
				CreatedIsBot:          ua.Bot(),
				CreatedMozillaVersion: ua.Mozilla(),
				CreatedPlatform:       ua.Platform(),
				CreatedOS:             ua.OS(),
				CreatedEngineName:     engineName,
				CreatedEngineVersion:  engineversion,
				CreatedBrowserName:    browserName,
				CreatedBrowserVersion: browserVersion,
			}
			key, err := datastore.Put(c, key, &newVisitor)
			if err != nil {
				log.Errorf(c, "Error while storing cookie %v in datastore: %v", id, err)
			} else {
				log.Infof(c, "New visitor %v stored in datastore under key %v", id,
					key.IntID())
			}
		*/
	} else {
		id = cookie.Value
		Debug("Existing ID Cookie = %v", id)
	}
	return id
}
