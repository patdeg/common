package track

// This file contains helper functions used to stream data to BigQuery while
// automatically creating tables that may not yet exist. The core logic retries
// the insert once after attempting to create the table when BigQuery returns a
// "not found" error.

import (
	"github.com/patdeg/common"
	"github.com/patdeg/common/gcp"

	"golang.org/x/net/context"
	bigquery "google.golang.org/api/bigquery/v2"
	"google.golang.org/api/googleapi"
)

// streamDataFn is the function used to stream rows into BigQuery. It is a
// variable so tests can replace it with a stub implementation.
var streamDataFn = gcp.StreamDataInBigquery

// insertWithTableCreation streams data to BigQuery and creates the table if it
// does not exist. When the initial insert returns a 404 error, the provided
// createTable callback is invoked to ensure the dataset and table exist before
// retrying the insert.
func insertWithTableCreation(c context.Context, projectID, datasetID, tableID string, req *bigquery.TableDataInsertAllRequest, createTable func(context.Context, string) error) error {
	common.Debug("insertWithTableCreation dataset=%s table=%s", datasetID, tableID)
	err := streamDataFn(c, projectID, datasetID, tableID, req)

	if err != nil {
		common.Error("Error while streaming data to BigQuery: %v", err)
		// gerr is of type *googleapi.Error when the BigQuery API returns
		// a structured error. We specifically look for HTTP 404 to know
		// the table does not exist.
		if gerr, ok := err.(*googleapi.Error); ok && gerr.Code == 404 {
			common.Debug("BigQuery returned 404, attempting to create dataset/table")
			common.Info("BigQuery table %s not found, creating table and retrying", tableID)
			if err2 := createTable(c, tableID); err2 != nil {
				common.Error("Error creating table %s: %v", tableID, err2)
				return err2
			}

			// Retry the insert after creating the table.
			if err3 := streamDataFn(c, projectID, datasetID, tableID, req); err3 != nil {
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
