package track

import (
	"context"
	"errors"
	"testing"

	bigquery "google.golang.org/api/bigquery/v2"
	"google.golang.org/api/googleapi"
)

// stubStreamer simulates StreamDataInBigquery behaviour.
type stubStreamer struct {
	errs  []error
	calls int
}

func (s *stubStreamer) Stream(c context.Context, projectID, datasetID, tableID string, req *bigquery.TableDataInsertAllRequest) error {
	if s.calls >= len(s.errs) {
		s.calls++
		return nil
	}
	err := s.errs[s.calls]
	s.calls++
	return err
}

// stubCreator records invocations of table creation.
type stubCreator struct {
	err   error
	calls int
}

func (s *stubCreator) Create(ctx context.Context, table string) error {
	s.calls++
	return s.err
}

func TestInsertWithTableCreationSuccess(t *testing.T) {
	streamer := &stubStreamer{errs: []error{nil}}
	creator := &stubCreator{}
	old := streamDataFn
	streamDataFn = streamer.Stream
	defer func() { streamDataFn = old }()

	req := &bigquery.TableDataInsertAllRequest{}
	err := insertWithTableCreation(context.Background(), "p", "d", "t", req, creator.Create)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if streamer.calls != 1 {
		t.Errorf("streamer called %d times, want 1", streamer.calls)
	}
	if creator.calls != 0 {
		t.Errorf("creator called %d times, want 0", creator.calls)
	}
}

func TestInsertWithTableCreationTableMissing(t *testing.T) {
	gerr := &googleapi.Error{Code: 404}
	streamer := &stubStreamer{errs: []error{gerr, nil}}
	creator := &stubCreator{}
	old := streamDataFn
	streamDataFn = streamer.Stream
	defer func() { streamDataFn = old }()

	req := &bigquery.TableDataInsertAllRequest{}
	err := insertWithTableCreation(context.Background(), "p", "d", "t", req, creator.Create)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if streamer.calls != 2 {
		t.Errorf("streamer called %d times, want 2", streamer.calls)
	}
	if creator.calls != 1 {
		t.Errorf("creator called %d times, want 1", creator.calls)
	}
}

func TestInsertWithTableCreationOtherError(t *testing.T) {
	streamer := &stubStreamer{errs: []error{errors.New("bad")}}
	creator := &stubCreator{}
	old := streamDataFn
	streamDataFn = streamer.Stream
	defer func() { streamDataFn = old }()

	req := &bigquery.TableDataInsertAllRequest{}
	err := insertWithTableCreation(context.Background(), "p", "d", "t", req, creator.Create)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if streamer.calls != 1 {
		t.Errorf("streamer called %d times, want 1", streamer.calls)
	}
	if creator.calls != 0 {
		t.Errorf("creator called %d times, want 0", creator.calls)
	}
}
