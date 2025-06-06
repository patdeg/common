// Package common contains request and response helpers for JSON and XML.
// It exposes utilities like GetBody, WriteJSON and UnmarshalResponse for
// working with HTTP handlers.
package common

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"

	"golang.org/x/net/context"
)

// GetBody reads the entire body from the provided HTTP request and returns it as
// a byte slice. The caller remains responsible for closing r.Body:
//
//	body := GetBody(r)
//	defer r.Body.Close()
func GetBody(r *http.Request) []byte {
	buffer := new(bytes.Buffer)
	buffer.ReadFrom(r.Body)
	return buffer.Bytes()
}

// GetBodyResponse reads the entire body from the given HTTP response and
// returns it. Always close r.Body after calling this helper:
//
//	body := GetBodyResponse(resp)
//	defer resp.Body.Close()
func GetBodyResponse(r *http.Response) []byte {
	buffer := new(bytes.Buffer)
	buffer.ReadFrom(r.Body)
	return buffer.Bytes()
}

// ReadXML unmarshals the provided XML bytes into the destination structure.
// Example:
//
//	var out MyStruct
//	if err := ReadXML(data, &out); err != nil { ... }
func ReadXML(b []byte, d interface{}) error {
	return xml.Unmarshal(b, d)
}

// WriteXML writes d to the http.ResponseWriter as XML. If withHeader is true
// the standard XML header is included. Example:
//
//	_ = WriteXML(w, data, true)
func WriteXML(w http.ResponseWriter, d interface{}, withHeader bool) error {
	xmlData, err := xml.Marshal(d)
	if err != nil {
		return err
	}
	w.Header().Set("Content-Type", "application/xml")
	if withHeader {
		fmt.Fprintf(w, "%s%s", xml.Header, xmlData)
	} else {
		fmt.Fprintf(w, "%s", xmlData)
	}
	return nil
}

// WriteJSON marshals d to JSON and writes it to the response writer.
// Example:
//
//	_ = WriteJSON(w, data)
func WriteJSON(w http.ResponseWriter, d interface{}) error {
	jsonData, err := json.Marshal(d)
	if err != nil {
		return err
	}
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, "%s", jsonData)
	return nil
}

// ReadJSON unmarshals the JSON byte slice into the destination value.
// Example:
//
//	var out MyStruct
//	if err := ReadJSON(data, &out); err != nil { ... }
func ReadJSON(b []byte, d interface{}) error {
	return json.Unmarshal(b, d)
}

// UnmarshalResponse dumps the response to the debug log, reads the body and
// unmarshals JSON into value. Example:
//
//	var out MyStruct
//	if err := UnmarshalResponse(ctx, resp, &out); err != nil { ... }
func UnmarshalResponse(c context.Context, resp *http.Response, value interface{}) error {

	DumpResponse(c, resp)

	body, err := ioutil.ReadAll(resp.Body) // load entire body for decoding
	if err != nil {                        // I/O error while reading body
		Error("Error while reading body: %v", err)
		return err
	}

	// Attempt JSON decoding into the caller provided structure.
	err = json.Unmarshal(body, value)
	if err != nil { // decoding failed, log the raw body for troubleshooting
		Error("Error while decoding JSON: %v", err)
		Info("JSON: %v", B2S(body))
		return err
	}

	return nil
}

// UnmarshalRequest reads the HTTP request body and decodes the JSON payload
// into value. Example:
//
//	var in MyStruct
//	if err := UnmarshalRequest(ctx, r, &in); err != nil { ... }
func UnmarshalRequest(c context.Context, r *http.Request, value interface{}) error {

	body := GetBody(r)

	err := json.Unmarshal(body, value)
	if err != nil {
		Error("Error while decoding JSON: %v", err)
		Info("JSON: %v", B2S(body))
		return err
	}

	return nil
}

// Marshal returns the JSON encoding of value as a string, logging any error.
// Example:
//
//	jsonStr := Marshal(ctx, data)
func Marshal(c context.Context, value interface{}) string {
	data, err := json.Marshal(value)
	if err != nil {
		Error("[common.Marshal] Error converting json: %v", err)
		return ""
	}
	return B2S(data)
}

// Left returns the first n characters of s. If s is shorter than n it returns
// s unchanged. Example:
//
//	prefix := Left("abcdef", 3) // "abc"
func Left(s string, n int) string {
	if len(s) < n {
		return s
	}
	return s[:n]
}
