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

// Package gcp provides Google Cloud helpers for mygoto.me service.
//
// This file contains helpers for interacting with BigQuery. It includes
// functions to authenticate using the default service account, create
// datasets and tables when they do not exist and stream rows using the
// BigQuery streaming API. All operations log errors via the common logging
// helpers and streaming inserts are retried once when failures occur.
package gcp

import (
	"encoding/json"
	"errors"
	"time"

	"golang.org/x/net/context"
	"golang.org/x/oauth2/google"
	bigquery "google.golang.org/api/bigquery/v2"
	"google.golang.org/api/googleapi"
)

// GetBQServiceAccountClient returns an authenticated BigQuery service using the
// default service account credentials. Errors encountered while creating the
// client are logged and returned to the caller.
func GetBQServiceAccountClient(c context.Context) (*bigquery.Service, error) {
	httpClient, err := google.DefaultClient(c,
		"https://www.googleapis.com/auth/userinfo.email",
		"https://www.googleapis.com/auth/bigquery",
	)
	if err != nil {
		Error("Error creating default HTTP client for BigQuery: %v", err)
		return nil, err
	}
	return bigquery.New(httpClient)
}

// CreateDatasetIfNotExists ensures the given dataset exists. It first attempts
// to retrieve the dataset and if it receives a 404 error, a new dataset is
// created. All errors are logged and returned to the caller.
func CreateDatasetIfNotExists(c context.Context, projectID, datasetID string) error {
	svc, err := GetBQServiceAccountClient(c)
	if err != nil {
		return err
	}

	// Check whether the dataset already exists.
	_, err = bigquery.NewDatasetsService(svc).Get(projectID, datasetID).Do()
	if err == nil {
		Debug("Dataset %s already exists", datasetID)
		return nil
	}

	// A 404 means the dataset does not exist yet, so create it now.
	if gerr, ok := err.(*googleapi.Error); ok && gerr.Code == 404 {
		Info("Dataset %s not found, creating", datasetID)
		ds := &bigquery.Dataset{DatasetReference: &bigquery.DatasetReference{ProjectId: projectID, DatasetId: datasetID}}
		_, err = bigquery.NewDatasetsService(svc).Insert(projectID, ds).Do()
		if err != nil {
			Error("Error creating dataset %s: %v", datasetID, err)
			return err
		}
		return nil
	}
	Error("Error getting dataset %s: %v", datasetID, err)
	return err
}

// CreateTableInBigQuery deletes any existing table with the same ID and then
// creates the provided table definition. Deletion errors are logged but do not
// stop table creation. The caller is expected to provide a table with a valid
// TableReference and Schema.
func CreateTableInBigQuery(c context.Context, newTable *bigquery.Table) error {

	if newTable == nil {
		return errors.New("No newTable defined for CreateTableInBigQuery")
	}

	if newTable.TableReference == nil {
		return errors.New("No newTable.TableReference defined for CreateTableInBigQuery")
	}

	if newTable.Schema == nil {
		return errors.New("No newTable.Schema defined for CreateTableInBigQuery")
	}

	bqServiceAccountService, err := GetBQServiceAccountClient(c)
	if err != nil {
		Error("Error getting BigQuery Service: %v", err)
		return err
	}

	// Attempt to delete any existing table with the same ID to start fresh.
	err = bigquery.
		NewTablesService(bqServiceAccountService).
		Delete(
			newTable.TableReference.ProjectId,
			newTable.TableReference.DatasetId,
			newTable.TableReference.TableId).
		Do()
	if err != nil {
		Info("There was an error while trying to delete old snapshot table: %v", err)
	}

	// Create the table using the supplied definition.
	_, err = bigquery.
		NewTablesService(bqServiceAccountService).
		Insert(
			newTable.TableReference.ProjectId,
			newTable.TableReference.DatasetId,
			newTable).
		Do()

	return err
}

// StreamDataInBigquery inserts rows into a BigQuery table using the streaming
// API. If the first attempt fails, the function waits 10 seconds and retries
// once. Errors from each attempt are logged and the error from the second
// attempt is returned.
func StreamDataInBigquery(c context.Context, projectId, datasetId, tableId string, req *bigquery.TableDataInsertAllRequest) error {
	Debug("[STREAM_BQ] Starting StreamDataInBigquery")
	Debug("[STREAM_BQ] Parameters: project=%s dataset=%s table=%s", projectId, datasetId, tableId)

	if req == nil {
		Error("[STREAM_BQ] Request is nil!")
		return errors.New("No req defined for StreamDataInBigquery")
	}

	Debug("[STREAM_BQ] Request Kind=%s NumRows=%d", req.Kind, len(req.Rows))
	for i, row := range req.Rows {
		Debug("[STREAM_BQ] Row[%d] InsertId=%s", i, row.InsertId)
		Debug("[STREAM_BQ] Row[%d] Json has %d fields", i, len(row.Json))
		for k, v := range row.Json {
			Debug("[STREAM_BQ] Row[%d] Json[%s] type=%T", i, k, v)
			// For maps, log the nested structure
			if m, ok := v.(map[string]interface{}); ok {
				Debug("[STREAM_BQ] Row[%d] Json[%s] is a map with %d keys", i, k, len(m))
				for mk, mv := range m {
					Debug("[STREAM_BQ] Row[%d] Json[%s][%s] type=%T value=%v", i, k, mk, mv, mv)
				}
			} else {
				Debug("[STREAM_BQ] Row[%d] Json[%s] value=%v", i, k, v)
			}
		}
	}

	// Serialize the request for comprehensive logging
	if reqJSON, err := json.Marshal(req); err == nil {
		Debug("[STREAM_BQ] Full request JSON: %s", string(reqJSON))
	} else {
		Error("[STREAM_BQ] Failed to marshal request for debug: %v", err)
	}

	Debug("[STREAM_BQ] Getting BigQuery service client...")
	bqServiceAccountService, err := GetBQServiceAccountClient(c)
	if err != nil {
		Error("[STREAM_BQ] Error getting BigQuery Service: %v", err)
		return err
	}
	Debug("[STREAM_BQ] BigQuery service client obtained successfully")

	Debug("[STREAM_BQ] Calling InsertAll API...")
	resp, err := bigquery.
		NewTabledataService(bqServiceAccountService).
		InsertAll(projectId, datasetId, tableId, req).
		Do()
	if err != nil {
		Info("[STREAM_BQ] First attempt failed: %v", err)
		Info("[STREAM_BQ] Error type: %T", err)
		Info("[STREAM_BQ] Will retry in 10 seconds...")
		time.Sleep(time.Second * 10)
		Debug("[STREAM_BQ] Retrying InsertAll API...")
		resp, err = bigquery.
			NewTabledataService(bqServiceAccountService).
			InsertAll(projectId, datasetId, tableId, req).
			Do()
		if err != nil {
			Error("[STREAM_BQ] Second attempt to stream data to BigQuery failed: %v", err)
			Error("[STREAM_BQ] Second attempt error type: %T", err)
			return err
		} else {
			Info("[STREAM_BQ] 2nd try was successful")
		}
	} else {
		Debug("[STREAM_BQ] InsertAll API call succeeded")
	}

	Debug("[STREAM_BQ] Checking response for insert errors...")
	Debug("[STREAM_BQ] Response has %d InsertErrors entries", len(resp.InsertErrors))

	isError := false
	for i, insertError := range resp.InsertErrors {
		if insertError != nil {
			Debug("[STREAM_BQ] InsertError[%d] Index=%d, has %d errors", i, insertError.Index, len(insertError.Errors))
			for j, e := range insertError.Errors {
				Debug("[STREAM_BQ] InsertError[%d].Errors[%d]: Reason=%s Message=%s Location=%s DebugInfo=%s", i, j, e.Reason, e.Message, e.Location, e.DebugInfo)
				if (e.DebugInfo != "") || (e.Message != "") || (e.Reason != "") {
					Error("[STREAM_BQ] BigQuery error %v: %v at %v/%v", e.Reason, e.Message, i, j)
					Error("[STREAM_BQ] Error location: %s", e.Location)
					Error("[STREAM_BQ] Error debugInfo: %s", e.DebugInfo)
					isError = true
				}
			}
		}
	}

	if isError {
		Error("[STREAM_BQ] Returning error due to insert errors")
		return errors.New("There was an error streaming data to Big Query")
	}

	Debug("[STREAM_BQ] StreamDataInBigquery completed successfully")
	return nil

}
