package track

import (
	"errors"

	"github.com/patdeg/common"
	"github.com/patdeg/common/gcp"

	"golang.org/x/net/context"
	bigquery "google.golang.org/api/bigquery/v2"
)

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
				{Name: "Cookie", Type: "STRING", Description: "Cookie"},
				{Name: "Session", Type: "STRING", Description: "Session"},
				{Name: "URI", Type: "STRING", Description: "URI"},
				{Name: "Referer", Type: "STRING", Description: "Referer"},
				{Name: "Time", Type: "TIMESTAMP", Description: "Time"},
				{Name: "Host", Type: "STRING", Description: "Host"},
				{Name: "RemoteAddr", Type: "STRING", Description: "RemoteAddr"},
				{Name: "InstanceId", Type: "STRING", Description: "InstanceId"},
				{Name: "VersionId", Type: "STRING", Description: "VersionId"},
				{Name: "Scheme", Type: "STRING", Description: "Scheme"},
				{Name: "Country", Type: "STRING", Description: "Country"},
				{Name: "Region", Type: "STRING", Description: "Region"},
				{Name: "City", Type: "STRING", Description: "City"},
				{Name: "Lat", Type: "FLOAT", Description: "City latitude"},
				{Name: "Lon", Type: "FLOAT", Description: "City longitude"},
				{Name: "AcceptLanguage", Type: "STRING", Description: "AcceptLanguage"},
				{Name: "UserAgent", Type: "STRING", Description: "UserAgent"},
				{Name: "IsMobile", Type: "BOOLEAN", Description: "IsMobile"},
				{Name: "IsBot", Type: "BOOLEAN", Description: "IsBot"},
				{Name: "MozillaVersion", Type: "STRING", Description: "MozillaVersion"},
				{Name: "Platform", Type: "STRING", Description: "Platform"},
				{Name: "OS", Type: "STRING", Description: "OS"},
				{Name: "EngineName", Type: "STRING", Description: "EngineName"},
				{Name: "EngineVersion", Type: "STRING", Description: "EngineVersion"},
				{Name: "BrowserName", Type: "STRING", Description: "BrowserName"},
				{Name: "BrowserVersion", Type: "STRING", Description: "BrowserVersion"},
			},
		},
	}
	return gcp.CreateTableInBigQuery(c, newTable)
}

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
				{Name: "Cookie", Type: "STRING", Description: "Cookie"},
				{Name: "Session", Type: "STRING", Description: "Session"},
				{Name: "Category", Type: "STRING", Description: "Session"},
				{Name: "Action", Type: "STRING", Description: "Action"},
				{Name: "Label", Type: "STRING", Description: "Label"},
				{Name: "Value", Type: "FLOAT", Description: "Value"},
				{Name: "URI", Type: "STRING", Description: "URI"},
				{Name: "Referer", Type: "STRING", Description: "Referer"},
				{Name: "Time", Type: "TIMESTAMP", Description: "Time"},
				{Name: "Host", Type: "STRING", Description: "Host"},
				{Name: "RemoteAddr", Type: "STRING", Description: "RemoteAddr"},
				{Name: "InstanceId", Type: "STRING", Description: "InstanceId"},
				{Name: "VersionId", Type: "STRING", Description: "VersionId"},
				{Name: "Scheme", Type: "STRING", Description: "Scheme"},
				{Name: "Country", Type: "STRING", Description: "Country"},
				{Name: "Region", Type: "STRING", Description: "Region"},
				{Name: "City", Type: "STRING", Description: "City"},
				{Name: "Lat", Type: "FLOAT", Description: "City latitude"},
				{Name: "Lon", Type: "FLOAT", Description: "City longitude"},
				{Name: "AcceptLanguage", Type: "STRING", Description: "AcceptLanguage"},
				{Name: "UserAgent", Type: "STRING", Description: "UserAgent"},
				{Name: "IsMobile", Type: "BOOLEAN", Description: "IsMobile"},
				{Name: "IsBot", Type: "BOOLEAN", Description: "IsBot"},
				{Name: "MozillaVersion", Type: "STRING", Description: "MozillaVersion"},
				{Name: "Platform", Type: "STRING", Description: "Platform"},
				{Name: "OS", Type: "STRING", Description: "OS"},
				{Name: "EngineName", Type: "STRING", Description: "EngineName"},
				{Name: "EngineVersion", Type: "STRING", Description: "EngineVersion"},
				{Name: "BrowserName", Type: "STRING", Description: "BrowserName"},
				{Name: "BrowserVersion", Type: "STRING", Description: "BrowserVersion"},
			},
		},
	}
	return gcp.CreateTableInBigQuery(c, newTable)
}
