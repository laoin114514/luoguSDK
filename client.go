package luogusdk

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

const defaultUA = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/148.0.0.0 Safari/537.36 Edg/148.0.0.0"

// Client 洛谷 SDK 客户端
type Client struct {
	httpClient *http.Client
	cookieJar  *ExportableCookieJar
	csrfToken  string
	cookieFile string
	maxRetries int
	backoffFn  func(int) time.Duration
	userAgent  string

	Auth    *AuthService
	Problem *ProblemService
}

// ClientOption 客户端配置函数
type ClientOption func(*Client)

// WithCookieFile 设置 cookie 持久化文件路径
func WithCookieFile(path string) ClientOption {
	return func(c *Client) {
		c.cookieFile = path
	}
}

// WithRetry 设置重试参数
func WithRetry(maxRetries int, backoff func(int) time.Duration) ClientOption {
	return func(c *Client) {
		c.maxRetries = maxRetries
		if backoff != nil {
			c.backoffFn = backoff
		}
	}
}

// WithTimeout 设置 HTTP 超时
func WithTimeout(d time.Duration) ClientOption {
	return func(c *Client) {
		c.httpClient.Timeout = d
	}
}

// WithUserAgent 设置自定义 User-Agent
func WithUserAgent(ua string) ClientOption {
	return func(c *Client) {
		c.userAgent = ua
	}
}

// NewClient 创建新的洛谷客户端
func NewClient(opts ...ClientOption) (*Client, error) {
	jar, err := newExportableCookieJar()
	if err != nil {
		return nil, fmt.Errorf("create cookie jar: %w", err)
	}

	cookiePath, err := defaultCookiePath()
	if err != nil {
		cookiePath = "luogu_cookies.json"
	}

	c := &Client{
		cookieJar:  jar,
		cookieFile: cookiePath,
		maxRetries: 3,
		backoffFn:  defaultBackoff,
		userAgent:  defaultUA,
		httpClient: &http.Client{
			Jar:     jar,
			Timeout: 30 * time.Second,
		},
	}

	for _, opt := range opts {
		opt(c)
	}

	c.Auth = &AuthService{client: c}
	c.Problem = &ProblemService{client: c}

	// 尝试加载持久化的 cookie（文件不存在不算错误）
	if err := loadCookies(c.cookieJar, c.cookieFile); err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("load cookies from %s: %w", c.cookieFile, err)
	}

	return c, nil
}

// newRequest 创建带默认请求头的 HTTP 请求
func (c *Client) newRequest(method, path string, body interface{}) (*http.Request, error) {
	url := luoguBaseURL + strings.TrimPrefix(path, "/")

	var bodyReader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("marshal request body: %w", err)
		}
		bodyReader = bytes.NewReader(data)
	}

	req, err := http.NewRequest(method, url, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("User-Agent", c.userAgent)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("Referer", luoguBaseURL)

	if c.csrfToken != "" && method != "GET" {
		req.Header.Set("X-CSRF-TOKEN", c.csrfToken)
	}

	return req, nil
}

// do 执行 HTTP 请求，带重试逻辑
func (c *Client) do(req *http.Request) (*http.Response, error) {
	var lastErr error
	for attempt := 0; attempt <= c.maxRetries; attempt++ {
		resp, err := c.httpClient.Do(req)
		if err != nil {
			if shouldRetry(err) && attempt < c.maxRetries {
				time.Sleep(c.backoffFn(attempt))
				lastErr = err
				continue
			}
			return nil, &NetworkError{Err: err}
		}

		if shouldRetryStatus(resp.StatusCode) && attempt < c.maxRetries {
			resp.Body.Close()
			time.Sleep(c.backoffFn(attempt))
			continue
		}

		return resp, nil
	}
	return nil, &NetworkError{Err: fmt.Errorf("max retries exceeded: %w", lastErr)}
}

// get 发送 GET 请求
func (c *Client) get(path string) (*http.Response, error) {
	req, err := c.newRequest("GET", path, nil)
	if err != nil {
		return nil, err
	}
	return c.do(req)
}

// post 发送 POST 请求
func (c *Client) post(path string, body interface{}) (*http.Response, error) {
	req, err := c.newRequest("POST", path, body)
	if err != nil {
		return nil, err
	}
	return c.do(req)
}

// parseBody 解析响应体 JSON
func parseBody(resp *http.Response, v interface{}) error {
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read response body: %w", err)
	}
	if err := json.Unmarshal(data, v); err != nil {
		preview := string(data)
		if len(preview) > 300 {
			preview = preview[:300] + "..."
		}
		return fmt.Errorf("unmarshal response (status=%d, body=%s): %w", resp.StatusCode, preview, err)
	}
	return nil
}

// parseLentilleContext 从 HTML 页面中提取 <script id="lentille-context"> 内的 JSON 数据
func parseLentilleContext(resp *http.Response, v interface{}) error {
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return fmt.Errorf("parse HTML: %w", err)
	}

	jsonStr := doc.Find("script#lentille-context").Text()
	if jsonStr == "" {
		return fmt.Errorf("lentille-context script not found in page")
	}

	if err := json.Unmarshal([]byte(jsonStr), v); err != nil {
		return fmt.Errorf("unmarshal lentille-context: %w", err)
	}
	return nil
}

// refreshCSRF 从首页获取 CSRF token
func (c *Client) refreshCSRF() error {
	resp, err := c.get("/")
	if err != nil {
		return &CSRFError{Err: err}
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return &CSRFError{Err: fmt.Errorf("parse HTML: %w", err)}
	}

	token, exists := doc.Find("meta[name=csrf-token]").Attr("content")
	if !exists {
		return &CSRFError{Err: fmt.Errorf("csrf token not found in page")}
	}

	c.csrfToken = token
	return nil
}

// setCSRF 手动设置 CSRF token（用于测试）
func (c *Client) setCSRF(token string) {
	c.csrfToken = token
}

// verifyAuth 校验当前 cookie 是否仍有效
func (c *Client) verifyAuth() error {
	resp, err := c.get("/api/user/current")
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
		return &UnauthorizedError{}
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("verify auth: unexpected status %d", resp.StatusCode)
	}

	var result struct {
		Data struct {
			UID int `json:"uid"`
		} `json:"data"`
	}
	if err := parseBody(resp, &result); err != nil {
		return err
	}
	if result.Data.UID == 0 {
		return &UnauthorizedError{}
	}
	return nil
}

// saveCookiesToFile 持久化当前 cookie
func (c *Client) saveCookiesToFile() error {
	return saveCookies(c.cookieJar, c.cookieFile)
}
