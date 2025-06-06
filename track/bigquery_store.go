package track

// This file contains helpers to persist Visit and Event records in BigQuery.
// Both types are stored in daily tables created on demand. The data is
// streamed using bigquery.TableDataInsertAllRequest built from a Visit
// structure.

import (
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
func StoreVisitInBigQuery(c context.Context, v *Visit) error {
	common.Info(">>>> StoreVisitInBigQuery")
	common.Debug("Dataset=%s", visitsDataset)

	insertId := strconv.FormatInt(time.Now().UnixNano(), 10) + "-" + v.Cookie

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
