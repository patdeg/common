package common

import (
	"html/template"
	"net/http"
	"net/url"
	"strings"

	"github.com/mssola/user_agent"
	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/appengine/v2/urlfetch"
)

// List of spammers & traffic builder - MUST BE IN LOWERCASE, Domains or Full Name
var SPAMMERS = map[string]bool{
	"4webmasters.org":            true,
	"abiente.ru":                 true,
	"allmetalworking.ru":         true,
	"archidom.info":              true,
	"best-seo-report.com":        true,
	"betonka.pro":                true,
	"biznesluxe.ru":              true,
	"burger-imperia.com":         true,
	"buttons-for-website.com":    true,
	"buyessaynow.biz":            true,
	"с.новым.годом.рф":           true,
	"darodar.com":                true,
	"e-buyeasy.com":              true,
	"erot.co":                    true,
	"event-tracking.com":         true,
	"fast-wordpress-start.com":   true,
	"finteks.ru":                 true,
	"fix-website-errors.com":     true,
	"floating-share-buttons.com": true,
	"free-social-buttons.com":    true,
	"get-free-traffic-now.com":   true,
	"hundejo.com":                true,
	"hvd-store.com":              true,
	"ifmo.ru":                    true,
	"interesnie-faktu.ru":        true,
	"kinoflux.net":               true,
	"kruzakivrazbor.ru":          true,
	"lenpipet.ru":                true,
	"letous.ru":                  true,
	"net-profits.xyz":            true,
	"pizza-imperia.com":          true,
	"pizza-tycoon.com":           true,
	"rankings-analytics.com":     true,
	"seo-2-0.com":                true,
	"share-buttons.xyz":          true,
	"success-seo.com":            true,
	"top1-seo-service.com":       true,
	"traffic2cash.xyz":           true,
	"traffic2money.com":          true,
	"trafficmonetizer.org":       true,
	"vashsvet.com":               true,
	"video-chat.in":              true,
	"videochat.tv.br":            true,
	"video--production.com":      true,
	"webmonetizer.net":           true,
	"website-stealer.nufaq.com":  true,
	"web-revenue.xyz":            true,
	"xrus.org":                   true,
	"zahvat.ru":                  true,
}

// List of untracked bots
var CUSTOM_BOTS_USER_AGENT = []string{
	"Mozilla/5.0 (compatible; Dataprovider/6.92; +https://www.dataprovider.com/)",
	"SSL Labs (https://www.ssllabs.com/about/assessment.html)",
	"CRAZYWEBCRAWLER 0.9.10, http://www.crazywebcrawler.com",
	"facebookexternalhit/1.1",
	"AdnormCrawler www.adnorm.com/crawler",
	"Mozilla/5.0 (compatible; Qwantify/2.2w; +https://www.qwant.com/)/*",
}

func GetServiceAccountClient(c context.Context) *http.Client {
	serviceAccountClient := &http.Client{
		Transport: &oauth2.Transport{
			Source: google.AppEngineTokenSource(c,
				"https://www.googleapis.com/auth/userinfo.email",
				"https://www.googleapis.com/auth/bigquery"),
			Base: &urlfetch.Transport{
				Context: c,
			},
		},
	}
	return serviceAccountClient
}

func GetContentByUrl(c context.Context, url string) ([]byte, error) {

	resp, err := GetServiceAccountClient(c).Get(url)
	if err != nil {
		return []byte{}, err
	}

	bodyResp := GetBodyResponse(resp)

	return bodyResp, nil

}

var messageHTML = `<html>
<head>
	<title>[[.Message]]</title>
  	<meta name="viewport" content="width=device-width, initial-scale=1">
	<meta http-equiv="refresh" content="[[.Timeout]]; url=[[.Redirect]]">
  	<link href="/lib/bootstrap-3.3.4/css/bootstrap.min.css" rel="stylesheet">
</head>
<body>
	<div class="container">
		<div class="row">
			<div class="col-xs-12 col-sm-12 col-md-12 col-lg-12">
				<h2>[[.Message]]</h2>
				Click <a href="[[.Redirect]]">here</a> to continue.
			</div>
		</div>
	</div>
</body>
</html>`

var messagelTemplate = template.
	Must(template.
		New("message.html").
		Delims("[[", "]]").
		Parse(messageHTML))

func MessageHandler(c context.Context, w http.ResponseWriter, message string, redirectUrl string, timeoutSec int64) {
	if err := messagelTemplate.Execute(w, template.FuncMap{
		"Message":  message,
		"Redirect": redirectUrl,
		"Timeout":  timeoutSec,
	}); err != nil {
		Error("Error with messagelTemplate: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func IsHacker(r *http.Request) bool {

	c := r.Context()

	if GetMemCacheString(c, "hacker-"+r.RemoteAddr) != "" {
		Info("IsHacker: Repeat IP %v", r.RemoteAddr)
		return true
	}

	if IsSpam(c, r.Referer()) {
		Info("IsHacker: Is Spam")
		SetMemCacheString(c, "hacker-"+r.RemoteAddr, "1", 4)
		return true
	}

	if r.UserAgent() == "" {
		Info("IsHacker: UserAgent empty")
		SetMemCacheString(c, "hacker-"+r.RemoteAddr, "1", 4)
		return true
	}

	if strings.Contains(r.URL.Path, ".php") {
		Info("IsHacker: Requesting .php page, rejecting: %v", r.URL.Path)
		SetMemCacheString(c, "hacker-"+r.RemoteAddr, "1", 4)
		return true
	}

	if strings.HasPrefix(r.URL.Path, "/wp/") {
		Info("IsHacker: WordPress path: %v", r.URL.Path)
		SetMemCacheString(c, "hacker-"+r.RemoteAddr, "1", 4)
		return true
	}

	if strings.HasPrefix(r.URL.Path, "/wp-content/") {
		Info("IsHacker: WordPress path: %v", r.URL.Path)
		SetMemCacheString(c, "hacker-"+r.RemoteAddr, "1", 4)
		return true
	}

	if strings.HasPrefix(r.URL.Path, "/blog/") {
		Info("IsHacker: Blog path: %v", r.URL.Path)
		SetMemCacheString(c, "hacker-"+r.RemoteAddr, "1", 4)
		return true
	}

	if strings.HasPrefix(r.URL.Path, "/wordpress/") {
		Info("IsHacker: WordPress path: %v", r.URL.Path)
		SetMemCacheString(c, "hacker-"+r.RemoteAddr, "1", 4)
		return true
	}

	if r.Header.Get("X-AppEngine-Country") == "UA" {
		if (r.Header.Get("X-AppEngine-City") == "lviv") || (r.Header.Get("X-AppEngine-City") == "kyiv") {
			Info("IsHacker: Ukraine traffic - City : %v", r.Header.Get("X-AppEngine-City"))
			SetMemCacheString(c, "hacker-"+r.RemoteAddr, "1", 4)
			return true
		}
	}

	return false

}

func IsMobile(useragent string) bool {
	ua := user_agent.New(useragent)
	return ua.Mobile()
}

func IsBot(useragent string) bool {
	ua := user_agent.New(useragent)
	browserName, _ := ua.Browser()
	return (ua.Bot()) || (browserName == "Java") || (StringInSlice(useragent, CUSTOM_BOTS_USER_AGENT))
}

func IsSpam(c context.Context, referer string) bool {
	if referer == "" {
		return false
	}
	referer = strings.ToLower(referer)

	if SPAMMERS[referer] {
		Debug("Referer in black list, rejecting: %v", referer)
		return true
	}
	u, err := url.Parse(referer)
	if err != nil {
		Error("Error parsing referer: %v", err)
		return false
	}

	segments := strings.Split(strings.ToLower(u.Host), ".")
	n := len(segments)
	if n < 2 {
		Error("Error with host '%v' from referer '%v', found %v segments", u.Host, referer, n)
		return false
	}

	domain := segments[n-2] + "." + segments[n-1]

	if SPAMMERS[domain] {
		Debug("Referer in black list, rejecting: %v", referer)
		return true
	}
	return false
}

func IsCrawler(r *http.Request) bool {
	userAgent := r.Header.Get("User-Agent")
	if strings.Contains(r.RequestURI, "_escaped_fragment_") {
		Info("Google Escaped Fragment: %v", r.RequestURI)
		return true
	}
	if strings.Contains(userAgent, "facebookexternalhit") {
		Info("Facebook bot: %v (%v)", r.RequestURI, userAgent)
		return true
	}
	if strings.Contains(userAgent, "LinkedInBot") {
		Info("Linkedin bot: %v (%v)", r.RequestURI, userAgent)
		return true
	}
	if strings.Contains(userAgent, "Googlebot") {
		Info("Google bot: %v (%v)", r.RequestURI, userAgent)
		return true
	}
	if strings.Contains(userAgent, "AdsBot") && strings.Contains(userAgent, "Google") {
		Info("Google AdsBot: %v (%v)", r.RequestURI, userAgent)
		return true
	}
	if strings.Contains(userAgent, "OrangeBot") {
		Info("OrangeBot bot: %v (%v)", r.RequestURI, userAgent)
		return true
	}
	if strings.Contains(userAgent, "Baiduspider") {
		Info("Baidu bot: %v (%v)", r.RequestURI, userAgent)
		return true
	}
	if strings.Contains(userAgent, "CRAZYWEBCRAWLER") {
		Info("CRAZYWEBCRAWLER bot: %v (%v)", r.RequestURI, userAgent)
		return true
	}
	if strings.Contains(userAgent, "CATExplorador") {
		Info("CATExplorador bot: %v (%v)", r.RequestURI, userAgent)
		return true
	}
	if (r.FormValue("SEO") != "") || (r.FormValue("FB") != "") {
		Info("SEO or FB parameter in url: %v", r.RequestURI)
		return true
	}
	ua := user_agent.New(r.Header.Get("User-Agent"))
	return ua.Bot()

}
