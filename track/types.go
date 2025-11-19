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

// Package track contains analytics helpers for recording visits and events.
// This file defines the structures used to store visit details and robot hits
// in Datastore and BigQuery.
package track

import (
	"time"

	"google.golang.org/appengine/v2/datastore"
)

type Visit struct {
	// DatastoreKey holds the datastore entity key when a visit is stored.
	DatastoreKey *datastore.Key `json:"datastoreKey" datastore:"-"`
	// Cookie uniquely identifies the visitor across requests.
	Cookie string `json:"cookie,omitempty"`
	// Session tracks consecutive page views within the same visit.
	Session string `json:"session,omitempty"`
	// URI is the request URI that was visited.
	URI string `json:"uri,omitempty"`
	// Referer records the Referer header from the request.
	Referer string `json:"referer,omitempty"`
	// Time is when the visit occurred.
	Time time.Time `json:"time,omitempty"`
	// Host is the HTTP host serving the request.
	Host string `json:"host,omitempty"`
	// RemoteAddr is the IP address of the visitor.
	RemoteAddr string `json:"remoteAddr,omitempty"`
	// InstanceId identifies the App Engine instance serving the visit.
	InstanceId string `json:"instanceId,omitempty"`
	// VersionId is the App Engine version serving the visit.
	VersionId string `json:"versionId,omitempty"`
	// Scheme is the HTTP scheme used (http or https).
	Scheme string `json:"scheme,omitempty"`
	// Country is the visitor country derived from geolocation.
	Country string `json:"country,omitempty"`
	// Region is the visitor region derived from geolocation.
	Region string `json:"region,omitempty"`
	// City is the visitor city derived from geolocation.
	City string `json:"city,omitempty"`
	// Lat is the latitude of the visitor city.
	Lat float64 `json:"lat,omitempty"`
	// Lon is the longitude of the visitor city.
	Lon float64 `json:"lon,omitempty"`
	// AcceptLanguage records the visitor's Accept-Language header.
	AcceptLanguage string `json:"acceptLanguage,omitempty"`
	// UserAgent captures the full user agent string.
	UserAgent string `json:"userAgent,omitempty"`
	// IsMobile reports whether the user agent is mobile.
	IsMobile bool `json:"isMobile,omitempty"`
	// IsBot indicates whether the visit came from a bot.
	IsBot bool `json:"isBot,omitempty"`
	// MozillaVersion is the Mozilla version extracted from the user agent.
	MozillaVersion string `json:"mozillaVersion,omitempty"`
	// Platform identifies the device platform in the user agent.
	Platform string `json:"platform,omitempty"`
	// OS is the operating system reported by the user agent.
	OS string `json:"os,omitempty"`
	// EngineName names the rendering engine used by the browser.
	EngineName string `json:"engineName,omitempty"`
	// EngineVersion is the version of the rendering engine.
	EngineVersion string `json:"engineVersion,omitempty"`
	// BrowserName is the browser name parsed from the user agent.
	BrowserName string `json:"browserName,omitempty"`
	// BrowserVersion is the browser version parsed from the user agent.
	BrowserVersion string `json:"browserVersion,omitempty"`
	// Category categorizes an analytics event when tracking actions.
	Category string `json:"category,omitempty"`
	// Action describes the user action that triggered the event.
	Action string `json:"action,omitempty"`
	// Label provides an optional label for the tracked event.
	Label string `json:"label,omitempty"`
	// Value stores a numeric value associated with the event.
	Value float64 `json:"value,omitempty"`
}

type RobotPage struct {
	// Time records when the robot accessed the page.
	Time time.Time `json:"time,omitempty"`
	// Name is an optional identifier for the robot event.
	Name string `json:"name,omitempty"`
	// URL is the full request URL visited by the robot.
	URL string `json:"url,omitempty"`
	// URI is the request URI visited by the robot.
	URI string `json:"uri,omitempty"`
	// Host is the host that served the request.
	Host string `json:"host,omitempty"`
	// RemoteAddr is the robot's IP address.
	RemoteAddr string `json:"remoteAddr,omitempty"`
	// UserAgent contains the robot's user agent string.
	UserAgent string `json:"userAgent,omitempty"`
	// Country is the robot's country based on geolocation.
	Country string `json:"country,omitempty"`
	// Region is the robot's region based on geolocation.
	Region string `json:"region,omitempty"`
	// City is the robot's city based on geolocation.
	City string `json:"city,omitempty"`
	// BotName is the name of the bot parsed from the user agent.
	BotName string `json:"botName,omitempty"`
	// BotVersion is the bot version parsed from the user agent.
	BotVersion string `json:"botVersion,omitempty"`
}

// TouchPointEvent captures a marketing touch point for web visitors. It records
// standard event metadata (category, action, label) plus request context and a
// JSON-encoded payload for event specific fields such as UTM parameters.
type TouchPointEvent struct {
	// Time is when the touch point occurred.
	Time time.Time `json:"time,omitempty"`
	// Category groups touch points by high-level category (for example "landing" or "campaign").
	Category string `json:"category,omitempty"`
	// Action describes what happened (for example "view" or "cta_click").
	Action string `json:"action,omitempty"`
	// Label provides an optional label for the touch point.
	Label string `json:"label,omitempty"`
	// Referer holds the HTTP Referer header for the request.
	Referer string `json:"referer,omitempty"`
	// Path is the HTTP request path (for example "/pricing").
	Path string `json:"path,omitempty"`
	// Host is the HTTP host serving the request.
	Host string `json:"host,omitempty"`
	// RemoteAddr is the client IP address.
	RemoteAddr string `json:"remoteAddr,omitempty"`
	// UserAgent captures the full user agent string for the visitor.
	UserAgent string `json:"userAgent,omitempty"`
	// PayloadJSON stores a JSON-encoded payload with arbitrary event fields.
	PayloadJSON string `json:"payloadJson,omitempty"`
}
