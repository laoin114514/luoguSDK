package luogusdk

import (
	"encoding/json"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"path/filepath"
)

// exportableCookie 可序列化的 cookie 结构
type exportableCookie struct {
	Name   string `json:"name"`
	Value  string `json:"value"`
	Domain string `json:"domain"`
	Path   string `json:"path"`
}

// ExportableCookieJar 包装 cookiejar.Jar，支持导出/导入
type ExportableCookieJar struct {
	jar *cookiejar.Jar
}

func newExportableCookieJar() (*ExportableCookieJar, error) {
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, err
	}
	return &ExportableCookieJar{jar: jar}, nil
}

func (j *ExportableCookieJar) SetCookies(u *url.URL, cookies []*http.Cookie) {
	j.jar.SetCookies(u, cookies)
}

func (j *ExportableCookieJar) Cookies(u *url.URL) []*http.Cookie {
	return j.jar.Cookies(u)
}

// Export 导出所有 cookie 为 JSON 字节
func (j *ExportableCookieJar) Export() ([]byte, error) {
	u, _ := url.Parse("https://www.luogu.com.cn/")
	cookies := j.jar.Cookies(u)
	exported := make([]exportableCookie, 0, len(cookies))
	for _, c := range cookies {
		exported = append(exported, exportableCookie{
			Name:   c.Name,
			Value:  c.Value,
			Domain: c.Domain,
			Path:   c.Path,
		})
	}
	return json.Marshal(exported)
}

// Import 从 JSON 字节导入 cookie
func (j *ExportableCookieJar) Import(data []byte) error {
	var cookies []exportableCookie
	if err := json.Unmarshal(data, &cookies); err != nil {
		return err
	}
	u, _ := url.Parse("https://www.luogu.com.cn/")
	for _, c := range cookies {
		httpCookie := &http.Cookie{
			Name:   c.Name,
			Value:  c.Value,
			Domain: c.Domain,
			Path:   c.Path,
		}
		j.jar.SetCookies(u, []*http.Cookie{httpCookie})
	}
	return nil
}

// defaultCookiePath 返回默认 cookie 文件路径
func defaultCookiePath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".luogu", "cookies.json"), nil
}

// saveCookies 保存 cookie 到文件
func saveCookies(jar *ExportableCookieJar, filePath string) error {
	data, err := jar.Export()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(filePath), 0700); err != nil {
		return err
	}
	return os.WriteFile(filePath, data, 0600)
}

// loadCookies 从文件加载 cookie
func loadCookies(jar *ExportableCookieJar, filePath string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}
	return jar.Import(data)
}
