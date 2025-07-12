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
	streamDataFn = func(context.Context, string, string, string, *bigquery.TableDataInsertAllRequest) error {
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
	streamDataFn = func(context.Context, string, string, string, *bigquery.TableDataInsertAllRequest) error {
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
	streamDataFn = func(context.Context, string, string, string, *bigquery.TableDataInsertAllRequest) error {
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
