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

/*
Package csrf provides Cross-Site Request Forgery (CSRF) protection for web applications.

CSRF attacks occur when a malicious website causes a user's browser to perform
unwanted actions on a trusted site where the user is authenticated. This package
prevents such attacks by requiring a cryptographically secure token to be included
with all state-changing requests.

# Basic Usage

Create a token store and wrap your HTTP handler with the CSRF middleware:

	import "github.com/patdeg/common/csrf"

	func main() {
		store := csrf.NewTokenStore()
		mux := http.NewServeMux()
		mux.HandleFunc("/", homeHandler)

		// Wrap with CSRF protection
		handler := store.Middleware(mux)

		http.ListenAndServe(":8080", handler)
	}

# How It Works

The middleware operates differently based on the HTTP method:

Safe methods (GET, HEAD, OPTIONS):
  - Generate a new CSRF token
  - Set it in a cookie named "csrf_token"
  - Allow the request to proceed

State-changing methods (POST, PUT, DELETE, PATCH):
  - Require a valid CSRF token
  - Check both the cookie and the request (header or form field)
  - Validate the token matches and hasn't expired
  - Return 403 Forbidden if validation fails

# HTML Forms

Include the CSRF token as a hidden input field:

	<form method="POST" action="/submit">
	    <input type="hidden" name="csrf_token" value="{{.CSRFToken}}">
	    <input type="text" name="username">
	    <button type="submit">Submit</button>
	</form>

In your handler, pass the token to the template:

	func homeHandler(w http.ResponseWriter, r *http.Request) {
	    data := struct {
	        CSRFToken string
	    }{
	        CSRFToken: csrf.GetToken(r),
	    }
	    tmpl.Execute(w, data)
	}

Or use the common package helper:

	import "github.com/patdeg/common"

	data := struct {
	    CSRFToken string
	}{
	    CSRFToken: common.GetCSRFToken(r),
	}

# AJAX Requests

For AJAX requests, include the token in the X-CSRF-Token header:

	// Vanilla JavaScript
	const token = getCookie('csrf_token');

	fetch('/api/endpoint', {
	    method: 'POST',
	    headers: {
	        'Content-Type': 'application/json',
	        'X-CSRF-Token': token
	    },
	    body: JSON.stringify(data)
	});

	function getCookie(name) {
	    const value = `; ${document.cookie}`;
	    const parts = value.split(`; ${name}=`);
	    if (parts.length === 2) return parts.pop().split(';').shift();
	}

# HTMX Integration

For HTMX, add the token to all requests automatically:

	<script>
	document.body.addEventListener('htmx:configRequest', function(evt) {
	    const token = document.cookie
	        .split('; ')
	        .find(row => row.startsWith('csrf_token='))
	        ?.split('=')[1];

	    if (token) {
	        evt.detail.headers['X-CSRF-Token'] = token;
	    }
	});
	</script>

# jQuery Integration

	$.ajaxSetup({
	    beforeSend: function(xhr) {
	        const token = getCookie('csrf_token');
	        if (token) {
	            xhr.setRequestHeader('X-CSRF-Token', token);
	        }
	    }
	});

# Token Lifecycle

  - Tokens are generated with 256 bits of cryptographic randomness
  - Tokens expire after 24 hours
  - Expired tokens are automatically cleaned up every hour
  - Each token is validated using constant-time comparison to prevent timing attacks

# Security Considerations

  - The CSRF cookie has HttpOnly=false so JavaScript can read it for AJAX requests
  - The cookie uses SameSite=Strict for additional protection
  - The cookie is Secure (HTTPS only) except on localhost for development
  - Token validation uses constant-time comparison to prevent timing attacks
  - Tokens are cryptographically random (crypto/rand, not math/rand)

# Cookie Attributes

The CSRF token cookie is set with the following attributes:
  - Name: csrf_token
  - Path: /
  - MaxAge: 86400 (24 hours)
  - HttpOnly: false (JavaScript needs to read this)
  - Secure: true (except localhost)
  - SameSite: Strict

# Error Responses

The middleware returns HTTP 403 Forbidden with descriptive messages:
  - "CSRF token cookie missing" - No csrf_token cookie in request
  - "CSRF token missing from request" - No token in header or form
  - "CSRF token invalid or expired" - Token doesn't exist or has expired
  - "CSRF token validation failed" - Cookie and request tokens don't match

# Testing

When writing tests, you can either:

1. Make a GET request first to obtain a token:

	// Get token
	getReq := httptest.NewRequest("GET", "/", nil)
	getW := httptest.NewRecorder()
	handler.ServeHTTP(getW, getReq)

	var token string
	for _, c := range getW.Result().Cookies() {
	    if c.Name == "csrf_token" {
	        token = c.Value
	        break
	    }
	}

	// Use token in POST
	postReq := httptest.NewRequest("POST", "/", nil)
	postReq.Header.Set("X-CSRF-Token", token)
	postReq.AddCookie(&http.Cookie{Name: "csrf_token", Value: token})

2. Generate a token directly:

	store := csrf.NewTokenStore()
	token, _ := store.GenerateToken()

	req := httptest.NewRequest("POST", "/", nil)
	req.Header.Set("X-CSRF-Token", token)
	req.AddCookie(&http.Cookie{Name: "csrf_token", Value: token})

# Complete Example

	package main

	import (
	    "html/template"
	    "net/http"

	    "github.com/patdeg/common/csrf"
	)

	var tmpl = template.Must(template.New("form").Parse(`
	<!DOCTYPE html>
	<html>
	<head><title>CSRF Example</title></head>
	<body>
	    <form method="POST" action="/submit">
	        <input type="hidden" name="csrf_token" value="{{.CSRFToken}}">
	        <input type="text" name="username" placeholder="Username">
	        <button type="submit">Submit</button>
	    </form>
	</body>
	</html>
	`))

	func main() {
	    store := csrf.NewTokenStore()

	    mux := http.NewServeMux()
	    mux.HandleFunc("/", formHandler)
	    mux.HandleFunc("/submit", submitHandler)

	    handler := store.Middleware(mux)

	    http.ListenAndServe(":8080", handler)
	}

	func formHandler(w http.ResponseWriter, r *http.Request) {
	    data := struct {
	        CSRFToken string
	    }{
	        CSRFToken: csrf.GetToken(r),
	    }
	    tmpl.Execute(w, data)
	}

	func submitHandler(w http.ResponseWriter, r *http.Request) {
	    username := r.FormValue("username")
	    w.Write([]byte("Hello, " + username))
	}

# Performance

The CSRF middleware is lightweight and adds minimal overhead:
  - Token generation: ~100-200 ns/op (uses crypto/rand)
  - Token validation: ~50-100 ns/op (map lookup + constant-time compare)
  - Middleware overhead: ~500-1000 ns/op for GET requests

The token store uses a mutex-protected map, which scales well for most applications.
For high-traffic scenarios with millions of concurrent users, consider implementing
a distributed token store using Redis or Memcache.

# Thread Safety

All operations in this package are thread-safe:
  - TokenStore uses sync.RWMutex for concurrent access
  - Token generation uses crypto/rand which is thread-safe
  - Middleware can be called concurrently from multiple goroutines

*/
package csrf
