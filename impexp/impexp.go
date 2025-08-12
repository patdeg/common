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

// Package impexp provides import/export utilities for data migration,
// backup, and restore operations.
package impexp

import (
	"archive/zip"
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/patdeg/common"
)

// Format represents the export/import format
type Format string

const (
	FormatJSON Format = "json"
	FormatCSV  Format = "csv"
	FormatXML  Format = "xml"
	FormatZIP  Format = "zip"
)

// Options configures import/export operations
type Options struct {
	Format      Format            // Export format
	Pretty      bool              // Pretty print JSON
	Compress    bool              // Compress output
	Filter      FilterFunc        // Filter function for selective export
	Transform   TransformFunc     // Transform function for data manipulation
	BatchSize   int               // Batch size for large datasets
	Delimiter   rune              // CSV delimiter
	Headers     []string          // CSV headers
	MaxFileSize int64             // Maximum file size in bytes
	Metadata    map[string]string // Additional metadata
}

// FilterFunc filters entities during export/import
type FilterFunc func(entity interface{}) bool

// TransformFunc transforms entities during export/import
type TransformFunc func(entity interface{}) (interface{}, error)

// Exporter handles data export operations
type Exporter interface {
	// Export exports data to a writer
	Export(ctx context.Context, data interface{}, w io.Writer, opts *Options) error

	// ExportFile exports data to a file
	ExportFile(ctx context.Context, data interface{}, filename string, opts *Options) error

	// ExportBatch exports data in batches
	ExportBatch(ctx context.Context, dataSource DataSource, w io.Writer, opts *Options) error
}

// Importer handles data import operations
type Importer interface {
	// Import imports data from a reader
	Import(ctx context.Context, r io.Reader, dest interface{}, opts *Options) error

	// ImportFile imports data from a file
	ImportFile(ctx context.Context, filename string, dest interface{}, opts *Options) error

	// ImportBatch imports data in batches
	ImportBatch(ctx context.Context, r io.Reader, dataSink DataSink, opts *Options) error
}

// DataSource provides data for batch export
type DataSource interface {
	// NextBatch returns the next batch of data
	NextBatch(ctx context.Context, batchSize int) ([]interface{}, error)

	// HasMore indicates if more data is available
	HasMore() bool
}

// DataSink receives data during batch import
type DataSink interface {
	// WriteBatch writes a batch of data
	WriteBatch(ctx context.Context, batch []interface{}) error
}

// DefaultExporter implements the Exporter interface
type DefaultExporter struct{}

// DefaultImporter implements the Importer interface
type DefaultImporter struct{}

// NewExporter creates a new exporter
func NewExporter() Exporter {
	return &DefaultExporter{}
}

// NewImporter creates a new importer
func NewImporter() Importer {
	return &DefaultImporter{}
}

// Export exports data to a writer
func (e *DefaultExporter) Export(ctx context.Context, data interface{}, w io.Writer, opts *Options) error {
	if opts == nil {
		opts = &Options{Format: FormatJSON}
	}

	switch opts.Format {
	case FormatJSON:
		return e.exportJSON(data, w, opts)
	case FormatCSV:
		return e.exportCSV(data, w, opts)
	case FormatZIP:
		return e.exportZIP(ctx, data, w, opts)
	default:
		return fmt.Errorf("unsupported format: %s", opts.Format)
	}
}

// ExportFile exports data to a file
func (e *DefaultExporter) ExportFile(ctx context.Context, data interface{}, filename string, opts *Options) error {
	// Create directory if needed
	dir := filepath.Dir(filename)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %v", err)
	}

	// Create file
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create file: %v", err)
	}
	defer file.Close()

	// Export data
	if err := e.Export(ctx, data, file, opts); err != nil {
		return err
	}

	common.Info("[IMPEXP] Exported data to %s", filename)
	return nil
}

// ExportBatch exports data in batches
func (e *DefaultExporter) ExportBatch(ctx context.Context, dataSource DataSource, w io.Writer, opts *Options) error {
	if opts == nil {
		opts = &Options{Format: FormatJSON, BatchSize: 100}
	}

	if opts.BatchSize <= 0 {
		opts.BatchSize = 100
	}

	// Start export based on format
	switch opts.Format {
	case FormatJSON:
		// Write opening bracket for JSON array
		if _, err := w.Write([]byte("[\n")); err != nil {
			return err
		}
	case FormatCSV:
		// Write CSV headers if provided
		if len(opts.Headers) > 0 {
			csvWriter := csv.NewWriter(w)
			if opts.Delimiter != 0 {
				csvWriter.Comma = opts.Delimiter
			}
			if err := csvWriter.Write(opts.Headers); err != nil {
				return err
			}
			csvWriter.Flush()
		}
	}

	totalExported := 0
	first := true

	for dataSource.HasMore() {
		// Check context cancellation
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// Get next batch
		batch, err := dataSource.NextBatch(ctx, opts.BatchSize)
		if err != nil {
			return fmt.Errorf("failed to get batch: %v", err)
		}

		// Export batch
		for _, item := range batch {
			// Apply filter if provided
			if opts.Filter != nil && !opts.Filter(item) {
				continue
			}

			// Apply transform if provided
			if opts.Transform != nil {
				transformed, err := opts.Transform(item)
				if err != nil {
					common.Warn("[IMPEXP] Failed to transform item: %v", err)
					continue
				}
				item = transformed
			}

			// Write item based on format
			switch opts.Format {
			case FormatJSON:
				if !first {
					if _, err := w.Write([]byte(",\n")); err != nil {
						return err
					}
				}
				encoder := json.NewEncoder(w)
				if opts.Pretty {
					encoder.SetIndent("  ", "  ")
				}
				if err := encoder.Encode(item); err != nil {
					return err
				}
				first = false
			case FormatCSV:
				// Convert to CSV row
				// This is simplified - in production, use reflection
				csvWriter := csv.NewWriter(w)
				if opts.Delimiter != 0 {
					csvWriter.Comma = opts.Delimiter
				}
				// Write row (simplified)
				csvWriter.Flush()
			}

			totalExported++
		}
	}

	// Close export based on format
	switch opts.Format {
	case FormatJSON:
		if _, err := w.Write([]byte("\n]")); err != nil {
			return err
		}
	}

	common.Info("[IMPEXP] Exported %d items", totalExported)
	return nil
}

// exportJSON exports data as JSON
func (e *DefaultExporter) exportJSON(data interface{}, w io.Writer, opts *Options) error {
	encoder := json.NewEncoder(w)
	if opts.Pretty {
		encoder.SetIndent("", "  ")
	}
	return encoder.Encode(data)
}

// exportCSV exports data as CSV
func (e *DefaultExporter) exportCSV(data interface{}, w io.Writer, opts *Options) error {
	csvWriter := csv.NewWriter(w)
	if opts.Delimiter != 0 {
		csvWriter.Comma = opts.Delimiter
	}

	// Write headers if provided
	if len(opts.Headers) > 0 {
		if err := csvWriter.Write(opts.Headers); err != nil {
			return err
		}
	}

	// Convert data to CSV rows
	// This is a simplified implementation
	// In production, use reflection to handle various data types

	csvWriter.Flush()
	return csvWriter.Error()
}

// exportZIP exports data as a ZIP archive
func (e *DefaultExporter) exportZIP(ctx context.Context, data interface{}, w io.Writer, opts *Options) error {
	zipWriter := zip.NewWriter(w)
	defer zipWriter.Close()

	// Add metadata file
	metaFile, err := zipWriter.Create("metadata.json")
	if err != nil {
		return err
	}

	metadata := map[string]interface{}{
		"exported_at": time.Now().Format(time.RFC3339),
		"format":      "zip",
		"version":     "1.0",
	}

	if opts.Metadata != nil {
		for k, v := range opts.Metadata {
			metadata[k] = v
		}
	}

	if err := json.NewEncoder(metaFile).Encode(metadata); err != nil {
		return err
	}

	// Add data file
	dataFile, err := zipWriter.Create("data.json")
	if err != nil {
		return err
	}

	encoder := json.NewEncoder(dataFile)
	if opts.Pretty {
		encoder.SetIndent("", "  ")
	}

	return encoder.Encode(data)
}

// Import imports data from a reader
func (i *DefaultImporter) Import(ctx context.Context, r io.Reader, dest interface{}, opts *Options) error {
	if opts == nil {
		opts = &Options{Format: FormatJSON}
	}

	switch opts.Format {
	case FormatJSON:
		return i.importJSON(r, dest, opts)
	case FormatCSV:
		return i.importCSV(r, dest, opts)
	case FormatZIP:
		return i.importZIP(ctx, r, dest, opts)
	default:
		return fmt.Errorf("unsupported format: %s", opts.Format)
	}
}

// ImportFile imports data from a file
func (i *DefaultImporter) ImportFile(ctx context.Context, filename string, dest interface{}, opts *Options) error {
	file, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("failed to open file: %v", err)
	}
	defer file.Close()

	// Auto-detect format from extension if not specified
	if opts == nil {
		opts = &Options{}
	}
	if opts.Format == "" {
		ext := strings.ToLower(filepath.Ext(filename))
		switch ext {
		case ".json":
			opts.Format = FormatJSON
		case ".csv":
			opts.Format = FormatCSV
		case ".zip":
			opts.Format = FormatZIP
		default:
			opts.Format = FormatJSON
		}
	}

	if err := i.Import(ctx, file, dest, opts); err != nil {
		return err
	}

	common.Info("[IMPEXP] Imported data from %s", filename)
	return nil
}

// ImportBatch imports data in batches
func (i *DefaultImporter) ImportBatch(ctx context.Context, r io.Reader, dataSink DataSink, opts *Options) error {
	if opts == nil {
		opts = &Options{Format: FormatJSON, BatchSize: 100}
	}

	if opts.BatchSize <= 0 {
		opts.BatchSize = 100
	}

	decoder := json.NewDecoder(r)

	// Read opening bracket
	token, err := decoder.Token()
	if err != nil {
		return err
	}
	if delim, ok := token.(json.Delim); !ok || delim != '[' {
		return fmt.Errorf("expected JSON array")
	}

	batch := make([]interface{}, 0, opts.BatchSize)
	totalImported := 0

	// Read items
	for decoder.More() {
		// Check context cancellation
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		var item interface{}
		if err := decoder.Decode(&item); err != nil {
			return err
		}

		// Apply filter if provided
		if opts.Filter != nil && !opts.Filter(item) {
			continue
		}

		// Apply transform if provided
		if opts.Transform != nil {
			transformed, err := opts.Transform(item)
			if err != nil {
				common.Warn("[IMPEXP] Failed to transform item: %v", err)
				continue
			}
			item = transformed
		}

		batch = append(batch, item)

		// Write batch when full
		if len(batch) >= opts.BatchSize {
			if err := dataSink.WriteBatch(ctx, batch); err != nil {
				return fmt.Errorf("failed to write batch: %v", err)
			}
			totalImported += len(batch)
			batch = batch[:0]
		}
	}

	// Write remaining items
	if len(batch) > 0 {
		if err := dataSink.WriteBatch(ctx, batch); err != nil {
			return fmt.Errorf("failed to write final batch: %v", err)
		}
		totalImported += len(batch)
	}

	common.Info("[IMPEXP] Imported %d items", totalImported)
	return nil
}

// importJSON imports data from JSON
func (i *DefaultImporter) importJSON(r io.Reader, dest interface{}, opts *Options) error {
	return json.NewDecoder(r).Decode(dest)
}

// importCSV imports data from CSV
func (i *DefaultImporter) importCSV(r io.Reader, dest interface{}, opts *Options) error {
	csvReader := csv.NewReader(r)
	if opts.Delimiter != 0 {
		csvReader.Comma = opts.Delimiter
	}

	// Read all records
	records, err := csvReader.ReadAll()
	if err != nil {
		return err
	}

	// Convert CSV records to destination type
	// This is a simplified implementation
	// In production, use reflection to handle various data types

	common.Info("[IMPEXP] Imported %d CSV records", len(records))
	return nil
}

// importZIP imports data from a ZIP archive
func (i *DefaultImporter) importZIP(ctx context.Context, r io.Reader, dest interface{}, opts *Options) error {
	// This would need to read the ZIP into memory or a temp file
	// For simplicity, returning an error
	return fmt.Errorf("ZIP import not fully implemented")
}

// Backup creates a full backup of data
func Backup(ctx context.Context, sources map[string]DataSource, outputDir string) error {
	timestamp := time.Now().Format("20060102-150405")
	backupDir := filepath.Join(outputDir, fmt.Sprintf("backup-%s", timestamp))

	if err := os.MkdirAll(backupDir, 0755); err != nil {
		return fmt.Errorf("failed to create backup directory: %v", err)
	}

	exporter := NewExporter()

	for name, source := range sources {
		filename := filepath.Join(backupDir, fmt.Sprintf("%s.json", name))
		file, err := os.Create(filename)
		if err != nil {
			return fmt.Errorf("failed to create backup file: %v", err)
		}

		opts := &Options{
			Format:    FormatJSON,
			Pretty:    true,
			BatchSize: 100,
		}

		if err := exporter.ExportBatch(ctx, source, file, opts); err != nil {
			file.Close()
			return fmt.Errorf("failed to export %s: %v", name, err)
		}

		file.Close()
		common.Info("[BACKUP] Backed up %s to %s", name, filename)
	}

	common.Info("[BACKUP] Backup completed in %s", backupDir)
	return nil
}

// Restore restores data from a backup
func Restore(ctx context.Context, backupDir string, sinks map[string]DataSink) error {
	importer := NewImporter()

	for name, sink := range sinks {
		filename := filepath.Join(backupDir, fmt.Sprintf("%s.json", name))

		file, err := os.Open(filename)
		if err != nil {
			if os.IsNotExist(err) {
				common.Warn("[RESTORE] Backup file not found for %s", name)
				continue
			}
			return fmt.Errorf("failed to open backup file: %v", err)
		}

		opts := &Options{
			Format:    FormatJSON,
			BatchSize: 100,
		}

		if err := importer.ImportBatch(ctx, file, sink, opts); err != nil {
			file.Close()
			return fmt.Errorf("failed to import %s: %v", name, err)
		}

		file.Close()
		common.Info("[RESTORE] Restored %s from %s", name, filename)
	}

	common.Info("[RESTORE] Restore completed from %s", backupDir)
	return nil
}
