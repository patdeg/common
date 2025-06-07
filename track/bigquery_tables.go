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

package track

import (
	"errors"

	"github.com/patdeg/common"
	"github.com/patdeg/common/gcp"

	"golang.org/x/net/context"
	bigquery "google.golang.org/api/bigquery/v2"
)

// createVisitsTableInBigQuery creates a daily visits table in BigQuery using the
// provided date string `d` (format YYYYMMDD). It is typically invoked by cron
// handlers such as CreateTodayVisitsTableInBigQueryHandler and
// CreateTomorrowVisitsTableInBigQueryHandler before Visit rows are streamed.
func createVisitsTableInBigQuery(c context.Context, d string) error {
	common.Info(">>>> createVisitsTableInBigQuery")

	if err := gcp.CreateDatasetIfNotExists(c, bqProjectID, visitsDataset); err != nil {
		common.Error("Error ensuring dataset %s: %v", visitsDataset, err)
		return err
	}

	if len(d) != 8 {
		return errors.New("table name is badly formatted - expected 8 characters")
	}
	newTable := &bigquery.Table{
		TableReference: &bigquery.TableReference{
			ProjectId: bqProjectID,
			DatasetId: visitsDataset,
			TableId:   d,
		},
		FriendlyName: "Daily Visits table",
		Description:  "This table is created automatically to store daily visits to Deglon Consulting properties ",
		Schema: &bigquery.TableSchema{
			Fields: []*bigquery.TableFieldSchema{
				{Name: "Cookie", Type: "STRING", Description: "Cookie"},                 // Visitor cookie ID
				{Name: "Session", Type: "STRING", Description: "Session"},               // Random session identifier
				{Name: "URI", Type: "STRING", Description: "URI"},                       // Request URI
				{Name: "Referer", Type: "STRING", Description: "Referer"},               // HTTP referer
				{Name: "Time", Type: "TIMESTAMP", Description: "Time"},                  // Timestamp of the visit
				{Name: "Host", Type: "STRING", Description: "Host"},                     // HTTP host header
				{Name: "RemoteAddr", Type: "STRING", Description: "RemoteAddr"},         // Client IP address
				{Name: "InstanceId", Type: "STRING", Description: "InstanceId"},         // App Engine instance ID
				{Name: "VersionId", Type: "STRING", Description: "VersionId"},           // App Engine version ID
				{Name: "Scheme", Type: "STRING", Description: "Scheme"},                 // HTTP scheme (http/https)
				{Name: "Country", Type: "STRING", Description: "Country"},               // Country derived from IP
				{Name: "Region", Type: "STRING", Description: "Region"},                 // Region derived from IP
				{Name: "City", Type: "STRING", Description: "City"},                     // City derived from IP
				{Name: "Lat", Type: "FLOAT", Description: "City latitude"},              // City latitude
				{Name: "Lon", Type: "FLOAT", Description: "City longitude"},             // City longitude
				{Name: "AcceptLanguage", Type: "STRING", Description: "AcceptLanguage"}, // Accept-Language header
				{Name: "UserAgent", Type: "STRING", Description: "UserAgent"},           // User-Agent header
				{Name: "IsMobile", Type: "BOOLEAN", Description: "IsMobile"},            // true if UA is mobile
				{Name: "IsBot", Type: "BOOLEAN", Description: "IsBot"},                  // true if request is from a bot
				{Name: "MozillaVersion", Type: "STRING", Description: "MozillaVersion"}, // Mozilla version from UA
				{Name: "Platform", Type: "STRING", Description: "Platform"},             // Platform extracted from UA
				{Name: "OS", Type: "STRING", Description: "OS"},                         // Operating system
				{Name: "EngineName", Type: "STRING", Description: "EngineName"},         // Rendering engine name
				{Name: "EngineVersion", Type: "STRING", Description: "EngineVersion"},   // Rendering engine version
				{Name: "BrowserName", Type: "STRING", Description: "BrowserName"},       // Browser name
				{Name: "BrowserVersion", Type: "STRING", Description: "BrowserVersion"}, // Browser version
			},
		},
	}
	return gcp.CreateTableInBigQuery(c, newTable)
}

// createEventsTableInBigQuery creates a daily events table in BigQuery using the
// provided date string `d` (format YYYYMMDD). It is used by cron handlers
// CreateTodayEventsTableInBigQueryHandler and
// CreateTomorrowEventsTableInBigQueryHandler before Event rows are streamed.
func createEventsTableInBigQuery(c context.Context, d string) error {
	common.Info(">>>> createEventsTableInBigQuery")

	if err := gcp.CreateDatasetIfNotExists(c, bqProjectID, eventsDataset); err != nil {
		common.Error("Error ensuring dataset %s: %v", eventsDataset, err)
		return err
	}

	if len(d) != 8 {
		return errors.New("table name is badly formatted - expected 8 characters")
	}
	newTable := &bigquery.Table{
		TableReference: &bigquery.TableReference{
			ProjectId: bqProjectID,
			DatasetId: eventsDataset,
			TableId:   d,
		},
		FriendlyName: "Daily Visits table",
		Description:  "This table is created automatically to store daily visits to Deglon Consulting properties ",
		Schema: &bigquery.TableSchema{
			Fields: []*bigquery.TableFieldSchema{
				{Name: "Cookie", Type: "STRING", Description: "Cookie"},                 // Visitor cookie ID
				{Name: "Session", Type: "STRING", Description: "Session"},               // Random session identifier
				{Name: "Category", Type: "STRING", Description: "Session"},              // Event category
				{Name: "Action", Type: "STRING", Description: "Action"},                 // Event action
				{Name: "Label", Type: "STRING", Description: "Label"},                   // Event label
				{Name: "Value", Type: "FLOAT", Description: "Value"},                    // Event value
				{Name: "URI", Type: "STRING", Description: "URI"},                       // Request URI
				{Name: "Referer", Type: "STRING", Description: "Referer"},               // HTTP referer
				{Name: "Time", Type: "TIMESTAMP", Description: "Time"},                  // Timestamp of the event
				{Name: "Host", Type: "STRING", Description: "Host"},                     // HTTP host header
				{Name: "RemoteAddr", Type: "STRING", Description: "RemoteAddr"},         // Client IP address
				{Name: "InstanceId", Type: "STRING", Description: "InstanceId"},         // App Engine instance ID
				{Name: "VersionId", Type: "STRING", Description: "VersionId"},           // App Engine version ID
				{Name: "Scheme", Type: "STRING", Description: "Scheme"},                 // HTTP scheme (http/https)
				{Name: "Country", Type: "STRING", Description: "Country"},               // Country derived from IP
				{Name: "Region", Type: "STRING", Description: "Region"},                 // Region derived from IP
				{Name: "City", Type: "STRING", Description: "City"},                     // City derived from IP
				{Name: "Lat", Type: "FLOAT", Description: "City latitude"},              // City latitude
				{Name: "Lon", Type: "FLOAT", Description: "City longitude"},             // City longitude
				{Name: "AcceptLanguage", Type: "STRING", Description: "AcceptLanguage"}, // Accept-Language header
				{Name: "UserAgent", Type: "STRING", Description: "UserAgent"},           // User-Agent header
				{Name: "IsMobile", Type: "BOOLEAN", Description: "IsMobile"},            // true if UA is mobile
				{Name: "IsBot", Type: "BOOLEAN", Description: "IsBot"},                  // true if request is from a bot
				{Name: "MozillaVersion", Type: "STRING", Description: "MozillaVersion"}, // Mozilla version from UA
				{Name: "Platform", Type: "STRING", Description: "Platform"},             // Platform extracted from UA
				{Name: "OS", Type: "STRING", Description: "OS"},                         // Operating system
				{Name: "EngineName", Type: "STRING", Description: "EngineName"},         // Rendering engine name
				{Name: "EngineVersion", Type: "STRING", Description: "EngineVersion"},   // Rendering engine version
				{Name: "BrowserName", Type: "STRING", Description: "BrowserName"},       // Browser name
				{Name: "BrowserVersion", Type: "STRING", Description: "BrowserVersion"}, // Browser version
			},
		},
	}
	return gcp.CreateTableInBigQuery(c, newTable)
}
