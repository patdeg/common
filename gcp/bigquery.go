// Package gcp provides Google Cloud helpers for mygoto.me service.
//
// This file contains helpers for interacting with BigQuery. It includes
// functions to authenticate using the default service account, create
// datasets and tables when they do not exist and stream rows using the
// BigQuery streaming API. All operations log errors via the common logging
// helpers and streaming inserts are retried once when failures occur.
package gcp

import (
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

	if req == nil {
		return errors.New("No req defined for StreamDataInBigquery")
	}

	bqServiceAccountService, err := GetBQServiceAccountClient(c)
	if err != nil {
		Error("Error getting BigQuery Service: %v", err)
		return err
	}

	resp, err := bigquery.
		NewTabledataService(bqServiceAccountService).
		InsertAll(projectId, datasetId, tableId, req).
		Do()
	if err != nil {
		Info("Error streaming data to BigQuery, will retry in 10 seconds: %v", err)
		time.Sleep(time.Second * 10)
		resp, err = bigquery.
			NewTabledataService(bqServiceAccountService).
			InsertAll(projectId, datasetId, tableId, req).
			Do()
		if err != nil {
			Error("Second attempt to stream data to BigQuery failed: %v", err)
			return err
		} else {
			Info("2nd try was successful")
		}
	}

	isError := false
	for i, insertError := range resp.InsertErrors {
		if insertError != nil {
			for j, e := range insertError.Errors {
				if (e.DebugInfo != "") || (e.Message != "") || (e.Reason != "") {
					Error("BigQuery error %v: %v at %v/%v", e.Reason, e.Message, i, j)
					isError = true
				}
			}
		}
	}

	if isError {
		return errors.New("There was an error streaming data to Big Query")
	}

	return nil

}
