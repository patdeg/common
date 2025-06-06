package track

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"mygotome/common"

	"github.com/mssola/user_agent"
	appengine "google.golang.org/appengine/v2"
	"google.golang.org/appengine/v2/datastore"
	"google.golang.org/appengine/v2/memcache"
)

func TrackVisit(w http.ResponseWriter, r *http.Request, cookie string) {
	c := r.Context()
	common.Info(">>>> TrackVisit")

	if _, err := memcache.Get(c, "visit-"+cookie); err == memcache.ErrCacheMiss {
		common.Info("Cookie not in memcache")
	} else if err != nil {
		common.Error("Error getting item: %v", err)
	} else {
		common.Info("Cookie in memcache, do not track visit again")
		return
	}

	ua := user_agent.New(r.Header.Get("User-Agent"))
	engineName, engineversion := ua.Engine()
	browserName, browserVersion := ua.Browser()

	if common.IsBot(r.Header.Get("User-Agent")) {
		common.Info("TrackVisit: Events from Bots, ignoring")
		return
	}

	if r.Header.Get("X-AppEngine-Country") == "ZZ" {
		common.Info("TrackVisit: Country is ZZ - most likely a bot, ignoring")
		return
	}

	lat := float64(0)
	lon := float64(0)
	latlon := strings.Split(r.Header.Get("X-AppEngine-CityLatLong"), ",")
	if len(latlon) == 2 {
		lat = common.S2F(latlon[0])
		lon = common.S2F(latlon[1])
	}

	session := ""
	item, err := memcache.Get(c, "session-"+cookie)
	if err != nil {
		session = strconv.FormatInt(time.Now().UnixNano(), 10) + "-" + cookie
		item = &memcache.Item{
			Key:        "session-" + cookie,
			Value:      []byte(session),
			Expiration: time.Minute * 30,
		}
		if err := memcache.Add(c, item); err == memcache.ErrNotStored {
			common.Info("TrackEventDetails: item with key %q already exists", item.Key)
		} else if err != nil {
			common.Error("TrackEventDetails: Error adding item: %v", err)
		}
	} else {
		session = common.B2S(item.Value)
		common.Info("TrackEventDetails: cookie in memcache: %v", session)
	}
	common.Info("TrackEventDetails: Session = %v", session)

	visit := &Visit{
		Cookie:         cookie,
		Session:        session,
		URI:            r.RequestURI,
		Referer:        r.Header.Get("Referer"),
		Time:           time.Now(),
		Host:           r.Host,
		RemoteAddr:     r.RemoteAddr,
		InstanceId:     appengine.InstanceID(),
		VersionId:      appengine.VersionID(c),
		Scheme:         r.URL.Scheme,
		Country:        r.Header.Get("X-AppEngine-Country"),
		Region:         r.Header.Get("X-AppEngine-Region"),
		City:           r.Header.Get("X-AppEngine-City"),
		Lat:            lat,
		Lon:            lon,
		AcceptLanguage: r.Header.Get("Accept-Language"),
		UserAgent:      r.Header.Get("User-Agent"),
		IsMobile:       ua.Mobile(),
		IsBot:          ua.Bot(),
		MozillaVersion: ua.Mozilla(),
		Platform:       ua.Platform(),
		OS:             ua.OS(),
		EngineName:     engineName,
		EngineVersion:  engineversion,
		BrowserName:    browserName,
		BrowserVersion: browserVersion,
	}

	err = StoreVisitInBigQuery(c, visit)
	if err != nil {
		common.Error("Error while storing visit in datastore: %v", err)
	} else {
		common.Info("Visit stored in datastore")
	}
}

func TrackEventDetails(w http.ResponseWriter, r *http.Request, cookie, category, action, label string, value float64) {
	defer func() {
		if rec := recover(); rec != nil {
			fmt.Printf("Recovered panic in TrackEventDetails: %v\n", rec)
		}
	}()

	// Use a background context associated with the request so work can continue
	// after the HTTP handler has returned.
	ctx := appengine.WithContext(context.Background(), r)
	reqCopy := r.Clone(ctx)

	go func() {
		c := ctx
		common.Info(">>>> TrackEventDetails")

		ua := user_agent.New(reqCopy.Header.Get("User-Agent"))
		engineName, engineversion := ua.Engine()
		browserName, browserVersion := ua.Browser()

		if common.IsBot(reqCopy.Header.Get("User-Agent")) {
			common.Info("TrackEventDetails: Events from Bots, ignoring")
			return
		}

		lat := float64(0)
		lon := float64(0)
		latlon := strings.Split(reqCopy.Header.Get("X-AppEngine-CityLatLong"), ",")
		if len(latlon) == 2 {
			lat = common.S2F(latlon[0])
			lon = common.S2F(latlon[1])
		}

		uniqueId := common.MD5(reqCopy.RemoteAddr + reqCopy.Header.Get("User-Agent"))
		session := ""
		item, err := memcache.Get(c, "s-"+uniqueId)
		if err != nil {
			session = strconv.FormatInt(time.Now().UnixNano(), 10) + "-" + uniqueId
			item = &memcache.Item{
				Key:        "s-" + uniqueId,
				Value:      []byte(session),
				Expiration: time.Minute * 30,
			}
			if err := memcache.Add(c, item); err == memcache.ErrNotStored {
				common.Info("TrackEventDetails: item with key %q already exists", item.Key)
			} else if err != nil {
				common.Error("TrackEventDetails: Error adding item: %v", err)
			}
		} else {
			session = common.B2S(item.Value)
			common.Info("TrackEventDetails: uniqueid in memcache: %v", session)
		}
		common.Info("TrackEventDetails: Unique Id = %v Session = %v", uniqueId, session)

		event := &Visit{
			Cookie:         cookie,
			Session:        session,
			URI:            reqCopy.RequestURI,
			Referer:        reqCopy.Header.Get("Referer"),
			Time:           time.Now(),
			Host:           reqCopy.Host,
			RemoteAddr:     reqCopy.RemoteAddr,
			InstanceId:     appengine.InstanceID(),
			VersionId:      appengine.VersionID(c),
			Scheme:         reqCopy.URL.Scheme,
			Country:        reqCopy.Header.Get("X-AppEngine-Country"),
			Region:         reqCopy.Header.Get("X-AppEngine-Region"),
			City:           reqCopy.Header.Get("X-AppEngine-City"),
			Lat:            lat,
			Lon:            lon,
			AcceptLanguage: reqCopy.Header.Get("Accept-Language"),
			UserAgent:      reqCopy.Header.Get("User-Agent"),
			IsMobile:       ua.Mobile(),
			IsBot:          ua.Bot(),
			MozillaVersion: ua.Mozilla(),
			Platform:       ua.Platform(),
			OS:             ua.OS(),
			EngineName:     engineName,
			EngineVersion:  engineversion,
			BrowserName:    browserName,
			BrowserVersion: browserVersion,
			Category:       common.Trunc500(category),
			Action:         common.Trunc500(action),
			Label:          common.Trunc500(label),
			Value:          value,
		}

		err = StoreEventInBigQuery(c, event)
		if err != nil {
			common.Error("Error while storing event in BigQuery: %v", err)
		} else {
			common.Info("Event stored in BigQuery")
		}
	}()
}

func TrackEvent(w http.ResponseWriter, r *http.Request, cookie string) {
	common.Info(">>>> TrackEvent")
	TrackEventDetails(w, r, cookie, r.FormValue("c"), r.FormValue("a"), r.FormValue("l"), common.S2F(r.FormValue("v")))
}

func TrackRobots(r *http.Request) {
	c := r.Context()
	common.Info(">>>> TrackRobots")

	userAgent := r.Header.Get("User-Agent")
	ua := user_agent.New(r.Header.Get("User-Agent"))
	botName, botVersion := ua.Browser()
	robotPage := RobotPage{
		Time:       time.Now(),
		URL:        r.URL.String(),
		URI:        r.RequestURI,
		Host:       r.Host,
		RemoteAddr: r.RemoteAddr,
		UserAgent:  userAgent,
		Country:    r.Header.Get("X-AppEngine-Country"),
		Region:     r.Header.Get("X-AppEngine-Region"),
		City:       r.Header.Get("X-AppEngine-City"),
		BotName:    botName,
		BotVersion: botVersion,
	}
	if strings.Contains(r.RequestURI, "_escaped_fragment_") {
		robotPage.Name = "escaped_fragment"
	}
	if strings.Contains(userAgent, "facebookexternalhit") {
		robotPage.Name = "Facebook"
	}
	if strings.Contains(userAgent, "LinkedInBot") {
		robotPage.Name = "Linkedin"
	}
	if strings.Contains(userAgent, "Googlebot") {
		robotPage.Name = "Google"
	}
	if strings.Contains(userAgent, "OrangeBot") {
		robotPage.Name = "Orange"
	}

	_, err := datastore.Put(c, datastore.NewIncompleteKey(c, "RobotPages", nil), &robotPage)
	if err != nil {
		common.Error("Error while storing robot page in datastore: %v", err)
	} else {
		common.Info("Robot page stored in datastore")
	}
}
