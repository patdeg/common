package track

import (
	"errors"
	"testing"

	"golang.org/x/net/context"
	bigquery "google.golang.org/api/bigquery/v2"
	"google.golang.org/api/googleapi"
)

// TestInsertWithTableCreation404 ensures that a 404 error triggers table creation and a retry.
func TestInsertWithTableCreation404(t *testing.T) {
	ctx := context.Background()
	called := 0
	streamData = func(context.Context, string, string, string, *bigquery.TableDataInsertAllRequest) error {
		called++
		if called == 1 {
			return &googleapi.Error{Code: 404}
		}
		return nil
	}
	created := 0
	create := func(context.Context, string) error { created++; return nil }
	err := insertWithTableCreation(ctx, "p", "d", "t", &bigquery.TableDataInsertAllRequest{}, create)
	if err != nil {
		t.Fatalf("insertWithTableCreation returned error: %v", err)
	}
	if called != 2 {
		t.Errorf("streamData called %d times, want 2", called)
	}
	if created != 1 {
		t.Errorf("createTable called %d times, want 1", created)
	}
}

// TestInsertWithTableCreationError ensures non-404 errors are returned.
func TestInsertWithTableCreationError(t *testing.T) {
	ctx := context.Background()
	streamData = func(context.Context, string, string, string, *bigquery.TableDataInsertAllRequest) error {
		return errors.New("fail")
	}
	create := func(context.Context, string) error { t.Error("createTable should not be called"); return nil }
	if err := insertWithTableCreation(ctx, "p", "d", "t", &bigquery.TableDataInsertAllRequest{}, create); err == nil {
		t.Fatal("expected error")
	}
}

// TestInsertWithTableCreationCreateError ensures errors from createTable propagate.
func TestInsertWithTableCreationCreateError(t *testing.T) {
	ctx := context.Background()
	called := 0
	streamData = func(context.Context, string, string, string, *bigquery.TableDataInsertAllRequest) error {
		called++
		return &googleapi.Error{Code: 404}
	}
	create := func(context.Context, string) error { return errors.New("create fail") }
	if err := insertWithTableCreation(ctx, "p", "d", "t", &bigquery.TableDataInsertAllRequest{}, create); err == nil {
		t.Fatal("expected error")
	}
	if called != 1 {
		t.Errorf("streamData called %d times, want 1", called)
	}
}
