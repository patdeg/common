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

// This file provides helper functions used across HTTP handlers to detect
// bots and referrer spam as well as a tiny HTML template for timed redirects.
// The anti-spam logic relies on a blacklist of known domains while bot
// detection combines heuristics and the user_agent package. The message
// template renders a simple page that redirects after a given timeout.

import (
	"crypto/sha256"
	"encoding/hex"
	"html/template"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/mssola/user_agent"
	"github.com/patdeg/common/gcp"
	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/appengine/v2/urlfetch"
)

// spamDomainList holds known referrer spam domains. The list was compiled from
// various public blacklists around 2015 and must be kept in lowercase.
var spamDomainList = []string{
	"4webmasters.org",
	"abiente.ru",
	"allmetalworking.ru",
	"archidom.info",
	"best-seo-report.com",
	"betonka.pro",
	"biznesluxe.ru",
	"burger-imperia.com",
	"buttons-for-website.com",
	"buyessaynow.biz",
	"с.новым.годом.рф",
	"darodar.com",
	"e-buyeasy.com",
	"erot.co",
	"event-tracking.com",
	"fast-wordpress-start.com",
	"finteks.ru",
	"fix-website-errors.com",
	"floating-share-buttons.com",
	"free-social-buttons.com",
	"get-free-traffic-now.com",
	"hundejo.com",
	"hvd-store.com",
	"ifmo.ru",
	"interesnie-faktu.ru",
	"kinoflux.net",
	"kruzakivrazbor.ru",
	"lenpipet.ru",
	"letous.ru",
	"net-profits.xyz",
	"pizza-imperia.com",
	"pizza-tycoon.com",
	"rankings-analytics.com",
	"seo-2-0.com",
	"share-buttons.xyz",
	"success-seo.com",
	"top1-seo-service.com",
	"traffic2cash.xyz",
	"traffic2money.com",
	"trafficmonetizer.org",
	"vashsvet.com",
	"video-chat.in",
	"videochat.tv.br",
	"video--production.com",
	"webmonetizer.net",
	"website-stealer.nufaq.com",
	"web-revenue.xyz",
	"xrus.org",
	"zahvat.ru",
}

// SPAMMERS maps domains from spamDomainList for quick lookup.
var SPAMMERS map[string]bool

func init() {
	SPAMMERS = make(map[string]bool, len(spamDomainList))
	for _, d := range spamDomainList {
		SPAMMERS[d] = true
	}
}

// botUserAgents lists user agent strings for crawlers that are not detected by
// the user_agent package. The entries originate from server logs.
var botUserAgents = []string{
	"Mozilla/5.0 (compatible; Dataprovider/6.92; +https://www.dataprovider.com/)",
	"SSL Labs (https://www.ssllabs.com/about/assessment.html)",
	"CRAZYWEBCRAWLER 0.9.10, http://www.crazywebcrawler.com",
	"facebookexternalhit/1.1",
	"AdnormCrawler www.adnorm.com/crawler",
	"Mozilla/5.0 (compatible; Qwantify/2.2w; +https://www.qwant.com/)/*",
}

// CUSTOM_BOTS_USER_AGENT is kept for backward compatibility.
var CUSTOM_BOTS_USER_AGENT = botUserAgents

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

// ipHashSalt is used for hashing IP addresses to protect user privacy
// In production, this should be set via the IP_HASH_SALT environment variable
var ipHashSalt string

func init() {
	ipHashSalt = os.Getenv("IP_HASH_SALT")
	if ipHashSalt == "" {
		// Generate a default salt (WARNING: in production, use environment variable)
		ipHashSalt = "default-salt-change-in-production"
		Error("IP_HASH_SALT environment variable not set, using default (NOT FOR PRODUCTION)")
	}
}

// HashIP creates a one-way hash of an IP address for privacy-compliant logging
// This function is used to comply with GDPR/CCPA requirements by not storing
// plain IP addresses in logs. The hash is consistent for the same IP within
// a deployment (using the IP_HASH_SALT), allowing for session tracking while
// protecting user privacy.
func HashIP(ip string) string {
	h := sha256.Sum256([]byte(ip + ipHashSalt))
	// Use first 8 bytes (16 hex chars) for shorter logs while maintaining uniqueness
	return hex.EncodeToString(h[:8])
}

// MessageHandler renders a minimal HTML page using messagelTemplate. The page
// displays a message and performs a client-side redirect to redirectUrl after
// timeoutSec seconds via a meta-refresh tag.
func MessageHandler(c context.Context, w http.ResponseWriter, message string, redirectUrl string, timeoutSec int64) {
	// Use a struct for template data instead of FuncMap to ensure proper HTML escaping
	data := struct {
		Message  string
		Redirect string
		Timeout  int64
	}{
		Message:  message,
		Redirect: redirectUrl,
		Timeout:  timeoutSec,
	}
	if err := messagelTemplate.Execute(w, data); err != nil {
		Error("Error with messagelTemplate: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func IsHacker(r *http.Request) bool {

	c := r.Context()

	// Quickly reject IPs that were previously flagged as malicious.
	// Hash IP address for privacy compliance (GDPR/CCPA)
	ipHash := HashIP(r.RemoteAddr)

	if gcp.GetMemCacheString(c, "hacker-"+ipHash) != "" {
		// Don't log actual IP address - use hash for privacy
		Debug("IsHacker: Repeat request from cached IP hash")
		return true
	}

	// Block requests with spammy referrers.
	if IsSpam(c, r.Referer()) {
		Info("IsHacker: Is Spam")
		gcp.SetMemCacheString(c, "hacker-"+ipHash, "1", 4)
		return true
	}

	// Empty user agents are suspicious.
	if r.UserAgent() == "" {
		Info("IsHacker: UserAgent empty")
		gcp.SetMemCacheString(c, "hacker-"+ipHash, "1", 4)
		return true
	}

	// Reject attempts to access PHP scripts.
	if strings.Contains(r.URL.Path, ".php") {
		Info("IsHacker: Requesting .php page, rejecting: %v", r.URL.Path)
		gcp.SetMemCacheString(c, "hacker-"+ipHash, "1", 4)
		return true
	}

	// WordPress probing is treated as malicious.
	if strings.HasPrefix(r.URL.Path, "/wp/") {
		Info("IsHacker: WordPress path: %v", r.URL.Path)
		gcp.SetMemCacheString(c, "hacker-"+ipHash, "1", 4)
		return true
	}

	if strings.HasPrefix(r.URL.Path, "/wp-content/") {
		Info("IsHacker: WordPress path: %v", r.URL.Path)
		gcp.SetMemCacheString(c, "hacker-"+ipHash, "1", 4)
		return true
	}

	// Old blog paths are not served anymore; accessing them is suspicious.
	if strings.HasPrefix(r.URL.Path, "/blog/") {
		Info("IsHacker: Blog path: %v", r.URL.Path)
		gcp.SetMemCacheString(c, "hacker-"+ipHash, "1", 4)
		return true
	}

	if strings.HasPrefix(r.URL.Path, "/wordpress/") {
		Info("IsHacker: WordPress path: %v", r.URL.Path)
		gcp.SetMemCacheString(c, "hacker-"+ipHash, "1", 4)
		return true
	}

	// Geographic-based filtering for specific abuse patterns
	// NOTE: Geographic blocking may have GDPR implications - ensure legal basis
	// is documented and justified (e.g., fraud prevention, security incident response)
	// Consider: Is this blocking still necessary? Review with legal/compliance team.
	if r.Header.Get("X-AppEngine-Country") == "UA" {
		if (r.Header.Get("X-AppEngine-City") == "lviv") || (r.Header.Get("X-AppEngine-City") == "kyiv") {
			// Log aggregate data only, not specific city for privacy
			Info("IsHacker: Suspicious pattern detected from region: UA")
			gcp.SetMemCacheString(c, "hacker-"+ipHash, "1", 4)
			return true
		}
	}

	return false

}

// IsMobile returns true when the user agent represents a mobile device.
// It uses the github.com/mssola/user_agent library to parse the UA string.
func IsMobile(useragent string) bool {
	ua := user_agent.New(useragent)
	return ua.Mobile()
}

// IsBot reports whether the agent is a known crawler. The heuristics rely on
// the user_agent library and a small list of custom user agents.
func IsBot(useragent string) bool {
	ua := user_agent.New(useragent)
	browserName, _ := ua.Browser()
	return (ua.Bot()) || (browserName == "Java") || (StringInSlice(useragent, CUSTOM_BOTS_USER_AGENT))
}

func IsSpam(c context.Context, referer string) bool {
	// Empty referrers are ignored.
	if referer == "" {
		return false
	}
	referer = strings.ToLower(referer)

	// First check for an exact match against known spam hosts.
	if SPAMMERS[referer] {
		Debug("Referer in black list, rejecting: %v", referer)
		return true
	}
	u, err := url.Parse(referer)
	if err != nil {
		Error("Error parsing referer: %v", err)
		return false
	}

	// Check the registrable domain part of the referrer as well.
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
	// Look for explicit crawler indicators first.
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
	// Some bots mark the request with specific query parameters.
	if (r.FormValue("SEO") != "") || (r.FormValue("FB") != "") {
		Info("SEO or FB parameter in url: %v", r.RequestURI)
		return true
	}
	// Fallback to the user_agent library for generic bot detection.
	ua := user_agent.New(r.Header.Get("User-Agent"))
	return ua.Bot()

}

// GetCSRFToken returns the CSRF token from the request for template injection
// This is a convenience wrapper around csrf.GetToken() for use in templates
func GetCSRFToken(r *http.Request) string {
	cookie, err := r.Cookie("csrf_token")
	if err != nil {
		return ""
	}
	return cookie.Value
}

// SecurityHeadersMiddleware adds security headers to all HTTP responses
// This middleware should be applied to all HTTP handlers to protect against
// common web vulnerabilities including XSS, clickjacking, and MIME-sniffing attacks.
//
// Headers set:
//   - X-Frame-Options: DENY (prevents clickjacking)
//   - X-Content-Type-Options: nosniff (prevents MIME-sniffing)
//   - X-XSS-Protection: 1; mode=block (legacy XSS protection)
//   - Content-Security-Policy: restricts resource loading (configurable via env)
//   - Referrer-Policy: strict-origin-when-cross-origin (controls referrer info)
//   - Strict-Transport-Security: enforces HTTPS (when using HTTPS)
//   - Permissions-Policy: restricts browser features
//
// Usage:
//
//	mux := http.NewServeMux()
//	mux.HandleFunc("/", homeHandler)
//	secureHandler := common.SecurityHeadersMiddleware(mux)
//	http.ListenAndServe(":8080", secureHandler)
func SecurityHeadersMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Prevent clickjacking by blocking iframe embedding
		w.Header().Set("X-Frame-Options", "DENY")

		// Prevent MIME type sniffing
		w.Header().Set("X-Content-Type-Options", "nosniff")

		// Enable XSS filter in legacy browsers
		w.Header().Set("X-XSS-Protection", "1; mode=block")

		// Content Security Policy - adjust based on application needs
		// This is a strict baseline policy
		csp := "default-src 'self'; " +
			"script-src 'self'; " +
			"style-src 'self' 'unsafe-inline'; " + // unsafe-inline needed for some frameworks
			"img-src 'self' data: https:; " +
			"font-src 'self'; " +
			"connect-src 'self'; " +
			"frame-ancestors 'none'; " +
			"base-uri 'self'; " +
			"form-action 'self'"

		// Allow override via environment variable for flexibility
		if envCSP := os.Getenv("CONTENT_SECURITY_POLICY"); envCSP != "" {
			csp = envCSP
		}
		w.Header().Set("Content-Security-Policy", csp)

		// Control referrer information leakage
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")

		// HSTS (only on HTTPS connections)
		if r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https" {
			// max-age=31536000 is 1 year
			w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains; preload")
		}

		// Permissions Policy (restrict browser features)
		w.Header().Set("Permissions-Policy", "geolocation=(), microphone=(), camera=(), payment=()")

		next.ServeHTTP(w, r)
	})
}
