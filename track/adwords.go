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

// Package track contains analytics helpers.
//
// This file implements AdWords click tracking. The handler stores the
// query parameters provided by Google Ads in BigQuery before redirecting
// the visitor to the landing page.
package track

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/patdeg/common"
	"github.com/patdeg/common/gcp"

	"github.com/mssola/user_agent"
	"golang.org/x/net/context"
	bigquery "google.golang.org/api/bigquery/v2"
	appengine "google.golang.org/appengine/v2"
	"google.golang.org/appengine/v2/user"
)

type Click struct {
	Time            time.Time `json:"time,omitempty"`
	RedirectUrl     string    `json:"redirectUrl,omitempty"`
	Query           string    `json:"query,omitempty"`
	Campaignid      string    `json:"campaignid,omitempty"`
	Adgroupid       string    `json:"adgroupid,omitempty"`
	Feeditemid      string    `json:"feeditemid,omitempty"`
	Targetid        string    `json:"targetid,omitempty"`
	Loc_Physical_Ms string    `json:"loc_physical_ms,omitempty"`
	Loc_Interest_Ms string    `json:"loc_interest_ms,omitempty"`
	Matchtype       string    `json:"matchtype,omitempty"`
	Network         string    `json:"network,omitempty"`
	Device          string    `json:"device,omitempty"`
	Devicemodel     string    `json:"devicemodel,omitempty"`
	Creative        string    `json:"creative,omitempty"`
	Keyword         string    `json:"keyword,omitempty"`
	Placement       string    `json:"placement,omitempty"`
	Target          string    `json:"target,omitempty"`
	Param1          string    `json:"param1,omitempty"`
	Param2          string    `json:"param2,omitempty"`
	Random          string    `json:"random,omitempty"`
	Aceid           string    `json:"aceid,omitempty"`
	Adposition      string    `json:"adposition,omitempty"`
	Ignore          string    `json:"ignore,omitempty"`
	Lpurl           string    `json:"lpurl,omitempty"`
	Cookie          string    `json:"cookie,omitempty"`
	Referer         string    `json:"referer,omitempty"`
	Host            string    `json:"host,omitempty"`
	RemoteAddr      string    `json:"remoteAddr,omitempty"`
	InstanceId      string    `json:"instanceId,omitempty"`
	VersionId       string    `json:"versionId,omitempty"`
	Scheme          string    `json:"scheme,omitempty"`
	Country         string    `json:"country,omitempty"`
	Region          string    `json:"region,omitempty"`
	City            string    `json:"city,omitempty"`
	Lat             float64   `json:"lat,omitempty"`
	Lon             float64   `json:"lon,omitempty"`
	AcceptLanguage  string    `json:"acceptLanguage,omitempty"`
	UserAgent       string    `json:"userAgent,omitempty"`
	IsMobile        bool      `json:"isMobile,omitempty"`
	IsBot           bool      `json:"isBot,omitempty"`
	MozillaVersion  string    `json:"mozillaVersion,omitempty"`
	Platform        string    `json:"platform,omitempty"`
	OS              string    `json:"os,omitempty"`
	EngineName      string    `json:"engineName,omitempty"`
	EngineVersion   string    `json:"engineVersion,omitempty"`
	BrowserName     string    `json:"browserName,omitempty"`
	BrowserVersion  string    `json:"browserVersion,omitempty"`
}

// createClicksTableInBigQuery creates the daily AdWords clicks table named
// by the YYYYMMDD string d. The function ensures the dataset exists and
// then attempts to create the table. It returns any error encountered
// while creating the dataset or table, or when d does not match the
// expected date format.
func createClicksTableInBigQuery(c context.Context, d string) error {

	common.Info("Create a new daily clicks table in BigQuery")

	if err := gcp.CreateDatasetIfNotExists(c, adwordsProjectID, adwordsDataset); err != nil {
		common.Error("Error ensuring dataset %s: %v", adwordsDataset, err)
		return err
	}

	if len(d) != 8 {
		return errors.New("table name is badly formatted - expected 8 characters")
	}
	newTable := &bigquery.Table{
		TableReference: &bigquery.TableReference{
			ProjectId: adwordsProjectID,
			DatasetId: adwordsDataset,
			TableId:   d,
		},
		FriendlyName: "Daily Clicks table",
		Description:  "This table is created automatically to store daily AdWords clicks to Deglon Consulting properties ",
		//ExpirationTime: expirationTime.Unix() * 1000,
		Schema: &bigquery.TableSchema{
			Fields: []*bigquery.TableFieldSchema{
				{Name: "Time", Type: "TIMESTAMP", Description: "Time"},
				{Name: "RedirectUrl", Type: "STRING", Description: "RedirectUrl"},
				{Name: "Query", Type: "STRING", Description: "Query"},
				{Name: "Campaignid", Type: "STRING", Description: "Campaignid"},
				{Name: "Adgroupid", Type: "STRING", Description: "Adgroupid"},
				{Name: "Feeditemid", Type: "STRING", Description: "Feeditemid"},
				{Name: "Targetid", Type: "STRING", Description: "Targetid"},
				{Name: "Loc_Physical_Ms", Type: "STRING", Description: "Loc_Physical_Ms"},
				{Name: "Loc_Interest_Ms", Type: "STRING", Description: "Loc_Interest_Ms"},
				{Name: "Matchtype", Type: "STRING", Description: "Matchtype"},
				{Name: "Network", Type: "STRING", Description: "Network"},
				{Name: "Device", Type: "STRING", Description: "Device"},
				{Name: "Devicemodel", Type: "STRING", Description: "Devicemodel"},
				{Name: "Creative", Type: "STRING", Description: "Creative"},
				{Name: "Keyword", Type: "STRING", Description: "Keyword"},
				{Name: "Placement", Type: "STRING", Description: "Placement"},
				{Name: "Target", Type: "STRING", Description: "Target"},
				{Name: "Param1", Type: "STRING", Description: "Param1"},
				{Name: "Param2", Type: "STRING", Description: "Param2"},
				{Name: "Random", Type: "STRING", Description: "Random"},
				{Name: "Aceid", Type: "STRING", Description: "Aceid"},
				{Name: "Adposition", Type: "STRING", Description: "Adposition"},
				{Name: "Ignore", Type: "STRING", Description: "Ignore"},
				{Name: "Lpurl", Type: "STRING", Description: "Lpurl"},
				{Name: "Cookie", Type: "STRING", Description: "Cookie"},
				{Name: "Referer", Type: "STRING", Description: "Referer"},
				{Name: "Host", Type: "STRING", Description: "Host"},
				{Name: "RemoteAddr", Type: "STRING", Description: "RemoteAddr"},
				{Name: "InstanceId", Type: "STRING", Description: "InstanceId"},
				{Name: "VersionId", Type: "STRING", Description: "VersionId"},
				{Name: "Scheme", Type: "STRING", Description: "Scheme"},
				{Name: "Country", Type: "STRING", Description: "Country"},
				{Name: "Region", Type: "STRING", Description: "Region"},
				{Name: "City", Type: "STRING", Description: "City"},
				{Name: "Lat", Type: "FLOAT", Description: "City latitude"},
				{Name: "Lon", Type: "FLOAT", Description: "City longitude"},
				{Name: "AcceptLanguage", Type: "STRING", Description: "AcceptLanguage"},
				{Name: "UserAgent", Type: "STRING", Description: "UserAgent"},
				{Name: "IsMobile", Type: "BOOLEAN", Description: "IsMobile"},
				{Name: "IsBot", Type: "BOOLEAN", Description: "IsBot"},
				{Name: "MozillaVersion", Type: "STRING", Description: "MozillaVersion"},
				{Name: "Platform", Type: "STRING", Description: "Platform"},
				{Name: "OS", Type: "STRING", Description: "OS"},
				{Name: "EngineName", Type: "STRING", Description: "EngineName"},
				{Name: "EngineVersion", Type: "STRING", Description: "EngineVersion"},
				{Name: "BrowserName", Type: "STRING", Description: "BrowserName"},
				{Name: "BrowserVersion", Type: "STRING", Description: "BrowserVersion"},
			},
		},
	}

	return gcp.CreateTableInBigQuery(c, newTable)
}

// CreateTodayClicksTableInBigQueryHandler sets up today's clicks table in
// BigQuery. It is typically called by a cron job or an administrator.
// Errors while creating the table are sent as HTTP 500 responses.
func CreateTodayClicksTableInBigQueryHandler(w http.ResponseWriter, r *http.Request) {
	c := r.Context()
	common.Info(">>> CreateTodayClicksTableInBigQueryHandler")

	isAdmin := user.IsAdmin(c)

	if (r.Header.Get("X-AppEngine-Cron") != "true") && (isAdmin == false) {
		common.Error("Handler called without admin/cron privilege")
		http.Error(w, "Handler called without admin/cron privilege", http.StatusBadRequest)
		return
	}

	today := time.Now().Format("20060102")
	err := createClicksTableInBigQuery(c, today)
	if err != nil {
		common.Error("Error while creating table %v: %v", today, err)
		http.Error(w, "Error while creating today table: "+err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "Table %v created", today)

}

// CreateTomorrowClicksTableInBigQueryHandler creates the BigQuery table for
// tomorrow's clicks data. It behaves like CreateTodayClicksTableInBigQueryHandler
// but uses tomorrow's date when calling createClicksTableInBigQuery.
func CreateTomorrowClicksTableInBigQueryHandler(w http.ResponseWriter, r *http.Request) {
	c := r.Context()
	common.Info(">>> CreateTomorrowClicksTableInBigQueryHandler")

	isAdmin := user.IsAdmin(c)

	if (r.Header.Get("X-AppEngine-Cron") != "true") && (isAdmin == false) {
		common.Error("Handler called without admin/cron privilege")
		http.Error(w, "Handler called without admin/cron privilege", http.StatusBadRequest)
		return
	}

	tomorrow := time.Now().Add(time.Hour*23 + time.Minute*59).Format("20060102")
	err := createClicksTableInBigQuery(c, tomorrow)
	if err != nil {
		common.Error("Error while creating table %v: %v", tomorrow, err)
		http.Error(w, "Error while creating tomorrow table: "+err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "Table %v created", tomorrow)

}

// StoreClickInBigQuery streams an AdWords click record to BigQuery. If the
// daily table for today does not exist, insertWithTableCreation will create it
// before retrying the insert. Any error from BigQuery or table creation is
// returned to the caller.
func StoreClickInBigQuery(c context.Context, click *Click) error {

	req := &bigquery.TableDataInsertAllRequest{
		Kind: "bigquery#tableDataInsertAllRequest",
		Rows: []*bigquery.TableDataInsertAllRequestRows{
			{
				InsertId: click.RemoteAddr + common.I2S(click.Time.UnixNano()),
				Json: map[string]bigquery.JsonValue{
					"Time":            click.Time,
					"RedirectUrl":     click.RedirectUrl,
					"Query":           click.Query,
					"Campaignid":      click.Campaignid,
					"Adgroupid":       click.Adgroupid,
					"Feeditemid":      click.Feeditemid,
					"Targetid":        click.Targetid,
					"Loc_Physical_Ms": click.Loc_Physical_Ms,
					"Loc_Interest_Ms": click.Loc_Interest_Ms,
					"Matchtype":       click.Matchtype,
					"Network":         click.Network,
					"Device":          click.Device,
					"Devicemodel":     click.Devicemodel,
					"Creative":        click.Creative,
					"Keyword":         click.Keyword,
					"Placement":       click.Placement,
					"Target":          click.Target,
					"Param1":          click.Param1,
					"Param2":          click.Param2,
					"Random":          click.Random,
					"Aceid":           click.Aceid,
					"Adposition":      click.Adposition,
					"Ignore":          click.Ignore,
					"Lpurl":           click.Lpurl,
					"Cookie":          click.Cookie,
					"Referer":         click.Referer,
					"Host":            click.Host,
					"RemoteAddr":      click.RemoteAddr,
					"InstanceId":      click.InstanceId,
					"VersionId":       click.VersionId,
					"Scheme":          click.Scheme,
					"Country":         click.Country,
					"Region":          click.Region,
					"City":            click.City,
					"Lat":             click.Lat,
					"Lon":             click.Lon,
					"AcceptLanguage":  click.AcceptLanguage,
					"UserAgent":       click.UserAgent,
					"IsMobile":        click.IsMobile,
					"IsBot":           click.IsBot,
					"MozillaVersion":  click.MozillaVersion,
					"Platform":        click.Platform,
					"OS":              click.OS,
					"EngineName":      click.EngineName,
					"EngineVersion":   click.EngineVersion,
					"BrowserName":     click.BrowserName,
					"BrowserVersion":  click.BrowserVersion,
				},
			},
		},
	}

	tableName := time.Now().Format("20060102")

	return insertWithTableCreation(c, adwordsProjectID, adwordsDataset, tableName, req, createClicksTableInBigQuery)
}

// AdWordsTrackingHandler collects detailed AdWords click information from the
// request's query parameters and stores it in BigQuery. Expected parameters
// include `url` for the landing page as well as `k` (keyword), `cm` (campaign
// ID) and other standard Google Ads values. After recording the click, the
// handler redirects the user to the validated `url` parameter.
func AdWordsTrackingHandler(w http.ResponseWriter, r *http.Request) {
	c := r.Context()

	common.Debug(">>>> AdWordsTrackingHandler")
	common.Debug("Referer: %v", r.Host)
	common.Debug("Referer: %v", r.Referer())
	common.Debug("RequestURI: %v", r.RequestURI)

	// Track Visitor cookie ID
	cookie := common.GetCookieID(w, r) // visitor cookie
	common.Debug("Cookie ID: %v", cookie)

	// landing page specified by the `url` query parameter
	redirectUrl := r.FormValue("url")
	common.Debug("Redirect URL: %v", redirectUrl)
	// Validate redirectUrl before using it
	if !common.IsValidHTTPURL(redirectUrl) {
		common.Error("invalid redirect URL %q", redirectUrl)
		http.Error(w, "invalid redirect URL", http.StatusBadRequest)
		return
	}

	ua := user_agent.New(r.Header.Get("User-Agent"))
	engineName, engineversion := ua.Engine()
	browserName, browserVersion := ua.Browser()

	lat := float64(0)
	lon := float64(0)
	latlon := strings.Split(r.Header.Get("X-AppEngine-CityLatLong"), ",")
	if len(latlon) == 2 {
		lat = common.S2F(latlon[0])
		lon = common.S2F(latlon[1])
	}

	query := ""
	if r.Header.Get("Referer") != "" {
		if refererUrl, err := url.Parse(r.Header.Get("Referer")); err == nil {
			query = refererUrl.Query().Get("q")
			common.Debug("Search query: %v", query)
		} else {
			common.Error("Error, can't parse referer %v", r.Header.Get("Referer"))
		}
	}

	click := Click{
		Time:        time.Now(),
		RedirectUrl: redirectUrl,
		Query:       common.Trunc500(query),
		// AdWords specific query parameters
		Campaignid:      r.FormValue("cm"),  // campaign id
		Adgroupid:       r.FormValue("ag"),  // ad group id
		Feeditemid:      r.FormValue("f"),   // feed item id
		Targetid:        r.FormValue("tid"), // target id
		Loc_Physical_Ms: r.FormValue("lp"),
		Loc_Interest_Ms: r.FormValue("li"),
		Matchtype:       r.FormValue("m"),
		Network:         r.FormValue("n"),
		Device:          r.FormValue("d"),                  // device type
		Devicemodel:     r.FormValue("dm"),                 // device model
		Creative:        r.FormValue("cr"),                 // creative id
		Keyword:         common.Trunc500(r.FormValue("k")), // matched keyword
		Placement:       r.FormValue("p"),                  // placement
		Target:          r.FormValue("t"),                  // target
		Param1:          r.FormValue("p1"),                 // custom parameter 1
		Param2:          r.FormValue("p2"),                 // custom parameter 2
		Random:          r.FormValue("r"),                  // random number
		Aceid:           r.FormValue("a"),                  // aceid
		Adposition:      r.FormValue("ap"),                 // ad position
		Ignore:          r.FormValue("i"),                  // ignore flag
		Lpurl:           r.FormValue("url"),
		Cookie:          cookie, // visitor cookie id
		Referer:         common.Trunc500(r.Header.Get("Referer")),
		Host:            r.Host,
		RemoteAddr:      r.RemoteAddr,
		InstanceId:      appengine.InstanceID(),
		VersionId:       appengine.VersionID(c),
		Scheme:          r.URL.Scheme,
		Country:         r.Header.Get("X-AppEngine-Country"),
		Region:          r.Header.Get("X-AppEngine-Region"),
		City:            r.Header.Get("X-AppEngine-City"),
		Lat:             lat,
		Lon:             lon,
		AcceptLanguage:  r.Header.Get("Accept-Language"), // browser locale
		UserAgent:       r.Header.Get("User-Agent"),
		IsMobile:        ua.Mobile(),
		IsBot:           ua.Bot(),
		MozillaVersion:  ua.Mozilla(),
		Platform:        ua.Platform(),
		OS:              ua.OS(),
		EngineName:      engineName,
		EngineVersion:   engineversion,
		BrowserName:     browserName,
		BrowserVersion:  browserVersion,
	}

	TrackEventDetails(w, r, cookie, "AdWords Tracking", click.Keyword+";"+click.Matchtype, click.Adposition, 0.)

	err := StoreClickInBigQuery(c, &click)
	if err != nil {
		common.Error("Error while storing click in BigQuery: %v", err)
	} else {
		common.Info("Click stored in BigQuery")
	}

	common.Info("Redirect to %v", redirectUrl)
	http.Redirect(w, r, redirectUrl, http.StatusFound)

	//URL template:
	// http://<your-domain>/tracking?cm={campaignid}&ag={adgroupid}&f={feeditemid}&tid={targetid}&lp={loc_physical_ms}&li={loc_interest_ms}&m={matchtype}&n={network}&d={device}&dm={devicemodel}&cr={creative}&k={keyword}&p={placement}&t={target}&p1={param1}&p2={param2}&r={random}&a={aceid}&ap={adposition}&i={ignore}&url={lpurl}&url2={lpurl+2}&url3={lpurl+3}&uurl={unescapedlpurl}&eurl={escapedlpurl}&eurl2={escapedlpurl+2}&eurl3={escapedlpurl+3}
	// Example: http://dev-dot-mygotome.appspot.com/tracking?k=test&m=test&url=http%3A%2F%2Fwww.example.com
}
