// Package common contains small conversion helpers used throughout the
// repository. The functions here convert between strings, numbers and other
// primitive types and provide helpers for formatting values.
package common

import (
	"bytes"
	"fmt"
	"math"
	"net/url"
	"strconv"
	"strings"
	"time"
	"unicode"
)

// I2S converts an int64 to its decimal string representation.
// It returns the formatted string for the provided integer.
func I2S(i int64) string {
	return strconv.FormatInt(i, 10)
}

// B2S converts a null terminated byte slice to a string. If no null byte is
// present the entire slice is used.
// It returns the resulting string.
func B2S(b []byte) (s string) {
	n := bytes.Index(b, []byte{0})
	if n > 0 {
		s = string(b[:n])
	} else {
		s = string(b)
	}
	return
}

// F2S converts a float64 to a string using fixed notation with eight digits of
// precision. The returned string is suitable for display or logging purposes.
func F2S(f float64) (s string) {
	return strconv.FormatFloat(f, 'f', 8, 64)
}

// S2F parses a float value from the provided string. It returns the parsed
// number or 0 if the string cannot be interpreted as a number.
func S2F(s string) float64 {
	_, f := ToNumber(s)
	return f
}

// S2I converts a string to an int64. If parsing fails the returned value is 0.
func S2I(s string) int64 {
	i, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0
	} else {
		return i
	}
}

// ToString converts a variety of primitive types to their string
// representation. Supported types are int, int64, float64 and string. For any
// other type fmt.Sprint is used. Nil values return an empty string.
func ToString(x interface{}) string {
	if x == nil {
		return ""
	}
	switch x.(type) {
	case int:
		return strconv.Itoa(x.(int))
	case int64:
		return I2S(x.(int64))
	case float64:
		return F2S(x.(float64))
	case string:
		return x.(string)
	}
	return fmt.Sprint(x)
}

// ToSQLString converts an arbitrary value to a string using ToString and then
// escapes single quotes so the result can be embedded safely in SQL queries.
func ToSQLString(x interface{}) string {
	y := ToString(x)
	return strings.Replace(y, "'", "\\'", -1)
}

// NULLIfEmpty returns the provided string or the literal "NULL" if the value is
// empty or represents a NaN marker.
func NULLIfEmpty(x string) string {
	if x == "" {
		return "NULL"
	}
	if x == "NaN" {
		return "NULL"
	}
	if x == "NANA" {
		return "NULL"
	}
	return x
}

// http://stackoverflow.com/questions/13020308/how-to-fmt-printf-an-integer-with-thousands-comma
// NumberToString formats an integer with a thousands separator.
// The separator rune is inserted every three digits.
func NumberToString(n int, sep rune) string {

	s := strconv.Itoa(n)

	startOffset := 0
	var buff bytes.Buffer

	if n < 0 {
		startOffset = 1
		buff.WriteByte('-')
	}

	l := len(s)

	commaIndex := 3 - ((l - startOffset) % 3)

	if commaIndex == 3 {
		commaIndex = 0
	}

	for i := startOffset; i < l; i++ {

		if commaIndex == 3 {
			buff.WriteRune(sep)
			commaIndex = 0
		}
		commaIndex++

		buff.WriteByte(s[i])
	}

	return buff.String()
}

// ToNumber attempts to parse the provided string as either a float64 or an
// integer. The first return value indicates success and the second contains the
// parsed number. When parsing fails for both float and int, the function returns
// false and 0.
func ToNumber(s string) (bool, float64) {
	f64, err := strconv.ParseFloat(s, 64)
	if err == nil {
		return true, f64
	}
	i64, err := strconv.ParseInt(s, 10, 64)
	if err == nil {
		return true, float64(i64)
	}
	// Non numeric strings result in a failure flag and zero value.
	return false, float64(0)
}

// MonetaryToString formats a float as a currency like value with two decimal
// places.
func MonetaryToString(f float64) string {
	return strings.Trim(fmt.Sprintf("%7.2f", f), " ")
}

// TS converts a Unix timestamp in milliseconds to an ANSI formatted time
// string.
func TS(unixTime int64) (timeFormated string) {
	return time.Unix(int64(unixTime/1000), 0).Format(time.ANSIC)
}

// Reverse returns the string with its characters in reverse order.
func Reverse(s string) string {
	n := len(s)
	runes := make([]rune, n)
	for _, rune := range s {
		n--
		runes[n] = rune
	}
	return string(runes[n:])
}

// Trunc500 truncates a string to a maximum length of 500 characters.
func Trunc500(s string) string {
	if len(s) > 500 {
		return s[:500]
	}
	return s
}

// GetSuffix returns the portion of a string after the final occurrence of the
// supplied split delimiter.
func GetSuffix(s string, split string) string {
	segments := strings.Split(Reverse(s), split)
	n := len(segments)
	if n == 0 {
		return s
	}
	return Reverse(segments[0])
}

// FirstPart returns the first semi-colon separated component of a string.
func FirstPart(s string) string {
	array := strings.Split(s, ";")
	if len(array) == 1 {
		return s
	}
	return array[0]
}

// punctuation lists characters that cause the next letter to be capitalised in
// CamelCase.
var punctuation []string = []string{
	" ",
	"-",
	".",
	":",
	",",
	";",
	"'",
	"`",
	"&",
	"+",
	"=",
	"|",
	"*",
	"/",
	"\\",
	"\"",
	"!",
	"?",
	"(",
	")",
}

// CamelCase converts a string containing separators into camel case. Any
// character found in the punctuation list starts a new word.
func CamelCase(txt string) string {
	out := ""
	isNextUpper := true
	for _, c := range txt {
		if StringInSlice(string(c), punctuation) {
			// Skip punctuation characters and force the next
			// letter to be upper case.
			isNextUpper = true
			continue
		}
		if isNextUpper {
			out += string(unicode.ToUpper(c))
		} else {
			out += string(c)
		}
		isNextUpper = false
	}
	return strings.TrimSpace(out)
}

// Clean lowercases a string, replaces spaces with underscores and URL escapes
// the result.
func Clean(txt string) string {
	return url.QueryEscape(strings.Replace(strings.ToLower(txt), " ", "_", -1))
}

// round rounds a float to the nearest integer using standard mathematical
// rounding. math.Copysign ensures values close to zero round away from zero.
func round(num float64) int {
	return int(num + math.Copysign(0.5, num))
}

// Round rounds a float to the given precision using round.
// Negative numbers are handled correctly thanks to math.Copysign in round.
func Round(num float64, precision int) float64 {
	output := math.Pow(10, float64(precision))
	return float64(round(num*output)) / output
}
