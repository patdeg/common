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

package common

import (
	"bytes"
	"sync"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
)

var (
	markdownConverter     goldmark.Markdown
	markdownConverterOnce sync.Once
)

// getMarkdownConverter returns a singleton markdown converter configured with
// GitHub Flavored Markdown extensions.
func getMarkdownConverter() goldmark.Markdown {
	markdownConverterOnce.Do(func() {
		markdownConverter = goldmark.New(
			goldmark.WithExtensions(
				extension.GFM,           // GitHub Flavored Markdown
				extension.Table,         // Tables
				extension.Strikethrough, // Strikethrough text
				extension.Linkify,       // Auto-link URLs
				extension.TaskList,      // Task list items
			),
			goldmark.WithParserOptions(
				parser.WithAutoHeadingID(), // Auto-generate heading IDs
			),
			goldmark.WithRendererOptions(
				html.WithHardWraps(), // Preserve line breaks
				html.WithXHTML(),     // XHTML-compatible output
			),
		)
	})
	return markdownConverter
}

// MarkdownToHTML converts markdown text to HTML.
// It uses GitHub Flavored Markdown extensions including tables, strikethrough,
// auto-linking, and task lists.
func MarkdownToHTML(markdown string) (string, error) {
	var buf bytes.Buffer
	if err := getMarkdownConverter().Convert([]byte(markdown), &buf); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// MarkdownToHTMLWithWrapper converts markdown text to HTML and wraps it in a
// basic HTML document structure suitable for email. The wrapper includes
// minimal styling for readability.
func MarkdownToHTMLWithWrapper(markdown string) (string, error) {
	content, err := MarkdownToHTML(markdown)
	if err != nil {
		return "", err
	}

	// Wrap in a basic HTML document with minimal inline styling for email clients
	html := `<!DOCTYPE html>
<html>
<head>
<meta charset="utf-8">
<style>
body {
    font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif;
    font-size: 14px;
    line-height: 1.6;
    color: #333;
    max-width: 800px;
    margin: 0 auto;
    padding: 20px;
}
h1, h2, h3, h4 { color: #222; margin-top: 24px; margin-bottom: 16px; }
h1 { font-size: 2em; border-bottom: 1px solid #eee; padding-bottom: 0.3em; }
h2 { font-size: 1.5em; border-bottom: 1px solid #eee; padding-bottom: 0.3em; }
h3 { font-size: 1.25em; }
code {
    background-color: #f6f8fa;
    border-radius: 3px;
    font-family: 'SFMono-Regular', Consolas, 'Liberation Mono', Menlo, monospace;
    font-size: 85%;
    padding: 0.2em 0.4em;
}
pre {
    background-color: #f6f8fa;
    border-radius: 6px;
    font-size: 85%;
    line-height: 1.45;
    overflow: auto;
    padding: 16px;
}
pre code {
    background-color: transparent;
    padding: 0;
}
blockquote {
    border-left: 4px solid #dfe2e5;
    color: #6a737d;
    margin: 0;
    padding: 0 1em;
}
table {
    border-collapse: collapse;
    margin-bottom: 16px;
}
table th, table td {
    border: 1px solid #dfe2e5;
    padding: 6px 13px;
}
table tr:nth-child(2n) {
    background-color: #f6f8fa;
}
ul, ol {
    padding-left: 2em;
}
li + li {
    margin-top: 0.25em;
}
a {
    color: #0366d6;
    text-decoration: none;
}
a:hover {
    text-decoration: underline;
}
strong { font-weight: 600; }
em { font-style: italic; }
hr {
    background-color: #e1e4e8;
    border: 0;
    height: 0.25em;
    margin: 24px 0;
    padding: 0;
}
</style>
</head>
<body>
` + content + `
</body>
</html>`

	return html, nil
}
