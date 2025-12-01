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

// This file contains helpers to persist Visit and Event records in BigQuery.
// Both types are stored in daily tables created on demand. The data is
// streamed using bigquery.TableDataInsertAllRequest built from a Visit
// structure.

import (
	"encoding/json"
	"strconv"
	"time"

	"github.com/patdeg/common"

	"golang.org/x/net/context"
	bigquery "google.golang.org/api/bigquery/v2"
)

// visitInsertRequest builds the BigQuery request used by StoreVisitInBigQuery.
// The insertId combines the current timestamp in nanoseconds with the visitor
// cookie to ensure uniqueness and allow de-duplication on retries.
// Each Visit field is mapped directly to a column in BigQuery.
func visitInsertRequest(v *Visit, now time.Time) *bigquery.TableDataInsertAllRequest {
	insertId := strconv.FormatInt(now.UnixNano(), 10) + "-" + v.Cookie

	req := &bigquery.TableDataInsertAllRequest{
		Kind: "bigquery#tableDataInsertAllRequest",
		Rows: []*bigquery.TableDataInsertAllRequestRows{
			{
				InsertId: insertId,
				Json: map[string]bigquery.JsonValue{
					"Cookie":         v.Cookie,         // visitor cookie ID
					"Session":        v.Session,        // session ID cached in memcache
					"URI":            v.URI,            // request URI
					"Referer":        v.Referer,        // HTTP referer
					"Time":           v.Time,           // time of the visit
					"Host":           v.Host,           // request host
					"RemoteAddr":     v.RemoteAddr,     // remote address
					"InstanceId":     v.InstanceId,     // App Engine instance ID
					"VersionId":      v.VersionId,      // deployed version ID
					"Scheme":         v.Scheme,         // http or https
					"Country":        v.Country,        // geo country
					"Region":         v.Region,         // geo region
					"City":           v.City,           // geo city
					"Lat":            v.Lat,            // geo latitude
					"Lon":            v.Lon,            // geo longitude
					"AcceptLanguage": v.AcceptLanguage, // browser Accept-Language header
					"UserAgent":      v.UserAgent,      // raw User-Agent string
					"IsMobile":       v.IsMobile,       // true for mobile browsers
					"IsBot":          v.IsBot,          // true if the UA is a bot
					"MozillaVersion": v.MozillaVersion, // UA Mozilla version
					"Platform":       v.Platform,       // UA reported platform
					"OS":             v.OS,             // UA reported OS
					"EngineName":     v.EngineName,     // rendering engine name
					"EngineVersion":  v.EngineVersion,  // rendering engine version
					"BrowserName":    v.BrowserName,    // browser name
					"BrowserVersion": v.BrowserVersion, // browser version
				},
			},
		},
	}

	return req
}

// eventInsertRequest extends visitInsertRequest with event specific fields
// (Category, Action, Label and Value).
func eventInsertRequest(v *Visit, now time.Time) *bigquery.TableDataInsertAllRequest {
	req := visitInsertRequest(v, now)
	for _, row := range req.Rows {
		row.Json["Category"] = v.Category
		row.Json["Action"] = v.Action
		row.Json["Label"] = v.Label
		row.Json["Value"] = v.Value
	}
	return req
}

// touchPointInsertRequest builds the BigQuery request used by
// StoreTouchPointInBigQuery. The insertId combines the current timestamp in
// nanoseconds with the RemoteAddr to provide a reasonably unique identifier
// while still allowing BigQuery to de-duplicate retried inserts.
//
// DUAL-COLUMN PATTERN FOR PAYLOAD DATA
// ====================================
// The touchpoints table uses a dual-column pattern for payload data:
//
// 1. PayloadString (STRING) - Used for INGESTION
//    - Raw JSON string, always succeeds with streaming insert
//    - Populated automatically by this function
//    - Never causes insert errors
//
// 2. Payload (JSON) - Used for QUERIES
//    - BigQuery native JSON type for dot-notation queries
//    - Populated MANUALLY via SQL UPDATE (not by this function)
//    - Enables queries like: SELECT Payload.utm_source FROM touchpoints
//
// WHY THIS PATTERN?
// -----------------
// BigQuery's streaming insert API (v2) has issues with JSON column types.
// Passing Go maps or JSON strings to JSON columns causes "not a record" errors.
// Using a STRING column for ingestion is 100% reliable.
//
// MANUAL CONVERSION (run as needed):
// ----------------------------------
//     UPDATE `demeterics.touchpoints.touchpoints`
//     SET Payload = SAFE.PARSE_JSON(PayloadString)
//     WHERE Payload IS NULL
//       AND PayloadString IS NOT NULL
//       AND PayloadString != '{}'
//       AND _PARTITIONTIME >= TIMESTAMP_SUB(CURRENT_TIMESTAMP(), INTERVAL 7 DAY)
//
// QUERYING BEFORE CONVERSION:
// ---------------------------
// You can query PayloadString directly using JSON functions:
//     SELECT JSON_VALUE(PayloadString, '$.utm_source') as utm_source
//     FROM touchpoints
//
// See docs/TOUCHPOINTS_PAYLOAD.md in this repository for full documentation.
func touchPointInsertRequest(tp *TouchPointEvent, now time.Time) *bigquery.TableDataInsertAllRequest {
	common.Debug("[TOUCHPOINT_INSERT] Starting touchPointInsertRequest")
	common.Debug("[TOUCHPOINT_INSERT] Input TouchPointEvent: Time=%v Category=%s Action=%s Label=%s", tp.Time, tp.Category, tp.Action, tp.Label)
	common.Debug("[TOUCHPOINT_INSERT] Input TouchPointEvent: Path=%s Host=%s RemoteAddr=%s", tp.Path, tp.Host, tp.RemoteAddr)
	common.Debug("[TOUCHPOINT_INSERT] Input TouchPointEvent: Referer=%s UserAgent=%s", tp.Referer, tp.UserAgent)
	common.Debug("[TOUCHPOINT_INSERT] Input TouchPointEvent: PayloadJSON length=%d", len(tp.PayloadJSON))
	if tp.PayloadJSON != "" {
		common.Debug("[TOUCHPOINT_INSERT] Input TouchPointEvent: PayloadJSON=%s", tp.PayloadJSON)
	}

	insertId := strconv.FormatInt(now.UnixNano(), 10) + "-" + tp.RemoteAddr
	common.Debug("[TOUCHPOINT_INSERT] Generated insertId=%s", insertId)

	// For BigQuery JSON columns via the streaming insert API (v2), we pass the
	// JSON as a string. The API will parse it into the JSON column type.
	// Using a Go map directly doesn't work - it causes "not a record" errors.
	payloadJSONStr := tp.PayloadJSON
	if payloadJSONStr == "" {
		payloadJSONStr = "{}"
	} else {
		// Validate it's proper JSON
		var testParse map[string]interface{}
		if err := json.Unmarshal([]byte(payloadJSONStr), &testParse); err != nil {
			common.Warn("[TOUCHPOINT_INSERT] PayloadJSON is not valid JSON: %v", err)
			common.Warn("[TOUCHPOINT_INSERT] PayloadJSON content: %s", payloadJSONStr)
			payloadJSONStr = "{}"
		}
	}
	common.Debug("[TOUCHPOINT_INSERT] Payload JSON string: %s", payloadJSONStr)

	req := &bigquery.TableDataInsertAllRequest{
		Kind: "bigquery#tableDataInsertAllRequest",
		Rows: []*bigquery.TableDataInsertAllRequestRows{
			{
				InsertId: insertId,
				Json: map[string]bigquery.JsonValue{
					"Time":       tp.Time,
					"Category":   tp.Category,
					"Action":     tp.Action,
					"Label":      tp.Label,
					"Referer":    tp.Referer,
					"Path":       tp.Path,
					"Host":       tp.Host,
					"RemoteAddr": tp.RemoteAddr,
					"UserAgent":  tp.UserAgent,
					// PayloadString for reliable streaming insert (STRING column)
					// See docs/TOUCHPOINTS_PAYLOAD.md for dual-column pattern
					"PayloadString": payloadJSONStr,
				},
			},
		},
	}

	common.Debug("[TOUCHPOINT_INSERT] Built TableDataInsertAllRequest with Kind=%s", req.Kind)
	common.Debug("[TOUCHPOINT_INSERT] Request has %d rows", len(req.Rows))
	if len(req.Rows) > 0 {
		row := req.Rows[0]
		common.Debug("[TOUCHPOINT_INSERT] Row[0] InsertId=%s", row.InsertId)
		common.Debug("[TOUCHPOINT_INSERT] Row[0] Json has %d fields", len(row.Json))
		for k, v := range row.Json {
			common.Debug("[TOUCHPOINT_INSERT] Row[0] Json[%s] = %v (type: %T)", k, v, v)
		}
	}

	// Serialize the entire request for comprehensive debugging
	if reqJSON, err := json.Marshal(req); err == nil {
		common.Debug("[TOUCHPOINT_INSERT] Full request as JSON: %s", string(reqJSON))
	} else {
		common.Error("[TOUCHPOINT_INSERT] Failed to marshal request for debug: %v", err)
	}

	common.Debug("[TOUCHPOINT_INSERT] Completed touchPointInsertRequest")
	return req
}

func StoreVisitInBigQuery(c context.Context, v *Visit) error {
	common.Info(">>>> StoreVisitInBigQuery")
	common.Debug("Dataset=%s", visitsDataset)

	// build the streaming request with a unique insertId
	req := visitInsertRequest(v, time.Now())

	tableName := time.Now().Format("20060102")
	common.Debug("Table=%s", tableName)

	return insertWithTableCreation(c, bqProjectID, visitsDataset, tableName, req, createVisitsTableInBigQuery)
}

// StoreEventInBigQuery streams an Event visit to BigQuery. The dataset and
// table are automatically created if necessary and the insert retried once.
func StoreEventInBigQuery(c context.Context, v *Visit) error {
	common.Info(">>>> StoreEventInBigQuery")
	common.Debug("Dataset=%s", eventsDataset)

	// build the streaming request including event specific fields
	req := eventInsertRequest(v, time.Now())

	tableName := time.Now().Format("20060102")
	common.Debug("Table=%s", tableName)

	return insertWithTableCreation(c, bqProjectID, eventsDataset, tableName, req, createEventsTableInBigQuery)
}

// StoreTouchPointInBigQuery streams a TouchPointEvent to BigQuery. The dataset
// and partitioned table are created on demand if they do not already exist.
// The table is partitioned by day on the Time field.
func StoreTouchPointInBigQuery(c context.Context, e *TouchPointEvent) error {
	common.Info("[TOUCHPOINT_STORE] >>>> StoreTouchPointInBigQuery starting")
	common.Debug("[TOUCHPOINT_STORE] Dataset=%s Project=%s", touchpointsDataset, touchpointsProjectID)
	common.Debug("[TOUCHPOINT_STORE] TouchPointEvent: Category=%s Action=%s Label=%s", e.Category, e.Action, e.Label)
	common.Debug("[TOUCHPOINT_STORE] TouchPointEvent: Time=%v Host=%s Path=%s", e.Time, e.Host, e.Path)
	common.Debug("[TOUCHPOINT_STORE] TouchPointEvent: RemoteAddr=%s Referer=%s", e.RemoteAddr, e.Referer)
	common.Debug("[TOUCHPOINT_STORE] TouchPointEvent: UserAgent length=%d PayloadJSON length=%d", len(e.UserAgent), len(e.PayloadJSON))

	now := time.Now()
	common.Debug("[TOUCHPOINT_STORE] Current time for insert: %v", now)

	req := touchPointInsertRequest(e, now)
	common.Debug("[TOUCHPOINT_STORE] touchPointInsertRequest returned, request built")

	tableName := "touchpoints"
	common.Debug("[TOUCHPOINT_STORE] Target table=%s", tableName)

	// Create a wrapper function that matches the expected signature
	createTableFunc := func(ctx context.Context, _ string) error {
		common.Debug("[TOUCHPOINT_STORE] createTableFunc called, invoking createTouchpointsTableInBigQuery")
		return createTouchpointsTableInBigQuery(ctx)
	}

	common.Debug("[TOUCHPOINT_STORE] Calling insertWithTableCreation with project=%s dataset=%s table=%s", touchpointsProjectID, touchpointsDataset, tableName)
	err := insertWithTableCreation(c, touchpointsProjectID, touchpointsDataset, tableName, req, createTableFunc)
	if err != nil {
		common.Error("[TOUCHPOINT_STORE] insertWithTableCreation failed: %v", err)
		common.Error("[TOUCHPOINT_STORE] Failed event details: Category=%s Action=%s Label=%s Host=%s Path=%s", e.Category, e.Action, e.Label, e.Host, e.Path)
	} else {
		common.Info("[TOUCHPOINT_STORE] StoreTouchPointInBigQuery completed successfully")
	}
	return err
}
