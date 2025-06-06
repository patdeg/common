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

func I2S(i int64) string {
	return strconv.FormatInt(i, 10)
}

// Small utility function to convert a byte to a string
func B2S(b []byte) (s string) {
	n := bytes.Index(b, []byte{0})
	if n > 0 {
		s = string(b[:n])
	} else {
		s = string(b)
	}
	return
}

// Small utility function to convert a float to a string
func F2S(f float64) (s string) {
	return strconv.FormatFloat(f, 'f', 8, 64)
}

// Small utility function to convert a string to a float
func S2F(s string) float64 {
	_, f := ToNumber(s)
	return f
}

// Small utility function to convert a string to an integer
func S2I(s string) int64 {
	i, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0
	} else {
		return i
	}
}

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
	return x.(string)
}

func ToSQLString(x interface{}) string {
	y := ToString(x)
	return strings.Replace(y, "'", "\\'", -1)
}

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

func ToNumber(s string) (bool, float64) {
	f64, err := strconv.ParseFloat(s, 64)
	if err == nil {
		return true, f64
	}
	i64, err := strconv.ParseInt(s, 10, 64)
	if err == nil {
		return true, float64(i64)
	}
	return false, float64(0)
}

func MonetaryToString(f float64) string {
	return strings.Trim(fmt.Sprintf("%7.2f", f), " ")
}

func TS(unixTime int64) (timeFormated string) {
	return time.Unix(int64(unixTime/1000), 0).Format(time.ANSIC)
}

func Reverse(s string) string {
	n := len(s)
	runes := make([]rune, n)
	for _, rune := range s {
		n--
		runes[n] = rune
	}
	return string(runes[n:])
}

func Trunc500(s string) string {
	if len(s) > 500 {
		return s[:500]
	}
	return s
}

func GetSuffix(s string, split string) string {
	segments := strings.Split(Reverse(s), split)
	n := len(segments)
	if n == 0 {
		return s
	}
	return Reverse(segments[0])
}

func FirstPart(s string) string {
	array := strings.Split(s, ";")
	if len(array) == 1 {
		return s
	}
	return array[0]
}

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

func CamelCase(txt string) string {
	out := ""
	isNextUpper := true
	for _, c := range txt {
		if StringInSlice(string(c), punctuation) {
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

func Clean(txt string) string {
	return url.QueryEscape(strings.Replace(strings.ToLower(txt), " ", "_", -1))
}

func round(num float64) int {
	return int(num + math.Copysign(0.5, num))
}

func Round(num float64, precision int) float64 {
	output := math.Pow(10, float64(precision))
	return float64(round(num*output)) / output
}
