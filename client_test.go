package luogusdk

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestClientNewRequestHeaders(t *testing.T) {
	jar, _ := newExportableCookieJar()
	c := &Client{
		cookieJar:  jar,
		csrfToken:  "test-csrf-token",
		maxRetries: 3,
		backoffFn:  defaultBackoff,
		userAgent:  "test-ua",
		httpClient: &http.Client{Jar: jar},
	}

	req, err := c.newRequest("POST", "/test", map[string]string{"key": "value"})
	if err != nil {
		t.Fatalf("newRequest: %v", err)
	}

	if ct := req.Header.Get("Content-Type"); ct != "application/json" {
		t.Errorf("Content-Type = %q, want application/json", ct)
	}
	if ref := req.Header.Get("Referer"); ref != luoguBaseURL {
		t.Errorf("Referer = %q, want %s", ref, luoguBaseURL)
	}
	if csrf := req.Header.Get("X-CSRF-TOKEN"); csrf != "test-csrf-token" {
		t.Errorf("X-CSRF-TOKEN = %q, want test-csrf-token", csrf)
	}
	if ua := req.Header.Get("User-Agent"); ua == "" {
		t.Error("User-Agent should not be empty")
	}
}

func TestClientNewRequestNoCSRFForGET(t *testing.T) {
	jar, _ := newExportableCookieJar()
	c := &Client{
		cookieJar:  jar,
		csrfToken:  "test-csrf-token",
		maxRetries: 3,
		backoffFn:  defaultBackoff,
		userAgent:  "test-ua",
		httpClient: &http.Client{Jar: jar},
	}

	req, err := c.newRequest("GET", "/test", nil)
	if err != nil {
		t.Fatalf("newRequest: %v", err)
	}

	if req.Header.Get("X-CSRF-TOKEN") != "" {
		t.Error("GET requests should not have X-CSRF-TOKEN header")
	}
}

func TestClientNoRetryOn4xx(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.WriteHeader(http.StatusBadRequest)
	}))
	defer server.Close()

	jar, _ := newExportableCookieJar()
	c := &Client{
		cookieJar:  jar,
		maxRetries: 3,
		backoffFn:  func(int) time.Duration { return 0 },
		userAgent:  "test-ua",
		httpClient: &http.Client{Jar: jar},
	}

	req, _ := http.NewRequest("GET", server.URL+"/test", nil)
	resp, _ := c.do(req)
	resp.Body.Close()

	if callCount != 1 {
		t.Errorf("4xx should not retry, called %d times", callCount)
	}
}

func TestClientRetryOn5xx(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	jar, _ := newExportableCookieJar()
	c := &Client{
		cookieJar:  jar,
		maxRetries: 2,
		backoffFn:  func(int) time.Duration { return 0 },
		userAgent:  "test-ua",
		httpClient: &http.Client{Jar: jar},
	}

	req, _ := http.NewRequest("GET", server.URL+"/test", nil)
	resp, _ := c.do(req)
	resp.Body.Close()

	if callCount != 3 {
		t.Errorf("5xx should retry, expected 3 calls, got %d", callCount)
	}
}
