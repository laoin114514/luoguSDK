package luogusdk

import (
	"net/http"
	"net/url"
	"path/filepath"
	"testing"
)

func TestExportableCookieJarRoundTrip(t *testing.T) {
	jar, err := newExportableCookieJar()
	if err != nil {
		t.Fatalf("create jar: %v", err)
	}

	u, _ := url.Parse(luoguBaseURL)
	jar.SetCookies(u, []*http.Cookie{
		{Name: "_uid", Value: "12345", Domain: ".luogu.com.cn", Path: "/"},
		{Name: "__client_id", Value: "abc123", Domain: ".luogu.com.cn", Path: "/"},
	})

	data, err := jar.Export()
	if err != nil {
		t.Fatalf("export: %v", err)
	}

	jar2, _ := newExportableCookieJar()
	if err := jar2.Import(data); err != nil {
		t.Fatalf("import: %v", err)
	}

	cookies := jar2.Cookies(u)
	if len(cookies) != 2 {
		t.Fatalf("expected 2 cookies, got %d", len(cookies))
	}

	findCookie := func(name string) *http.Cookie {
		for _, c := range cookies {
			if c.Name == name {
				return c
			}
		}
		return nil
	}

	if c := findCookie("_uid"); c == nil || c.Value != "12345" {
		t.Error("_uid cookie not restored correctly")
	}
	if c := findCookie("__client_id"); c == nil || c.Value != "abc123" {
		t.Error("__client_id cookie not restored correctly")
	}
}

func TestSaveLoadCookiesToFile(t *testing.T) {
	jar, _ := newExportableCookieJar()
	u, _ := url.Parse(luoguBaseURL)
	jar.SetCookies(u, []*http.Cookie{
		{Name: "_uid", Value: "99999", Domain: ".luogu.com.cn", Path: "/"},
	})

	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "cookies.json")

	if err := saveCookies(jar, filePath); err != nil {
		t.Fatalf("save: %v", err)
	}

	jar2, _ := newExportableCookieJar()
	if err := loadCookies(jar2, filePath); err != nil {
		t.Fatalf("load: %v", err)
	}

	cookies := jar2.Cookies(u)
	if len(cookies) != 1 || cookies[0].Value != "99999" {
		t.Errorf("cookie not restored correctly")
	}
}
