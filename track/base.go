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

// Package track defines helpers used for pixel tracking and BigQuery storage.
//
// Dataset variables hold the BigQuery project and dataset names for visits,
// events and AdWords tracking. Each variable defaults to a sensible name but
// can be overridden via the corresponding environment variable, allowing the
// runtime configuration to differ from the source defaults.
//
// onePixelPNG contains a transparent 1×1 PNG used as the response for tracking
// requests.
package track

// Dataset and project IDs are read from environment variables when available,
// otherwise the defaults below are used.
var (
	bqProjectID      = getEnv("BQ_PROJECT_ID", "mygotome")
	visitsDataset    = getEnv("VISITS_DATASET", "visits")
	eventsDataset    = getEnv("EVENTS_DATASET", "events")
	adwordsProjectID = getEnv("ADWORDS_PROJECT_ID", "mygotome")
	adwordsDataset   = getEnv("ADWORDS_DATASET", "adwords")
)

const onePixelPNG = "\x89\x50\x4e\x47\x0d\x0a\x1a\x0a\x00\x00\x00\x0d\x49\x48" +
	"\x44\x52\x00\x00\x00\x01\x00\x00\x00\x01\x08\x02\x00\x00\x00\x90\x77\x53" +
	"\xde\x00\x00\x00\x01\x73\x52\x47\x42\x00\xae\xce\x1c\xe9\x00\x00\x00\x04" +
	"\x67\x41\x4d\x41\x00\x00\xb1\x8f\x0b\xfc\x61\x05\x00\x00\x00\x09\x70\x48" +
	"\x59\x73\x00\x00\x0e\xc3\x00\x00\x0e\xc3\x01\xc7\x6f\xa8\x64\x00\x00\x00" +
	"\x0c\x49\x44\x41\x54\x18\x57\x63\xf8\xff\xff\x3f\x00\x05\xfe\x02\xfe\xa7" +
	"\x35\x81\x84\x00\x00\x00\x00\x49\x45\x4e\x44\xae\x42\x60\x82"
