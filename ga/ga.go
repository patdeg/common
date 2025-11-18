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

// Package ga provides helpers for sending hits to Google Analytics using the
// Measurement Protocol. The package exposes helper types and functions to
// collect information from an HTTP request and transmit it to the GA endpoint.
//
// Basic usage:
//
//	event := ga.GetEvent(r) // build GAEvent from *http.Request
//	ga.TrackGAPage(r.Context(), ga.PropertyID, event)
//
//	event.Category = "signup"
//	event.Action = "click"
//	ga.TrackGAEvent(r.Context(), ga.PropertyID, event)
//
// The PropertyID variable holds the default Google Analytics Tracking ID.
package ga

import (
	"bytes"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/patdeg/common"
	"golang.org/x/net/context"
	"google.golang.org/appengine/v2/mail"
	"google.golang.org/appengine/v2/urlfetch"
)

var (
	PropertyID string = "UA-63208527-1"
	// PropertyID string = "UA-68699208-1"
)

// https://developers.google.com/analytics/devguides/collection/protocol/v1/parameters
// GAEvent contains Measurement Protocol parameters for a Google Analytics hit.
// Each field corresponds to a query parameter in the protocol. Empty fields are
// omitted when sending the hit.
type GAEvent struct {
	// cc – Campaign Content parameter.
	CampaignContent string `json:"CampaignContent,omitempty"`
	// cd1 – Custom Dimension 1.
	CustomDimension1 string `json:"CustomDimension1,omitempty"`
	// cd2 – Custom Dimension 2.
	CustomDimension2 string `json:"CustomDimension2,omitempty"`
	// cd3 – Custom Dimension 3.
	CustomDimension3 string `json:"CustomDimension3,omitempty"`
	// cd4 – Custom Dimension 4.
	CustomDimension4 string `json:"CustomDimension4,omitempty"`
	// cd5 – Custom Dimension 5.
	CustomDimension5 string `json:"CustomDimension5,omitempty"`
	// cd6 – Custom Dimension 6.
	CustomDimension6 string `json:"CustomDimension6,omitempty"`
	// cd7 – Custom Dimension 7.
	CustomDimension7 string `json:"CustomDimension7,omitempty"`
	// cd8 – Custom Dimension 8.
	CustomDimension8 string `json:"CustomDimension8,omitempty"`
	// cd9 – Custom Dimension 9.
	CustomDimension9 string `json:"CustomDimension9,omitempty"`
	// cid – Client ID used to identify a visitor.
	Guid string `json:"Guid,omitempty"`
	// uid – User ID for logged in users.
	UserId string `json:"UserId,omitempty"`
	// ck – Campaign Keyword parameter.
	CampaignKeyword string `json:"CampaignKeyword,omitempty"`
	// cm – Campaign Medium parameter.
	CampaignMedium string `json:"CampaignMedium,omitempty"`
	// cm1 – Custom Metric 1.
	CustomMetric1 string `json:"CustomMetric1,omitempty"`
	// cm2 – Custom Metric 2.
	CustomMetric2 string `json:"CustomMetric2,omitempty"`
	// cm3 – Custom Metric 3.
	CustomMetric3 string `json:"CustomMetric3,omitempty"`
	// cm4 – Custom Metric 4.
	CustomMetric4 string `json:"CustomMetric4,omitempty"`
	// cm5 – Custom Metric 5.
	CustomMetric5 string `json:"CustomMetric5,omitempty"`
	// cm6 – Custom Metric 6.
	CustomMetric6 string `json:"CustomMetric6,omitempty"`
	// cm7 – Custom Metric 7.
	CustomMetric7 string `json:"CustomMetric7,omitempty"`
	// cm8 – Custom Metric 8.
	CustomMetric8 string `json:"CustomMetric8,omitempty"`
	// cm9 – Custom Metric 9.
	CustomMetric9 string `json:"CustomMetric9,omitempty"`
	// cn – Campaign Name.
	CampaignName string `json:"CampaignName,omitempty"`
	// cs – Campaign Source.
	CampaignSource string `json:"CampaignSource,omitempty"`
	// cu – Currency Code (e.g. USD).
	CurrencyCode string `json:"CurrencyCode,omitempty"`
	// dh – Document Hostname.
	DocumentHostName string `json:"DocumentHostName,omitempty"`
	// dl – Document Location URL.
	DocumentLocationURL string `json:"DocumentLocationURL,omitempty"`
	// dp – Document Path.
	DocumentPath string `json:"DocumentPath,omitempty"`
	// dr – HTTP Referer header.
	Referer string `json:"Referer,omitempty"`
	// dt – Document Title.
	DocumentTitle string `json:"DocumentTitle,omitempty"`
	// ea – Event Action parameter.
	Action string `json:"Action,omitempty"`
	// ec – Event Category parameter.
	Category string `json:"Category,omitempty"`
	// el – Event Label parameter.
	Label string `json:"Label,omitempty"`
	// ev – Event Value parameter.
	Value string `json:"Value,omitempty"`
	// exd – Exception description.
	ExceptionDescription string `json:"ExceptionDescription,omitempty"`
	// exf – 1 for fatal exceptions.
	IsExceptionFatal string `json:"IsExceptionFatal,omitempty"`
	// gclid – Google AdWords ID.
	GoogleAdWordsID string `json:"GoogleAdWordsID,omitempty"`
	// ic – Item Code.
	ItemCode string `json:"ItemCode,omitempty"`
	// in – Item Name.
	ItemName string `json:"ItemName,omitempty"`
	// ip – Item Price.
	ItemPrice string `json:"ItemPrice,omitempty"`
	// iq – Item Quantity.
	ItemQuantity string `json:"ItemQuantity,omitempty"`
	// iv – Item Category.
	ItemCategory string `json:"ItemCategory,omitempty"`
	// sa – Social Action.
	SocialAction string `json:"SocialAction,omitempty"`
	// sn – Social Network.
	SocialNetwork string `json:"SocialNetwork,omitempty"`
	// st – Social Action Target.
	SocialActionTarget string `json:"SocialActionTarget,omitempty"`
	// ta – Transaction Affiliation.
	TransactionAffiliation string `json:"TransactionAffiliation,omitempty"`
	// ti – Transaction ID.
	TransactionID string `json:"TransactionID,omitempty"`
	// ts – Transaction Shipping.
	TransactionShipping string `json:"TransactionShipping,omitempty"`
	// tt – Transaction Tax.
	TransactionTax string `json:"TransactionTax,omitempty"`
	// ua – User Agent string.
	Agent string `json:"Agent,omitempty"`
	// uip – IP address of the user.
	IP string `json:"IP,omitempty"`
	// ul – User Language.
	UserLanguage string `json:"UserLanguage,omitempty"`
	// xid – Experiment ID.
	ExperimentID string `json:"ExperimentID,omitempty"`
	// xvar – Experiment Variant.
	ExperimentVariant string `json:"ExperimentVariant,omitempty"`
}

func setIfNotEmpty(v *url.Values, key string, value string) {
	if value != "" {
		v.Set(key, value)
	}
}

func setEvent(etype string, event GAEvent) url.Values {
	v := url.Values{}
	v.Set("v", "1")
	v.Set("t", etype)

	setIfNotEmpty(&v, "cc", event.CampaignContent)
	setIfNotEmpty(&v, "cd1", event.CustomDimension1)
	setIfNotEmpty(&v, "cd2", event.CustomDimension2)
	setIfNotEmpty(&v, "cd3", event.CustomDimension3)
	setIfNotEmpty(&v, "cd4", event.CustomDimension4)
	setIfNotEmpty(&v, "cd5", event.CustomDimension5)
	setIfNotEmpty(&v, "cd6", event.CustomDimension6)
	setIfNotEmpty(&v, "cd7", event.CustomDimension7)
	setIfNotEmpty(&v, "cd8", event.CustomDimension8)
	setIfNotEmpty(&v, "cd9", event.CustomDimension9)
	setIfNotEmpty(&v, "cid", event.Guid)
	setIfNotEmpty(&v, "uid", event.UserId)
	setIfNotEmpty(&v, "ck", event.CampaignKeyword)
	setIfNotEmpty(&v, "cm", event.CampaignMedium)
	setIfNotEmpty(&v, "cm1", event.CustomMetric1)
	setIfNotEmpty(&v, "cm2", event.CustomMetric2)
	setIfNotEmpty(&v, "cm3", event.CustomMetric3)
	setIfNotEmpty(&v, "cm4", event.CustomMetric4)
	setIfNotEmpty(&v, "cm5", event.CustomMetric5)
	setIfNotEmpty(&v, "cm6", event.CustomMetric6)
	setIfNotEmpty(&v, "cm7", event.CustomMetric7)
	setIfNotEmpty(&v, "cm8", event.CustomMetric8)
	setIfNotEmpty(&v, "cm9", event.CustomMetric9)
	setIfNotEmpty(&v, "cn", event.CampaignName)
	setIfNotEmpty(&v, "cs", event.CampaignSource)
	setIfNotEmpty(&v, "cu", event.CurrencyCode)
	setIfNotEmpty(&v, "dh", event.DocumentHostName)
	setIfNotEmpty(&v, "dl", event.DocumentLocationURL)
	setIfNotEmpty(&v, "dp", event.DocumentPath)
	setIfNotEmpty(&v, "dr", event.Referer)
	setIfNotEmpty(&v, "dt", event.DocumentTitle)
	setIfNotEmpty(&v, "ea", event.Action)
	setIfNotEmpty(&v, "ec", event.Category)
	setIfNotEmpty(&v, "el", event.Label)
	setIfNotEmpty(&v, "ev", event.Value)
	setIfNotEmpty(&v, "exd", event.ExceptionDescription)
	setIfNotEmpty(&v, "exf", event.IsExceptionFatal)
	setIfNotEmpty(&v, "gclid", event.GoogleAdWordsID)
	setIfNotEmpty(&v, "ic", event.ItemCode)
	setIfNotEmpty(&v, "in", event.ItemName)
	setIfNotEmpty(&v, "ip", event.ItemPrice)
	setIfNotEmpty(&v, "iq", event.ItemQuantity)
	setIfNotEmpty(&v, "iv", event.ItemCategory)
	setIfNotEmpty(&v, "sa", event.SocialAction)
	setIfNotEmpty(&v, "sn", event.SocialNetwork)
	setIfNotEmpty(&v, "st", event.SocialActionTarget)
	setIfNotEmpty(&v, "ta", event.TransactionAffiliation)
	setIfNotEmpty(&v, "ti", event.TransactionID)
	setIfNotEmpty(&v, "ts", event.TransactionShipping)
	setIfNotEmpty(&v, "tt", event.TransactionTax)
	setIfNotEmpty(&v, "ua", event.Agent)
	setIfNotEmpty(&v, "uip", event.IP)
	setIfNotEmpty(&v, "ul", event.UserLanguage)
	setIfNotEmpty(&v, "xid", event.ExperimentID)
	setIfNotEmpty(&v, "xvar", event.ExperimentVariant)

	return v
}

// GetEvent populates a GAEvent from the HTTP request. It extracts common
// information such as the visitor ID, IP address and referer that are often
// included in Measurement Protocol hits.
func GetEvent(r *http.Request) GAEvent {
	guid := ""
	cookie, err := r.Cookie("ID")
	if err == nil {
		if cookie != nil {
			guid = cookie.Value
		}
	}
	if guid == "" {
		// Use SecureHash (SHA-256) instead of MD5 for visitor identification
		guid = common.Encrypt(r.Context(), "", common.SecureHash(r.RemoteAddr+r.Header.Get("User-Agent")))
	}

	query := ""
	if r.Header.Get("Referer") != "" {
		if refererUrl, err := url.Parse(r.Header.Get("Referer")); err == nil {
			query = refererUrl.Query().Get("q")
		}
	}
	if query == "" {
		query = r.FormValue("k")
	}

	socialNetwork := ""
	referer := strings.ToLower(r.Referer())
	switch {
	case strings.Contains(referer, "facebook"):
		socialNetwork = "Facebook"
	case strings.Contains(referer, "twitter"):
		socialNetwork = "Twitter"
	case strings.Contains(referer, "linkedin"):
		socialNetwork = "LinkedIn"
	}

	event := GAEvent{
		Guid:                guid,
		IP:                  r.RemoteAddr,
		DocumentHostName:    r.Host,
		DocumentLocationURL: r.RequestURI,
		DocumentPath:        r.URL.Path,
		Referer:             r.Referer(),
		DocumentTitle:       r.URL.Path,
		Agent:               r.Header.Get("User-Agent"),
		UserLanguage:        r.Header.Get("Accept-Language"),
		CampaignKeyword:     query,
		CampaignName:        r.FormValue("cm"),
		SocialNetwork:       socialNetwork,
	}

	return event
}

// TrackGAPage sends a pageview hit to Google Analytics. PropertyID represents
// the GA tracking ID (tid) and event provides the parameters for the hit.
//
// Example:
//
//	event := ga.GetEvent(r)
//	ga.TrackGAPage(r.Context(), ga.PropertyID, event)
func TrackGAPage(c context.Context, PropertyID string, event GAEvent) {
	endpointUrl := "https://www.google-analytics.com/collect?"
	v := setEvent("pageview", event)
	v.Set("tid", PropertyID)
	payload_data := v.Encode()
	common.Info("GA: Calling %v with %v", endpointUrl, payload_data)

	req, err := http.NewRequest("POST", endpointUrl, bytes.NewBufferString(payload_data))
	if err != nil {
		common.Error("Error while tracking Google Analytics: %v", err)
		return
	}
	resp, err := urlfetch.Client(c).Do(req)
	if err != nil {
		common.Error("Error while tracking Google Analytics: %v", err)
		return
	}

	common.Debug("GA status code %v", resp.StatusCode)
}

// TrackGAEvent sends an event hit to Google Analytics using the supplied
// tracking ID and parameters.
//
// Example:
//
//	event := ga.GetEvent(r)
//	event.Category = "signup"
//	event.Action = "click"
//	ga.TrackGAEvent(r.Context(), ga.PropertyID, event)
func TrackGAEvent(c context.Context, PropertyID string, event GAEvent) {
	endpointUrl := "https://www.google-analytics.com/collect?"
	v := setEvent("event", event)
	v.Set("tid", PropertyID)
	payload_data := v.Encode()
	common.Info("GA: Calling %v with %v", endpointUrl, payload_data)

	req, err := http.NewRequest("POST", endpointUrl, bytes.NewBufferString(payload_data))
	if err != nil {
		common.Error("Error while tracking Google Analytics: %v", err)
		return
	}
	resp, err := urlfetch.Client(c).Do(req)
	if err != nil {
		common.Error("Error while tracking Google Analytics: %v", err)
		return
	}

	common.Debug("GA status code %v", resp.StatusCode)
}

// GATrackServeError renders an HTTP error response and records the failure as a
// pageview in Google Analytics. If the error is fatal, a notification email is
// also sent to AppEngineEmail.
func GATrackServeError(w http.ResponseWriter, r *http.Request, PropertyID string,
	errorTitle, errorMessage string, err error, code int, isFatal bool, AppEngineEmail string) {
	c := r.Context()
	if err != nil {
		common.Error("%v: %v", errorMessage, err.Error())
	}
	event := GetEvent(r)
	event.ExceptionDescription = errorMessage
	if err != nil {
		event.ExceptionDescription += ": " + err.Error()
	}
	if isFatal {
		event.IsExceptionFatal = "1"
		msg := &mail.Message{
			Sender:  AppEngineEmail,
			To:      []string{AppEngineEmail},
			Subject: "Fatal Error " + errorTitle,
			Body:    fmt.Sprintf(`There was a fatal error on %v at %v: %v`, r.Host, time.Now(), errorMessage),
		}
		if err := mail.Send(c, msg); err != nil {
			common.Error("Couldn't send error email: %v", err)
		}
	} else {
		event.IsExceptionFatal = ""
	}
	TrackGAPage(c, PropertyID, event)
	http.Error(w, errorTitle, code)
}
