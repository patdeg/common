// Package common provides shared helpers used across the application.
package common

import (
	"net"
	"net/http"
	"strconv"
	"time"
)

type Visitor struct {
	Key                   string `json:"key,omitempty"`
	Cookie                string `json:"cookie,omitempty"`
	Host                  string `json:"host,omitempty"`
	CreatedTimestamp      string `json:"createdTimestamp,omitempty"`
	CreatedIP             string `json:"createdIP,omitempty"`
	CreatedReferer        string `json:"createdReferer,omitempty"`
	CreatedCountry        string `json:"createdCountry,omitempty"`
	CreatedRegion         string `json:"createdRegion,omitempty"`
	CreatedCity           string `json:"createdCity,omitempty"`
	CreatedUserAgent      string `json:"createdUserAgent,omitempty"`
	CreatedIsMobile       bool   `json:"createdIsMobile,omitempty"`
	CreatedIsBot          bool   `json:"createdIsBot,omitempty"`
	CreatedMozillaVersion string `json:"createdMozillaVersion,omitempty"`
	CreatedPlatform       string `json:"createdPlatform,omitempty"`
	CreatedOS             string `json:"createdOS,omitempty"`
	CreatedEngineName     string `json:"createdEngineName,omitempty"`
	CreatedEngineVersion  string `json:"createdEngineVersion,omitempty"`
	CreatedBrowserName    string `json:"createdBrowserName,omitempty"`
	CreatedBrowserVersion string `json:"createdBrowserVersion,omitempty"`
}

// ClearCookie removes the visitor ID cookie from the client.
func ClearCookie(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     "ID",
		Value:    "",
		Path:     "/",
		Expires:  time.Now(),
		SameSite: http.SameSiteLaxMode,
	})
}

// DoesCookieExists reports whether the visitor ID cookie is present.
func DoesCookieExists(r *http.Request) bool {
	cookie, err := r.Cookie("ID")
	if err != nil || cookie == nil || cookie.Value == "" {
		return false
	}
	return true
}

// GetCookieID retrieves the visitor ID cookie or creates a new one if missing.
//
// The created cookie is secured with the following attributes:
//   - Path is always set to "/" so the ID is sent for all application routes.
//   - Expires is 30 days in the future, giving the cookie a one month lifetime.
//   - HttpOnly and Secure are true to prevent JavaScript access and require HTTPS.
//   - SameSite is Lax to protect against CSRF while allowing normal navigation.
//   - Domain is only set when the host is neither localhost nor 127.0.0.1.
func GetCookieID(w http.ResponseWriter, r *http.Request) string {
	Debug(">>>> GetCookieID")

	var id string
	cookie, err := r.Cookie("ID")
	Debug("ID cookie: %v", cookie)
	if err != nil || cookie == nil || cookie.Value == "" {
		Debug("Error: %v", err)
		Debug("New Cookie...")
		ts := strconv.FormatInt(time.Now().UnixNano(), 10)
		id = MD5(ts + r.RemoteAddr)
		host := r.Host
		if h, _, err := net.SplitHostPort(host); err == nil {
			host = h
		}
		ck := &http.Cookie{
			Name:     "ID",
			Value:    id,
			Path:     "/",
			Expires:  time.Now().Add(time.Hour * 24 * 30),
			HttpOnly: true,
			Secure:   true,
			SameSite: http.SameSiteLaxMode,
		}
		if host != "localhost" && host != "127.0.0.1" {
			ck.Domain = host
		}
		http.SetCookie(w, ck)
		Debug("New Cookie = %v", id)
		/*
			key := datastore.NewKey(c, "Visitors", id, 0, nil)
			ua := user_agent.New(r.Header.Get("User-Agent"))
			engineName, engineversion := ua.Engine()
			browserName, browserVersion := ua.Browser()
			newVisitor := Visitor{
				Key:                   id,
				Cookie:                id,
				CreatedTimestamp:      ts,
				Host:                  r.Host,
				CreatedIP:             r.RemoteAddr,
				CreatedReferer:        r.Header.Get("Referer"),
				CreatedCountry:        r.Header.Get("X-AppEngine-Country"),
				CreatedRegion:         r.Header.Get("X-AppEngine-Region"),
				CreatedCity:           r.Header.Get("X-AppEngine-City"),
				CreatedUserAgent:      r.Header.Get("User-Agent"),
				CreatedIsMobile:       ua.Mobile(),
				CreatedIsBot:          ua.Bot(),
				CreatedMozillaVersion: ua.Mozilla(),
				CreatedPlatform:       ua.Platform(),
				CreatedOS:             ua.OS(),
				CreatedEngineName:     engineName,
				CreatedEngineVersion:  engineversion,
				CreatedBrowserName:    browserName,
				CreatedBrowserVersion: browserVersion,
			}
			key, err := datastore.Put(c, key, &newVisitor)
			if err != nil {
				log.Errorf(c, "Error while storing cookie %v in datastore: %v", id, err)
			} else {
				log.Infof(c, "New visitor %v stored in datastore under key %v", id,
					key.IntID())
			}
		*/
	} else {
		id = cookie.Value
		Debug("Existing ID Cookie = %v", id)
	}
	return id
}
