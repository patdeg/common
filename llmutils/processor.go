// Package llmutils provides utilities for processing LLM prompt templates
// with support for comment stripping and metadata extraction.
//
// This package combines comment stripping from the API repository with
// metadata extraction from both the web and API repositories to provide
// a unified interface for prompt processing.
//
// Comment Format:
// - Lines starting with /// are treated as comments
// - Inline comments after /// are supported
// - Comments are stripped before sending to LLM
//
// Special Directives:
// - /// param: value - Extracts metadata for BigQuery tracking
// - /// key: value - Generic metadata extraction
//
// Example:
//
//	/// This is a comment that will be removed
//	/// param: temperature=0.7
//	/// flow: checkout-process
//	You are a helpful assistant /// inline comment removed
package llmutils

import (
	"regexp"
	"strings"
)

// ProcessedPrompt contains the cleaned prompt and extracted metadata.
type ProcessedPrompt struct {
	// CleanedPrompt is the prompt with all /// comments removed
	CleanedPrompt string

	// Params contains extracted metadata from /// param: directives
	// Multiple params can be combined: temperature=0.7, max_tokens=1000
	Params map[string]string

	// Metadata contains all extracted key-value pairs from /// directives
	Metadata map[string]string

	// Flow is the application flow name (from /// flow: directive)
	Flow string

	// Node is the node/step name within the flow (from /// node: directive)
	Node string

	// Tags contains additional tags extracted from directives
	Tags []string
}

// Process removes /// comments from a prompt template and extracts metadata.
//
// Comment Processing:
//   - Full-line comments: Lines containing only /// and comment text are removed
//   - Inline comments: Content after /// on a line is removed
//   - Whitespace handling: Trailing whitespace is trimmed from lines with inline comments
//   - URL protection: URLs like http://example.com are NOT treated as comments
//
// Metadata Extraction:
//   - /// param: key=value - Extracts key-value pairs for BigQuery metadata
//   - /// param: temperature=0.7, max_tokens=1000 - Supports comma-separated values
//   - /// flow: name - Extracts application flow name
//   - /// node: name - Extracts node/step name
//   - /// tag: value - Extracts tags (comma-separated)
//   - /// key: value - Generic metadata extraction
//
// Example:
//
//	input := `
//	/// This is a full-line comment - will be removed
//	/// param: model=gpt-4, temperature=0.7
//	/// flow: checkout-process
//	/// node: payment-validation
//	You are a helpful assistant /// inline comment
//	Visit http://example.com for more /// URL preserved
//	`
//
//	result := Process(input)
//	// result.CleanedPrompt = "You are a helpful assistant\nVisit http://example.com for more"
//	// result.Params = {"model": "gpt-4", "temperature": "0.7"}
//	// result.Flow = "checkout-process"
//	// result.Node = "payment-validation"
func Process(content string) ProcessedPrompt {
	lines := strings.Split(content, "\n")
	var cleanedLines []string
	params := make(map[string]string)
	metadata := make(map[string]string)
	tags := make([]string, 0)
	var flow, node string

	// Regex pattern for metadata extraction (key: value)
	metadataPattern := regexp.MustCompile(`^\s*(\w+):\s*(.+)$`)

	for _, line := range lines {
		// Find the position of "///"
		commentPos := findCommentPosition(line)
		if commentPos == -1 {
			// No comment found, keep the line as is
			cleanedLines = append(cleanedLines, line)
			continue
		}

		// Extract the part before and after ///
		beforeComment := line[:commentPos]
		afterComment := line[commentPos+3:] // Everything after ///

		// Extract metadata from comment
		trimmedComment := strings.TrimSpace(afterComment)

		// Check for param directive
		if strings.HasPrefix(trimmedComment, "param:") {
			paramValue := strings.TrimSpace(strings.TrimPrefix(trimmedComment, "param:"))
			extractParams(paramValue, params)
		} else if match := metadataPattern.FindStringSubmatch(trimmedComment); match != nil {
			// Extract generic metadata (key: value)
			key := strings.ToLower(strings.TrimSpace(match[1]))
			value := strings.TrimSpace(match[2])

			// Store in metadata map
			metadata[key] = value

			// Handle special keys
			switch key {
			case "flow":
				flow = value
			case "node":
				node = value
			case "tag", "tags":
				// Split tags by comma if multiple
				tagList := strings.Split(value, ",")
				for _, tag := range tagList {
					tag = strings.TrimSpace(tag)
					if tag != "" {
						tags = append(tags, tag)
					}
				}
			}
		}

		// Handle line cleaning
		if strings.TrimSpace(beforeComment) == "" {
			// Full-line comment - skip the entire line
			continue
		}

		// Inline comment - keep the part before the comment (removing trailing whitespace)
		cleanedLines = append(cleanedLines, strings.TrimRight(beforeComment, " \t"))
	}

	// Auto-generate tags from flow and node if present
	if flow != "" {
		tags = append(tags, "flow:"+flow)
	}
	if node != "" {
		tags = append(tags, "node:"+node)
	}

	return ProcessedPrompt{
		CleanedPrompt: strings.Join(cleanedLines, "\n"),
		Params:        params,
		Metadata:      metadata,
		Flow:          flow,
		Node:          node,
		Tags:          tags,
	}
}

// extractParams parses comma-separated key=value pairs and adds them to the params map.
//
// Supported formats:
//   - key=value
//   - key=value, key2=value2
//   - key = value (whitespace allowed around =)
//
// Example:
//
//	extractParams("temperature=0.7, max_tokens=1000", params)
//	// params["temperature"] = "0.7"
//	// params["max_tokens"] = "1000"
func extractParams(paramValue string, params map[string]string) {
	// Split by comma to handle multiple params
	pairs := strings.Split(paramValue, ",")

	for _, pair := range pairs {
		pair = strings.TrimSpace(pair)
		if pair == "" {
			continue
		}

		// Split by = to get key and value
		parts := strings.SplitN(pair, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		if key != "" {
			params[key] = value
		}
	}
}

// findCommentPosition finds the position of /// comment marker in a line.
//
// Returns:
//   - Position of /// if found and not part of a URL
//   - -1 if no comment found or all /// are part of URLs
//
// Algorithm:
//  1. Search for all occurrences of ///
//  2. For each occurrence, check if it's part of http:// or https://
//  3. Return position of first non-URL ///
//  4. Return -1 if all /// are part of URLs
//
// URL Detection Logic:
//   - Check 5 characters before /// position
//   - If we find "http:" or "https:" (case-insensitive), it's a URL
//   - The :// sequence is unique to URLs and won't appear in regular text
func findCommentPosition(line string) int {
	// Simple case: no /// at all
	if !strings.Contains(line, "///") {
		return -1
	}

	// Find all occurrences of ///
	pos := 0
	for {
		idx := strings.Index(line[pos:], "///")
		if idx == -1 {
			// No more /// found
			return -1
		}

		// Adjust to absolute position
		absPos := pos + idx

		// Check if this /// is part of a URL (http:/// or https:///)
		if !isPartOfURL(line, absPos) {
			// This is a comment marker, not a URL
			return absPos
		}

		// This /// is part of a URL, keep searching
		pos = absPos + 3 // Move past this /// to continue searching
	}
}

// isPartOfURL checks if /// at the given position is part of a URL.
//
// Logic:
// - Look backwards from position to find protocol markers
// - Check for http:// or https:// patterns
// - Must be immediately before the ///
//
// Parameters:
//   - line: The full line of text
//   - slashPos: Position where /// starts
//
// Returns:
//   - true if /// is part of http:// or https://
//   - false otherwise
func isPartOfURL(line string, slashPos int) bool {
	// Need at least "http:" before /// (5 chars for http:, 6 for https:)
	if slashPos < 5 {
		return false // Not enough room for a protocol
	}

	// Check for http:// (5 chars before + ///)
	if slashPos >= 5 {
		prefix := strings.ToLower(line[slashPos-5 : slashPos])
		if prefix == "http:" {
			return true
		}
	}

	// Check for https:// (6 chars before + ///)
	if slashPos >= 6 {
		prefix := strings.ToLower(line[slashPos-6 : slashPos])
		if prefix == "https:" {
			return true
		}
	}

	return false
}

// StripComments is a convenience function that only removes comments without extracting params.
// Use Process() instead if you need param extraction.
//
// This function removes blank lines after comment removal and handles URLs correctly.
func StripComments(content string) string {
	if content == "" {
		return content
	}

	lines := strings.Split(content, "\n")
	result := make([]string, 0, len(lines))

	for _, line := range lines {
		cleaned := stripCommentFromLine(line)

		// Only keep non-empty lines (or lines with just whitespace that aren't all whitespace)
		if strings.TrimSpace(cleaned) != "" {
			result = append(result, cleaned)
		}
	}

	return strings.Join(result, "\n")
}

// stripCommentFromLine removes /// comments from a single line.
//
// Logic:
//  1. Check if line starts with /// (after whitespace) → return empty string
//  2. Find first occurrence of /// that's not part of a URL
//  3. Return text before the comment marker
//
// URL Detection:
// - If /// is preceded by "http:" or "https:", it's part of a URL and NOT a comment
// - We scan backwards from the /// position to check for protocol markers
func stripCommentFromLine(line string) string {
	// Check if line is a full-line comment (starts with /// after optional whitespace)
	trimmed := strings.TrimLeft(line, " \t")
	if strings.HasPrefix(trimmed, "///") {
		return ""
	}

	// Look for inline comment marker ///
	commentPos := findCommentPosition(line)
	if commentPos == -1 {
		// No comment found
		return line
	}

	// Return text before comment, trimming trailing whitespace
	return strings.TrimRight(line[:commentPos], " \t")
}

// ExtractParams extracts only the params from a prompt template without cleaning.
// This is useful if you want to inspect metadata before processing.
func ExtractParams(content string) map[string]string {
	return Process(content).Params
}

// ExtractMetadata extracts all metadata (params, flow, node, tags) from a prompt template.
// Returns a ProcessedPrompt with only metadata fields populated.
func ExtractMetadata(content string) ProcessedPrompt {
	result := Process(content)
	// Clear the cleaned prompt to indicate this is metadata-only
	result.CleanedPrompt = ""
	return result
}

// StripCommentsFromMessages strips /// comments from all messages in a chat completion request.
//
// This function is designed to integrate with LLM APIs (OpenAI, Groq, etc.).
// It processes both system and user messages, removing comments while preserving
// the message structure.
//
// Parameters:
//   - messages: Array of message maps with "role" and "content" fields
//
// Returns:
//   - Modified messages array with comments stripped from all content fields
//
// Example:
//
//	messages := []map[string]interface{}{
//	    {"role": "system", "content": "You are helpful /// be nice"},
//	    {"role": "user", "content": "/// Debug note\nTell me a joke"},
//	}
//	cleaned := StripCommentsFromMessages(messages)
//	// Result:
//	// [
//	//   {"role": "system", "content": "You are helpful"},
//	//   {"role": "user", "content": "Tell me a joke"},
//	// ]
func StripCommentsFromMessages(messages []interface{}) []interface{} {
	result := make([]interface{}, len(messages))

	for i, msg := range messages {
		// Create a copy of the message map
		if msgMap, ok := msg.(map[string]interface{}); ok {
			// Create new map with same keys
			cleanedMsg := make(map[string]interface{}, len(msgMap))
			for k, v := range msgMap {
				if k == "content" {
					// Strip comments from content field
					if content, ok := v.(string); ok {
						cleanedMsg[k] = StripComments(content)
					} else {
						cleanedMsg[k] = v // Non-string content, keep as-is
					}
				} else {
					cleanedMsg[k] = v // Other fields unchanged
				}
			}
			result[i] = cleanedMsg
		} else {
			// Not a map, keep as-is
			result[i] = msg
		}
	}

	return result
}

// IMPLEMENTATION NOTES
//
// Design Decisions:
// 1. Three-slash delimiter (///) chosen to differentiate from standard Go comments (//)
// 2. Support both full-line and inline comments for flexibility
// 3. Params use key=value format, compatible with URL query string syntax
// 4. Multiple params on same line supported via comma separation
// 5. Case-sensitive param keys to match BigQuery column names
// 6. Generic metadata extraction with key: value syntax (case-insensitive keys)
// 7. URL protection: http:// and https:// are NOT treated as comments
//
// Usage Patterns:
// - Template authoring: Use /// for documentation and guidance
// - LLM invocation: Strip comments to reduce token usage
// - Analytics: Extract params for request categorization
// - A/B testing: Use params to track prompt variations
// - Flow tracking: Use flow/node directives for application state tracking
//
// Security Considerations:
// - Comments are removed client-side, never sent to LLM
// - Param values are NOT sanitized - validate before storing
// - Do not put secrets in /// comments
// - Params are stored in BigQuery metadata field (JSON)
//
// Performance:
// - O(n) where n is number of lines
// - Single pass through content for comment stripping
// - Minimal allocations (pre-sized maps where possible)
// - String concatenation uses strings.Join for efficiency
// - Regex compilation is done once per Process call
//
// Future Enhancements:
// - Support for multi-line param values (e.g., JSON in param)
// - Param validation against schema
// - Template inheritance/composition
// - Conditional comment blocks (e.g., /// if: env=dev)
