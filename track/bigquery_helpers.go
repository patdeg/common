package track

import (
	"github.com/patdeg/common"
	"github.com/patdeg/common/gcp"

	"golang.org/x/net/context"
	bigquery "google.golang.org/api/bigquery/v2"
	"google.golang.org/api/googleapi"
)

// streamData points to gcp.StreamDataInBigquery and can be replaced in tests.
var streamData = gcp.StreamDataInBigquery

// insertWithTableCreation streams data to BigQuery and creates the table if it doesn't exist.
func insertWithTableCreation(c context.Context, projectID, datasetID, tableID string, req *bigquery.TableDataInsertAllRequest, createTable func(context.Context, string) error) error {
	common.Debug("insertWithTableCreation dataset=%s table=%s", datasetID, tableID)
	err := streamData(c, projectID, datasetID, tableID, req)
	if err != nil {
		common.Error("Error while streaming data to BigQuery: %v", err)
		if gerr, ok := err.(*googleapi.Error); ok && gerr.Code == 404 {
			common.Debug("BigQuery returned 404, attempting to create dataset/table")
			common.Info("BigQuery table %s not found, creating table and retrying", tableID)
			if err2 := createTable(c, tableID); err2 != nil {
				common.Error("Error creating table %s: %v", tableID, err2)
				return err2
			}
			if err3 := streamData(c, projectID, datasetID, tableID, req); err3 != nil {
				common.Error("Error streaming to BigQuery after creating table: %v", err3)
				return err3
			}
			common.Debug("Insert after table creation succeeded")
			return nil
		}
		return err
	}
	common.Debug("insertWithTableCreation succeeded without creating table")
	return nil
}
