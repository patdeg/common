package track

import (
	"strconv"
	"time"

	"github.com/patdeg/common"

	"golang.org/x/net/context"
	bigquery "google.golang.org/api/bigquery/v2"
)

// StoreVisitInBigQuery streams a Visit to BigQuery. If the target table does
// not exist, it will be created and the insert retried once.
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
					"Cookie":         v.Cookie,
					"Session":        v.Session,
					"URI":            v.URI,
					"Referer":        v.Referer,
					"Time":           v.Time,
					"Host":           v.Host,
					"RemoteAddr":     v.RemoteAddr,
					"InstanceId":     v.InstanceId,
					"VersionId":      v.VersionId,
					"Scheme":         v.Scheme,
					"Country":        v.Country,
					"Region":         v.Region,
					"City":           v.City,
					"Lat":            v.Lat,
					"Lon":            v.Lon,
					"AcceptLanguage": v.AcceptLanguage,
					"UserAgent":      v.UserAgent,
					"IsMobile":       v.IsMobile,
					"IsBot":          v.IsBot,
					"MozillaVersion": v.MozillaVersion,
					"Platform":       v.Platform,
					"OS":             v.OS,
					"EngineName":     v.EngineName,
					"EngineVersion":  v.EngineVersion,
					"BrowserName":    v.BrowserName,
					"BrowserVersion": v.BrowserVersion,
				},
			},
		},
	}

	tableName := time.Now().Format("20060102")
	common.Debug("Table=%s", tableName)

	return insertWithTableCreation(c, bqProjectID, visitsDataset, tableName, req, createVisitsTableInBigQuery)
}

// StoreEventInBigQuery streams an Event visit to BigQuery. The dataset and
// table are automatically created if necessary and the insert retried once.
func StoreEventInBigQuery(c context.Context, v *Visit) error {
	common.Info(">>>> StoreEventInBigQuery")
	common.Debug("Dataset=%s", eventsDataset)

	insertId := strconv.FormatInt(time.Now().UnixNano(), 10) + "-" + v.Cookie

	req := &bigquery.TableDataInsertAllRequest{
		Kind: "bigquery#tableDataInsertAllRequest",
		Rows: []*bigquery.TableDataInsertAllRequestRows{
			{
				InsertId: insertId,
				Json: map[string]bigquery.JsonValue{
					"Cookie":         v.Cookie,
					"Session":        v.Session,
					"URI":            v.URI,
					"Referer":        v.Referer,
					"Time":           v.Time,
					"Host":           v.Host,
					"RemoteAddr":     v.RemoteAddr,
					"InstanceId":     v.InstanceId,
					"VersionId":      v.VersionId,
					"Scheme":         v.Scheme,
					"Country":        v.Country,
					"Region":         v.Region,
					"City":           v.City,
					"Lat":            v.Lat,
					"Lon":            v.Lon,
					"AcceptLanguage": v.AcceptLanguage,
					"UserAgent":      v.UserAgent,
					"IsMobile":       v.IsMobile,
					"IsBot":          v.IsBot,
					"MozillaVersion": v.MozillaVersion,
					"Platform":       v.Platform,
					"OS":             v.OS,
					"EngineName":     v.EngineName,
					"EngineVersion":  v.EngineVersion,
					"BrowserName":    v.BrowserName,
					"BrowserVersion": v.BrowserVersion,
					"Category":       v.Category,
					"Action":         v.Action,
					"Label":          v.Label,
					"Value":          v.Value,
				},
			},
		},
	}

	tableName := time.Now().Format("20060102")
	common.Debug("Table=%s", tableName)

	return insertWithTableCreation(c, bqProjectID, eventsDataset, tableName, req, createEventsTableInBigQuery)
}
