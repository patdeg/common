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
	"reflect"
	"testing"
	"time"

	bigquery "google.golang.org/api/bigquery/v2"
)

func TestVisitInsertRequest(t *testing.T) {
	now := time.Unix(0, 123456789)
	visit := &Visit{
		Cookie:         "c",
		Session:        "s",
		URI:            "/foo",
		Referer:        "http://example.com",
		Time:           time.Date(2020, 1, 2, 3, 4, 5, 0, time.UTC),
		Host:           "example.com",
		RemoteAddr:     "192.168.0.1",
		InstanceId:     "iid",
		VersionId:      "v1",
		Scheme:         "https",
		Country:        "US",
		Region:         "CA",
		City:           "SF",
		Lat:            1.2,
		Lon:            3.4,
		AcceptLanguage: "en-US",
		UserAgent:      "agent",
		IsMobile:       false,
		IsBot:          false,
		MozillaVersion: "5.0",
		Platform:       "linux",
		OS:             "Linux",
		EngineName:     "webkit",
		EngineVersion:  "1",
		BrowserName:    "chrome",
		BrowserVersion: "100",
	}

	got := visitInsertRequest(visit, now)

	want := &bigquery.TableDataInsertAllRequest{
		Kind: "bigquery#tableDataInsertAllRequest",
		Rows: []*bigquery.TableDataInsertAllRequestRows{
			{
				InsertId: "123456789-c",
				Json: map[string]bigquery.JsonValue{
					"Cookie":         "c",
					"Session":        "s",
					"URI":            "/foo",
					"Referer":        "http://example.com",
					"Time":           visit.Time,
					"Host":           "example.com",
					"RemoteAddr":     "192.168.0.1",
					"InstanceId":     "iid",
					"VersionId":      "v1",
					"Scheme":         "https",
					"Country":        "US",
					"Region":         "CA",
					"City":           "SF",
					"Lat":            1.2,
					"Lon":            3.4,
					"AcceptLanguage": "en-US",
					"UserAgent":      "agent",
					"IsMobile":       false,
					"IsBot":          false,
					"MozillaVersion": "5.0",
					"Platform":       "linux",
					"OS":             "Linux",
					"EngineName":     "webkit",
					"EngineVersion":  "1",
					"BrowserName":    "chrome",
					"BrowserVersion": "100",
				},
			},
		},
	}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("visitInsertRequest mismatch\n got %#v\nwant %#v", got, want)
	}
}
