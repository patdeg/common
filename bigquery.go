// Package common provides shared helpers used across the mygoto.me service.
package common

import (
	"errors"
	"time"

	"golang.org/x/net/context"
	"golang.org/x/oauth2/google"
	bigquery "google.golang.org/api/bigquery/v2"
	"google.golang.org/api/googleapi"
)

// GetBQServiceAccountClient returns an authenticated BigQuery service using the
// default service account credentials.
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

// CreateDatasetIfNotExists ensures the given dataset exists.
// If the dataset is not found, it will be created.
func CreateDatasetIfNotExists(c context.Context, projectID, datasetID string) error {
	svc, err := GetBQServiceAccountClient(c)
	if err != nil {
		return err
	}

	_, err = bigquery.NewDatasetsService(svc).Get(projectID, datasetID).Do()
	if err == nil {
		Debug("Dataset %s already exists", datasetID)
		return nil
	}

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
// creates the provided table definition.
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
// API and retries once if the initial attempt fails.
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
		Info("Error streaming data to Big Query, trying again in 10 seconds: %v", err)
		time.Sleep(time.Second * 10)
		resp, err = bigquery.
			NewTabledataService(bqServiceAccountService).
			InsertAll(projectId, datasetId, tableId, req).
			Do()
		if err != nil {
			Error("Error again streaming data to Big Query: %v", err)
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
