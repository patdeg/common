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

// Package search provides full-text search capabilities with support
// for multiple backends including in-memory and external search services.
package search

import (
	"context"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/patdeg/common"
)

// Engine defines the search engine interface
type Engine interface {
	// Index adds or updates a document
	Index(ctx context.Context, doc Document) error

	// Search performs a search query
	Search(ctx context.Context, query Query) (*Results, error)

	// Delete removes a document
	Delete(ctx context.Context, id string) error

	// DeleteIndex removes all documents from an index
	DeleteIndex(ctx context.Context, index string) error

	// GetDocument retrieves a document by ID
	GetDocument(ctx context.Context, id string) (*Document, error)

	// UpdateDocument partially updates a document
	UpdateDocument(ctx context.Context, id string, updates map[string]interface{}) error
}

// Document represents a searchable document
type Document struct {
	ID        string                 `json:"id"`
	Index     string                 `json:"index"`
	Type      string                 `json:"type,omitempty"`
	Title     string                 `json:"title"`
	Content   string                 `json:"content"`
	Tags      []string               `json:"tags,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
	Score     float64                `json:"score,omitempty"`
}

// Query represents a search query
type Query struct {
	Text      string                 `json:"text"`
	Index     string                 `json:"index,omitempty"`
	Type      string                 `json:"type,omitempty"`
	Tags      []string               `json:"tags,omitempty"`
	Filters   map[string]interface{} `json:"filters,omitempty"`
	From      int                    `json:"from"`
	Size      int                    `json:"size"`
	Sort      []SortField            `json:"sort,omitempty"`
	Highlight bool                   `json:"highlight"`
	Facets    []string               `json:"facets,omitempty"`
}

// SortField defines sorting criteria
type SortField struct {
	Field string `json:"field"`
	Order string `json:"order"` // "asc" or "desc"
}

// Results represents search results
type Results struct {
	Total  int                    `json:"total"`
	Hits   []Document             `json:"hits"`
	Facets map[string][]FacetItem `json:"facets,omitempty"`
	Took   time.Duration          `json:"took"`
	Query  string                 `json:"query"`
}

// FacetItem represents a facet value and count
type FacetItem struct {
	Value string `json:"value"`
	Count int    `json:"count"`
}

// InMemoryEngine implements an in-memory search engine
type InMemoryEngine struct {
	documents map[string]*Document
	indices   map[string]map[string]*Document // index -> id -> document
	mu        sync.RWMutex
}

// NewInMemoryEngine creates a new in-memory search engine
func NewInMemoryEngine() *InMemoryEngine {
	return &InMemoryEngine{
		documents: make(map[string]*Document),
		indices:   make(map[string]map[string]*Document),
	}
}

// Index adds or updates a document
func (e *InMemoryEngine) Index(ctx context.Context, doc Document) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if doc.ID == "" {
		return fmt.Errorf("document ID is required")
	}

	if doc.Index == "" {
		doc.Index = "default"
	}

	if doc.Timestamp.IsZero() {
		doc.Timestamp = time.Now()
	}

	// Store document
	e.documents[doc.ID] = &doc

	// Add to index
	if e.indices[doc.Index] == nil {
		e.indices[doc.Index] = make(map[string]*Document)
	}
	e.indices[doc.Index][doc.ID] = &doc

	common.Debug("[SEARCH] Indexed document %s in index %s", doc.ID, doc.Index)
	return nil
}

// Search performs a search query
func (e *InMemoryEngine) Search(ctx context.Context, query Query) (*Results, error) {
	start := time.Now()

	e.mu.RLock()
	defer e.mu.RUnlock()

	// Get documents from specified index
	var searchDocs []*Document
	if query.Index != "" {
		if indexDocs, ok := e.indices[query.Index]; ok {
			for _, doc := range indexDocs {
				searchDocs = append(searchDocs, doc)
			}
		}
	} else {
		// Search all documents
		for _, doc := range e.documents {
			searchDocs = append(searchDocs, doc)
		}
	}

	// Filter by type
	if query.Type != "" {
		var filtered []*Document
		for _, doc := range searchDocs {
			if doc.Type == query.Type {
				filtered = append(filtered, doc)
			}
		}
		searchDocs = filtered
	}

	// Filter by tags
	if len(query.Tags) > 0 {
		var filtered []*Document
		for _, doc := range searchDocs {
			if hasAnyTag(doc.Tags, query.Tags) {
				filtered = append(filtered, doc)
			}
		}
		searchDocs = filtered
	}

	// Text search
	var results []Document
	if query.Text != "" {
		queryLower := strings.ToLower(query.Text)
		queryWords := strings.Fields(queryLower)

		for _, doc := range searchDocs {
			score := calculateScore(doc, queryWords)
			if score > 0 {
				docCopy := *doc
				docCopy.Score = score

				// Highlight matches if requested
				if query.Highlight {
					docCopy.Content = highlightMatches(docCopy.Content, queryWords)
					docCopy.Title = highlightMatches(docCopy.Title, queryWords)
				}

				results = append(results, docCopy)
			}
		}

		// Sort by score
		sort.Slice(results, func(i, j int) bool {
			return results[i].Score > results[j].Score
		})
	} else {
		// No text query - return all filtered documents
		for _, doc := range searchDocs {
			results = append(results, *doc)
		}
	}

	// Apply custom sorting
	if len(query.Sort) > 0 {
		applySorting(results, query.Sort)
	}

	// Calculate facets if requested
	var facets map[string][]FacetItem
	if len(query.Facets) > 0 {
		facets = calculateFacets(results, query.Facets)
	}

	// Pagination
	total := len(results)
	from := query.From
	if from < 0 {
		from = 0
	}

	size := query.Size
	if size <= 0 {
		size = 10
	}

	to := from + size
	if to > total {
		to = total
	}

	if from < total {
		results = results[from:to]
	} else {
		results = []Document{}
	}

	return &Results{
		Total:  total,
		Hits:   results,
		Facets: facets,
		Took:   time.Since(start),
		Query:  query.Text,
	}, nil
}

// Delete removes a document
func (e *InMemoryEngine) Delete(ctx context.Context, id string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	doc, ok := e.documents[id]
	if !ok {
		return fmt.Errorf("document not found: %s", id)
	}

	// Remove from index
	if indexDocs, ok := e.indices[doc.Index]; ok {
		delete(indexDocs, id)
	}

	// Remove from documents
	delete(e.documents, id)

	common.Debug("[SEARCH] Deleted document %s", id)
	return nil
}

// DeleteIndex removes all documents from an index
func (e *InMemoryEngine) DeleteIndex(ctx context.Context, index string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	// Get all documents in index
	indexDocs, ok := e.indices[index]
	if !ok {
		return nil // Index doesn't exist
	}

	// Remove documents
	for id := range indexDocs {
		delete(e.documents, id)
	}

	// Remove index
	delete(e.indices, index)

	common.Info("[SEARCH] Deleted index %s", index)
	return nil
}

// GetDocument retrieves a document by ID
func (e *InMemoryEngine) GetDocument(ctx context.Context, id string) (*Document, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	doc, ok := e.documents[id]
	if !ok {
		return nil, fmt.Errorf("document not found: %s", id)
	}

	return doc, nil
}

// UpdateDocument partially updates a document
func (e *InMemoryEngine) UpdateDocument(ctx context.Context, id string, updates map[string]interface{}) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	doc, ok := e.documents[id]
	if !ok {
		return fmt.Errorf("document not found: %s", id)
	}

	// Apply updates
	for key, value := range updates {
		switch key {
		case "title":
			if v, ok := value.(string); ok {
				doc.Title = v
			}
		case "content":
			if v, ok := value.(string); ok {
				doc.Content = v
			}
		case "tags":
			if v, ok := value.([]string); ok {
				doc.Tags = v
			}
		case "metadata":
			if v, ok := value.(map[string]interface{}); ok {
				doc.Metadata = v
			}
		}
	}

	doc.Timestamp = time.Now()

	common.Debug("[SEARCH] Updated document %s", id)
	return nil
}

// Helper functions

func hasAnyTag(docTags, queryTags []string) bool {
	tagMap := make(map[string]bool)
	for _, tag := range docTags {
		tagMap[tag] = true
	}

	for _, tag := range queryTags {
		if tagMap[tag] {
			return true
		}
	}

	return false
}

func calculateScore(doc *Document, queryWords []string) float64 {
	score := 0.0

	titleLower := strings.ToLower(doc.Title)
	contentLower := strings.ToLower(doc.Content)

	for _, word := range queryWords {
		// Title matches (weighted higher)
		titleCount := strings.Count(titleLower, word)
		score += float64(titleCount) * 2.0

		// Content matches
		contentCount := strings.Count(contentLower, word)
		score += float64(contentCount)

		// Tag matches
		for _, tag := range doc.Tags {
			if strings.Contains(strings.ToLower(tag), word) {
				score += 1.5
			}
		}
	}

	// Boost for exact phrase match
	fullQuery := strings.Join(queryWords, " ")
	if strings.Contains(titleLower, fullQuery) {
		score *= 2.0
	} else if strings.Contains(contentLower, fullQuery) {
		score *= 1.5
	}

	return score
}

func highlightMatches(text string, queryWords []string) string {
	result := text

	for _, word := range queryWords {
		// Create case-insensitive regex
		pattern := "(?i)" + regexp.QuoteMeta(word)
		re, err := regexp.Compile(pattern)
		if err != nil {
			continue
		}

		// Replace with highlighted version
		result = re.ReplaceAllString(result, "<mark>$0</mark>")
	}

	return result
}

func applySorting(results []Document, sortFields []SortField) {
	sort.Slice(results, func(i, j int) bool {
		for _, field := range sortFields {
			var cmp int

			switch field.Field {
			case "score":
				if results[i].Score < results[j].Score {
					cmp = -1
				} else if results[i].Score > results[j].Score {
					cmp = 1
				}
			case "timestamp":
				if results[i].Timestamp.Before(results[j].Timestamp) {
					cmp = -1
				} else if results[i].Timestamp.After(results[j].Timestamp) {
					cmp = 1
				}
			case "title":
				cmp = strings.Compare(results[i].Title, results[j].Title)
			}

			if cmp != 0 {
				if field.Order == "desc" {
					return cmp > 0
				}
				return cmp < 0
			}
		}

		return false
	})
}

func calculateFacets(results []Document, facetFields []string) map[string][]FacetItem {
	facets := make(map[string][]FacetItem)

	for _, field := range facetFields {
		counts := make(map[string]int)

		for _, doc := range results {
			switch field {
			case "type":
				if doc.Type != "" {
					counts[doc.Type]++
				}
			case "tags":
				for _, tag := range doc.Tags {
					counts[tag]++
				}
			case "index":
				counts[doc.Index]++
			}
		}

		// Convert to FacetItems
		var items []FacetItem
		for value, count := range counts {
			items = append(items, FacetItem{
				Value: value,
				Count: count,
			})
		}

		// Sort by count
		sort.Slice(items, func(i, j int) bool {
			return items[i].Count > items[j].Count
		})

		facets[field] = items
	}

	return facets
}

// QueryBuilder helps construct search queries
type QueryBuilder struct {
	query Query
}

// NewQueryBuilder creates a new query builder
func NewQueryBuilder(text string) *QueryBuilder {
	return &QueryBuilder{
		query: Query{
			Text: text,
			Size: 10,
		},
	}
}

// WithIndex sets the index to search
func (qb *QueryBuilder) WithIndex(index string) *QueryBuilder {
	qb.query.Index = index
	return qb
}

// WithType sets the document type to filter
func (qb *QueryBuilder) WithType(docType string) *QueryBuilder {
	qb.query.Type = docType
	return qb
}

// WithTags sets tags to filter
func (qb *QueryBuilder) WithTags(tags ...string) *QueryBuilder {
	qb.query.Tags = tags
	return qb
}

// WithPagination sets pagination parameters
func (qb *QueryBuilder) WithPagination(from, size int) *QueryBuilder {
	qb.query.From = from
	qb.query.Size = size
	return qb
}

// WithSort adds sorting criteria
func (qb *QueryBuilder) WithSort(field, order string) *QueryBuilder {
	qb.query.Sort = append(qb.query.Sort, SortField{
		Field: field,
		Order: order,
	})
	return qb
}

// WithHighlight enables highlighting
func (qb *QueryBuilder) WithHighlight() *QueryBuilder {
	qb.query.Highlight = true
	return qb
}

// WithFacets adds facet fields
func (qb *QueryBuilder) WithFacets(fields ...string) *QueryBuilder {
	qb.query.Facets = fields
	return qb
}

// Build returns the constructed query
func (qb *QueryBuilder) Build() Query {
	return qb.query
}
