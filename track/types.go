package track

import (
	"time"

	"google.golang.org/appengine/v2/datastore"
)

type Visit struct {
	DatastoreKey   *datastore.Key `json:"datastoreKey" datastore:"-"`
	Cookie         string         `json:"cookie,omitempty"`
	Session        string         `json:"session,omitempty"`
	URI            string         `json:"uri,omitempty"`
	Referer        string         `json:"referer,omitempty"`
	Time           time.Time      `json:"time,omitempty"`
	Host           string         `json:"host,omitempty"`
	RemoteAddr     string         `json:"remoteAddr,omitempty"`
	InstanceId     string         `json:"instanceId,omitempty"`
	VersionId      string         `json:"versionId,omitempty"`
	Scheme         string         `json:"scheme,omitempty"`
	Country        string         `json:"country,omitempty"`
	Region         string         `json:"region,omitempty"`
	City           string         `json:"city,omitempty"`
	Lat            float64        `json:"lat,omitempty"`
	Lon            float64        `json:"lon,omitempty"`
	AcceptLanguage string         `json:"acceptLanguage,omitempty"`
	UserAgent      string         `json:"userAgent,omitempty"`
	IsMobile       bool           `json:"isMobile,omitempty"`
	IsBot          bool           `json:"isBot,omitempty"`
	MozillaVersion string         `json:"mozillaVersion,omitempty"`
	Platform       string         `json:"platform,omitempty"`
	OS             string         `json:"os,omitempty"`
	EngineName     string         `json:"engineName,omitempty"`
	EngineVersion  string         `json:"engineVersion,omitempty"`
	BrowserName    string         `json:"browserName,omitempty"`
	BrowserVersion string         `json:"browserVersion,omitempty"`
	Category       string         `json:"category,omitempty"`
	Action         string         `json:"action,omitempty"`
	Label          string         `json:"label,omitempty"`
	Value          float64        `json:"value,omitempty"`
}

type RobotPage struct {
	Time       time.Time `json:"time,omitempty"`
	Name       string    `json:"name,omitempty"`
	URL        string    `json:"url,omitempty"`
	URI        string    `json:"uri,omitempty"`
	Host       string    `json:"host,omitempty"`
	RemoteAddr string    `json:"remoteAddr,omitempty"`
	UserAgent  string    `json:"userAgent,omitempty"`
	Country    string    `json:"country,omitempty"`
	Region     string    `json:"region,omitempty"`
	City       string    `json:"city,omitempty"`
	BotName    string    `json:"botName,omitempty"`
	BotVersion string    `json:"botVersion,omitempty"`
}
