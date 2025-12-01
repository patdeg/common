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
	common.Debug("[INSERT_WITH_TABLE] Starting insertWithTableCreation")
	common.Debug("[INSERT_WITH_TABLE] Parameters: project=%s dataset=%s table=%s", projectID, datasetID, tableID)
	common.Debug("[INSERT_WITH_TABLE] Request Kind=%s NumRows=%d", req.Kind, len(req.Rows))

	// Log request details for debugging
	for i, row := range req.Rows {
		common.Debug("[INSERT_WITH_TABLE] Row[%d] InsertId=%s", i, row.InsertId)
		common.Debug("[INSERT_WITH_TABLE] Row[%d] Json fields: %d", i, len(row.Json))
		for k, v := range row.Json {
			common.Debug("[INSERT_WITH_TABLE] Row[%d] Json[%s] type=%T value=%v", i, k, v, v)
		}
	}

	common.Debug("[INSERT_WITH_TABLE] Calling streamDataFn...")
	err := streamDataFn(c, projectID, datasetID, tableID, req)

	if err != nil {
		common.Error("[INSERT_WITH_TABLE] streamDataFn returned error: %v", err)
		common.Error("[INSERT_WITH_TABLE] Error type: %T", err)

		// gerr is of type *googleapi.Error when the BigQuery API returns
		// a structured error. We specifically look for HTTP 404 to know
		// the table does not exist.
		if gerr, ok := err.(*googleapi.Error); ok {
			common.Debug("[INSERT_WITH_TABLE] Error is googleapi.Error, Code=%d Message=%s", gerr.Code, gerr.Message)
			common.Debug("[INSERT_WITH_TABLE] googleapi.Error Body=%s", gerr.Body)
			for i, e := range gerr.Errors {
				common.Debug("[INSERT_WITH_TABLE] googleapi.Error.Errors[%d] Reason=%s Message=%s", i, e.Reason, e.Message)
			}

			if gerr.Code == 404 {
				common.Debug("[INSERT_WITH_TABLE] BigQuery returned 404, attempting to create dataset/table")
				common.Info("[INSERT_WITH_TABLE] BigQuery table %s not found, creating table and retrying", tableID)
				if err2 := createTable(c, tableID); err2 != nil {
					common.Error("[INSERT_WITH_TABLE] Error creating table %s: %v", tableID, err2)
					return err2
				}
				common.Debug("[INSERT_WITH_TABLE] Table creation succeeded, retrying insert...")

				// Retry the insert after creating the table.
				if err3 := streamDataFn(c, projectID, datasetID, tableID, req); err3 != nil {
					common.Error("[INSERT_WITH_TABLE] Error streaming to BigQuery after creating table: %v", err3)
					common.Error("[INSERT_WITH_TABLE] Retry error type: %T", err3)
					return err3
				}
				common.Debug("[INSERT_WITH_TABLE] Insert after table creation succeeded")
				return nil
			}
		} else {
			common.Debug("[INSERT_WITH_TABLE] Error is NOT googleapi.Error, type=%T", err)
		}
		common.Error("[INSERT_WITH_TABLE] Returning error from insertWithTableCreation")
		return err
	}
	common.Debug("[INSERT_WITH_TABLE] streamDataFn succeeded without error")
	common.Debug("[INSERT_WITH_TABLE] insertWithTableCreation completed successfully")
	return nil
}
