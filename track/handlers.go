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

// Package track provides helpers for analytics collection. This file contains
// HTTP handlers used by the tracking service. Handlers are provided to create
// daily BigQuery tables, serve the tracking pixel and record outbound clicks.
package track

import (
	"fmt"
	"net/http"
	"time"

	"github.com/patdeg/common"

	"google.golang.org/appengine/v2/user"
)

func CreateTodayVisitsTableInBigQueryHandler(w http.ResponseWriter, r *http.Request) {
	c := r.Context()
	common.Info(">>>>>>>> CreateTodayVisitsTableInBigQueryHandler")

	isAdmin := user.IsAdmin(c)

	// Only allow scheduled cron jobs or authenticated administrators to
	// create the table. The X-AppEngine-Cron header is set by App Engine
	// when a cron task invokes the handler.
	if (r.Header.Get("X-AppEngine-Cron") != "true") && (isAdmin == false) {
		common.Error("Handler called without admin/cron privilege")
		http.Error(w, "Handler called without admin/cron privilege", http.StatusBadRequest)
		return
	}

	today := time.Now().Format("20060102")
	err := createVisitsTableInBigQuery(c, today)
	if err != nil {
		common.Error("Error while creating table %v: %v", today, err)
		http.Error(w, "Error while creating today table: "+err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "Table %v created", today)
}

func CreateTomorrowVisitsTableInBigQueryHandler(w http.ResponseWriter, r *http.Request) {
	c := r.Context()
	common.Info(">>>>>>>> CreateTomorrowVisitsTableInBigQueryHandler")

	isAdmin := user.IsAdmin(c)

	// Protected endpoint: only cron or admin users may create tomorrow's table.
	// App Engine sets the X-AppEngine-Cron header for scheduled tasks.
	if (r.Header.Get("X-AppEngine-Cron") != "true") && (isAdmin == false) {
		common.Error("Handler called without admin/cron privilege")
		http.Error(w, "Handler called without admin/cron privilege", http.StatusBadRequest)
		return
	}

	tomorrow := time.Now().Add(time.Hour*23 + time.Minute*59).Format("20060102")
	err := createVisitsTableInBigQuery(c, tomorrow)
	if err != nil {
		common.Error("Error while creating table %v: %v", tomorrow, err)
		http.Error(w, "Error while creating tomorrow table: "+err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "Table %v created", tomorrow)
}

func CreateTodayEventsTableInBigQueryHandler(w http.ResponseWriter, r *http.Request) {
	c := r.Context()
	common.Info(">>>>>>>> CreateTomorrowEventsTableInBigQueryHandler")

	isAdmin := user.IsAdmin(c)

	// Only accessible to cron jobs or admin users to prevent unauthorized
	// creation of event tables.
	if (r.Header.Get("X-AppEngine-Cron") != "true") && (isAdmin == false) {
		common.Error("Handler called without admin/cron privilege")
		http.Error(w, "Handler called without admin/cron privilege", http.StatusBadRequest)
		return
	}

	today := time.Now().Format("20060102")
	err := createEventsTableInBigQuery(c, today)
	if err != nil {
		common.Error("Error while creating table %v: %v", today, err)
		http.Error(w, "Error while creating today table: "+err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "Table %v created", today)
}

func CreateTomorrowEventsTableInBigQueryHandler(w http.ResponseWriter, r *http.Request) {
	c := r.Context()
	common.Info(">>>>>>>> CreateTomorrowEventsTableInBigQueryHandler")

	isAdmin := user.IsAdmin(c)

	// Only cron or admin users are permitted to create tomorrow's events table.
	if (r.Header.Get("X-AppEngine-Cron") != "true") && (isAdmin == false) {
		common.Error("Handler called without admin/cron privilege")
		http.Error(w, "Handler called without admin/cron privilege", http.StatusBadRequest)
		return
	}

	tomorrow := time.Now().Add(time.Hour*23 + time.Minute*59).Format("20060102")
	err := createEventsTableInBigQuery(c, tomorrow)
	if err != nil {
		common.Error("Error while creating table %v: %v", tomorrow, err)
		http.Error(w, "Error while creating tomorrow table: "+err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "Table %v created", tomorrow)
}

func TrackHandler(w http.ResponseWriter, r *http.Request) {
	common.Info(">>>>>>>> TrackHandler")

	common.Info("c=%v a=%v l=%v v=%v", r.FormValue("c"), r.FormValue("a"), r.FormValue("l"), r.FormValue("v"))
	TrackEvent(w, r, common.GetCookieID(w, r))
	// The pixel response must look like an image and must not be cached by
	// the browser. A permissive CORS header allows the pixel to be embedded
	// from any origin.
	w.Header().Set("Content-Type", "image/png")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Add("Access-Control-Allow-Origin", "*")

	// onePixelPNG is a 1x1 transparent PNG defined in base.go. Writing it
	// triggers the image load that records the tracking event.
	w.Write([]byte(onePixelPNG))
}

func ClickHandler(w http.ResponseWriter, r *http.Request) {
	common.Info(">>>>>>>> ClickHandler")

	common.Info("c=%v a=%v l=%v v=%v", r.FormValue("c"), r.FormValue("a"), r.FormValue("l"), r.FormValue("v"))
	TrackEvent(w, r, common.GetCookieID(w, r))
	url := r.FormValue("url")
	// Validate the destination to avoid redirecting to arbitrary schemes.
	// Fallback to the site homepage when the URL is empty or invalid.
	if !common.IsValidHTTPURL(url) {
		common.Error("Invalid redirect URL: %v", url)
		url = "http://www.mygotome.com"
	}
	common.Info("Redirect to %v", url)
	http.Redirect(w, r, url, http.StatusFound)
}
