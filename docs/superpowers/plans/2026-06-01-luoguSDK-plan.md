# LuoguSDK 实现计划

> **对 agentic workers：** 必须使用 superpowers:subagent-driven-development（推荐）或 superpowers:executing-plans 按任务逐步实现。步骤使用 `- [ ]` 复选框语法追踪。

**目标：** 实现洛谷平台 Go SDK 第一阶段——认证（登录/验证码/Cookie持久化/重试）和题目只读操作（获取/搜索/题解/翻译）。

**架构：** Client + Service 分层。Client 持有 HTTP 会话、CSRF token、CookieJar；AuthService 和 ProblemService 通过 Client 发起请求。所有非 GET 请求自动注入 `X-CSRF-TOKEN` 和 `Referer`。

**技术栈：** Go 1.26、标准库 `net/http`、`github.com/PuerkitoBio/goquery`（HTML解析）、Python `ddddocr`（OCR）

---

### Task 1: 项目初始化与依赖

**文件：**
- 修改: `go.mod`
- 创建: 无新文件

- [ ] **Step 1: 更新 go.mod 添加依赖**

当前 `go.mod`:
```
module laoin114514/luoguSDK

go 1.26.2
```

替换为:
```
module laoin114514/luoguSDK

go 1.26.2

require github.com/PuerkitoBio/goquery v1.10.3

require (
	github.com/andybalholm/cascadia v1.3.3 // indirect
	golang.org/x/net v0.37.0 // indirect
)
```

- [ ] **Step 2: 下载依赖**

```bash
cd "c:/Users/29084/Desktop/code/go/luoguSDK" && go mod tidy
```

期望: 成功下载依赖，无错误输出。

- [ ] **Step 3: 提交**

```bash
git add go.mod go.sum && git commit -m "chore: add goquery dependency for HTML parsing"
```

---

### Task 2: 错误类型

**文件：**
- 创建: `errors.go`

- [ ] **Step 1: 创建 errors.go**

```go
package luogusdk

import "fmt"

// AuthError 登录/认证失败
type AuthError struct {
	Code    int
	Message string
}

func (e *AuthError) Error() string {
	return fmt.Sprintf("auth error [%d]: %s", e.Code, e.Message)
}

// CSRFError CSRF token 获取或过期
type CSRFError struct {
	Err error
}

func (e *CSRFError) Error() string {
	return fmt.Sprintf("csrf error: %v", e.Err)
}

func (e *CSRFError) Unwrap() error {
	return e.Err
}

// NetworkError 网络请求失败
type NetworkError struct {
	Err error
}

func (e *NetworkError) Error() string {
	return fmt.Sprintf("network error: %v", e.Err)
}

func (e *NetworkError) Unwrap() error {
	return e.Err
}

// UnauthorizedError 未登录调用需认证的 API
type UnauthorizedError struct{}

func (e *UnauthorizedError) Error() string {
	return "unauthorized: please login first"
}
```

- [ ] **Step 2: 编译验证**

```bash
cd "c:/Users/29084/Desktop/code/go/luoguSDK" && go build ./...
```

- [ ] **Step 3: 提交**

```bash
git add errors.go && git commit -m "feat: add error types (AuthError, CSRFError, NetworkError, UnauthorizedError)"
```

---

### Task 3: 公共类型定义

**文件：**
- 创建: `types.go`

- [ ] **Step 1: 创建 types.go**

```go
package luogusdk

// CaptchaSolver 验证码求解器，接收 JPEG 图片字节，返回识别结果
type CaptchaSolver func(image []byte) (string, error)

// LoginRequest 登录请求体
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Captcha  string `json:"captcha"`
}

// LoginResponse 登录响应
type LoginResponse struct {
	UID      int    `json:"uid"`
	ClientID string `json:"client_id"`
}

// Problem 题目详情
type Problem struct {
	PID          string   `json:"pid"`
	Title        string   `json:"title"`
	Difficulty   int      `json:"difficulty"`
	Background   string   `json:"background"`
	Description  string   `json:"description"`
	InputFormat  string   `json:"inputFormat"`
	OutputFormat string   `json:"outputFormat"`
	Samples      []Sample `json:"samples"`
	Hints        []string `json:"hints"`
	Tags         []Tag    `json:"tags"`
	TimeLimit    int      `json:"timeLimit"`
	MemoryLimit  int      `json:"memoryLimit"`
}

// Sample 输入输出样例
type Sample struct {
	Input  string `json:"input"`
	Output string `json:"output"`
}

// Tag 标签
type Tag struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// SearchParams 题目搜索参数
type SearchParams struct {
	Keyword    string
	Difficulty []int
	Tags       []int
	Page       int
	PageSize   int
}

// SearchResult 搜索结果
type SearchResult struct {
	Problems []ProblemSummary
	Total    int
	Page     int
}

// ProblemSummary 题目摘要
type ProblemSummary struct {
	PID        string
	Title      string
	Difficulty int
	Tags       []Tag
}

// Solution 题解详情
type Solution struct {
	ID      int
	Author  UserInfo
	Title   string
	Content string
	Likes   int
}

// SolutionList 题解列表
type SolutionList struct {
	Solutions []SolutionSummary
	Total     int
	Page      int
}

// SolutionSummary 题解摘要
type SolutionSummary struct {
	ID     int
	Title  string
	Author UserInfo
	Likes  int
}

// Translation 题目翻译
type Translation struct {
	Language    string
	Title       string
	Description string
}

// UserInfo 用户信息
type UserInfo struct {
	UID    int    `json:"uid"`
	Name   string `json:"name"`
	Avatar string `json:"avatar"`
}
```

- [ ] **Step 2: 编译验证**

```bash
cd "c:/Users/29084/Desktop/code/go/luoguSDK" && go build ./...
```

- [ ] **Step 3: 提交**

```bash
git add types.go && git commit -m "feat: add shared types for auth and problem APIs"
```

---

### Task 4: Cookie 持久化

**文件：**
- 创建: `cookiestore.go`

- [ ] **Step 1: 创建 cookiestore.go**

```go
package luogusdk

import (
	"encoding/json"
	"net/http"
	"net/http/cookiejar"
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

func (j *ExportableCookieJar) SetCookies(u *http.Request, cookies []*http.Cookie) {
	j.jar.SetCookies(u, cookies)
}

func (j *ExportableCookieJar) Cookies(u *http.Request) []*http.Cookie {
	return j.jar.Cookies(u)
}

// Export 导出所有 cookie 为 JSON 字节
func (j *ExportableCookieJar) Export() ([]byte, error) {
	// 使用洛谷首页 URL 获取所有相关 cookie
	u, _ := http.NewRequest("GET", "https://www.luogu.com.cn/", nil)
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
	u, _ := http.NewRequest("GET", "https://www.luogu.com.cn/", nil)
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
```

- [ ] **Step 2: 编译验证**

```bash
cd "c:/Users/29084/Desktop/code/go/luoguSDK" && go build ./...
```

- [ ] **Step 3: 提交**

```bash
git add cookiestore.go && git commit -m "feat: add exportable cookie jar with file persistence"
```

---

### Task 5: 重试逻辑

**文件：**
- 创建: `retry.go`

- [ ] **Step 1: 创建 retry.go**

```go
package luogusdk

import (
	"math"
	"net"
	"time"
)

// defaultBackoff 默认指数退避: 1s, 2s, 4s, ...
func defaultBackoff(attempt int) time.Duration {
	return time.Duration(math.Pow(2, float64(attempt))) * time.Second
}

// shouldRetry 判断错误是否应该重试（仅网络错误，不重试业务错误）
func shouldRetry(err error) bool {
	if err == nil {
		return false
	}
	// 网络超时等临时错误可重试
	if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
		return true
	}
	return false
}

// shouldRetryStatus 判断 HTTP 状态码是否应该重试（仅 5xx）
func shouldRetryStatus(statusCode int) bool {
	return statusCode >= 500
}
```

- [ ] **Step 2: 编译验证**

```bash
cd "c:/Users/29084/Desktop/code/go/luoguSDK" && go build ./...
```

- [ ] **Step 3: 提交**

```bash
git add retry.go && git commit -m "feat: add retry logic with exponential backoff"
```

---

### Task 6: Client 核心

**文件：**
- 修改: `client.go`

- [ ] **Step 1: 重写 client.go**

```go
package luogusdk

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

const (
	defaultBaseURL = "https://www.luogu.com.cn"
	defaultUA      = "Go-http-client/2.0"
)

// Client 洛谷 SDK 客户端
type Client struct {
	httpClient *http.Client
	cookieJar  *ExportableCookieJar
	csrfToken  string
	baseURL    string
	cookieFile string
	maxRetries int
	backoffFn  func(int) time.Duration

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
		baseURL:    defaultBaseURL,
		cookieFile: cookiePath,
		maxRetries: 3,
		backoffFn:  defaultBackoff,
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

	// 尝试加载持久化的 cookie
	_ = loadCookies(c.cookieJar, c.cookieFile)

	return c, nil
}

// newRequest 创建带默认请求头的 HTTP 请求
func (c *Client) newRequest(method, path string, body interface{}) (*http.Request, error) {
	url := c.baseURL + path

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

	req.Header.Set("User-Agent", defaultUA)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Referer", defaultBaseURL+"/")

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
		return fmt.Errorf("unmarshal response: %w", err)
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
```

**注意：** 现有的 `client.go` 只有 `package luogusdk` 一行，直接覆盖。

- [ ] **Step 2: 编译验证**

```bash
cd "c:/Users/29084/Desktop/code/go/luoguSDK" && go build ./...
```

- [ ] **Step 3: 提交**

```bash
git add client.go && git commit -m "feat: implement Client core with HTTP helpers, CSRF, cookie persistence"
```

---

### Task 7: CaptchaSolver 与内置 OCR

**文件：**
- 创建: `captcha.go`
- 创建: `internal/ocr/ocr.go`

- [ ] **Step 1: 创建 internal/ocr/ocr.go**

```go
package ocr

import (
	"bytes"
	"fmt"
	"os/exec"
)

// DDddocr 基于 Python ddddocr 的验证码识别器
type DDddocr struct {
	pythonCmd string // python 或 python3
}

// NewDDddocr 创建 ddddocr 识别器
func NewDDddocr() *DDddocr {
	return &DDddocr{pythonCmd: "python"}
}

// NewDDddocrWithPython 指定 Python 解释器路径
func NewDDddocrWithPython(python string) *DDddocr {
	return &DDddocr{pythonCmd: python}
}

// Recognize 识别验证码图片，返回识别结果
func (d *DDddocr) Recognize(image []byte) (string, error) {
	// 调用 Python 脚本: 通过 stdin 传入图片，stdout 输出识别结果
	script := `
import sys
import ddddocr

ocr = ddddocr.DdddOcr()
result = ocr.classification(sys.stdin.buffer.read())
print(result)
`
	cmd := exec.Command(d.pythonCmd, "-c", script)
	cmd.Stdin = bytes.NewReader(image)

	var stdout bytes.Buffer
	cmd.Stdout = &stdout

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("ddddocr run failed: %w, stderr: %s", err, stderr.String())
	}

	// 去除输出中的换行符和空白
	result := bytes.TrimSpace(stdout.Bytes())
	return string(result), nil
}
```

- [ ] **Step 2: 创建 captcha.go**

```go
package luogusdk

import "github.com/PuerkitoBio/laoin114514/luoguSDK/internal/ocr"

// DDDDOCRSolver 返回一个使用 ddddocr 的 CaptchaSolver
func DDDDOCRSolver() CaptchaSolver {
	engine := ocr.NewDDddocr()
	return func(image []byte) (string, error) {
		return engine.Recognize(image)
	}
}

// DDDDOCRSolverWithPython 指定 Python 解释器路径
func DDDDOCRSolverWithPython(python string) CaptchaSolver {
	engine := ocr.NewDDddocrWithPython(python)
	return func(image []byte) (string, error) {
		return engine.Recognize(image)
	}
}
```

**修正:** 上面的导入路径有问题，应该是本模块内部导入。

修正后的 `captcha.go`:

```go
package luogusdk

import "laoin114514/luoguSDK/internal/ocr"

// DDDDOCRSolver 返回一个使用 ddddocr 的 CaptchaSolver
func DDDDOCRSolver() CaptchaSolver {
	engine := ocr.NewDDddocr()
	return func(image []byte) (string, error) {
		return engine.Recognize(image)
	}
}

// DDDDOCRSolverWithPython 指定 Python 解释器路径
func DDDDOCRSolverWithPython(python string) CaptchaSolver {
	engine := ocr.NewDDddocrWithPython(python)
	return func(image []byte) (string, error) {
		return engine.Recognize(image)
	}
}
```

- [ ] **Step 3: 编译验证**

```bash
cd "c:/Users/29084/Desktop/code/go/luoguSDK" && go build ./...
```

- [ ] **Step 4: 提交**

```bash
git add internal/ocr/ocr.go captcha.go && git commit -m "feat: add CaptchaSolver type and built-in ddddocr wrapper"
```

---

### Task 8: AuthService

**文件：**
- 创建: `auth.go`

- [ ] **Step 1: 创建 auth.go**

```go
package luogusdk

// AuthService 认证服务
type AuthService struct {
	client *Client
}

// RefreshCSRF 刷新 CSRF token（需要先调用才能登录）
func (a *AuthService) RefreshCSRF() error {
	return a.client.refreshCSRF()
}

// GetCaptcha 获取验证码图片，返回 JPEG 字节
func (a *AuthService) GetCaptcha() ([]byte, error) {
	resp, err := a.client.get("/lg4/captcha")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return readAllBody(resp)
}

// Login 使用用户名、密码、验证码登录
func (a *AuthService) Login(username, password, captcha string) (*LoginResponse, error) {
	body := &LoginRequest{
		Username: username,
		Password: password,
		Captcha:  captcha,
	}

	resp, err := a.client.post("/do-auth/password", body)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errResp struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
		}
		if parseErr := parseBody(resp, &errResp); parseErr == nil && errResp.Message != "" {
			return nil, &AuthError{Code: errResp.Code, Message: errResp.Message}
		}
		return nil, &AuthError{Code: resp.StatusCode, Message: "login failed"}
	}

	var result LoginResponse
	if err := parseBody(resp, &result); err != nil {
		return nil, err
	}

	// 登录成功后持久化 cookie
	_ = a.client.saveCookiesToFile()

	return &result, nil
}

// LoginWithSolver 使用 CaptchaSolver 自动获取验证码并登录
func (a *AuthService) LoginWithSolver(username, password string, solver CaptchaSolver) (*LoginResponse, error) {
	// 1. 获取验证码图片
	image, err := a.GetCaptcha()
	if err != nil {
		return nil, err
	}

	// 2. 调用 solver 识别验证码
	captcha, err := solver(image)
	if err != nil {
		return nil, &AuthError{Code: 0, Message: "captcha solve failed: " + err.Error()}
	}

	// 3. 登录
	return a.Login(username, password, captcha)
}

// Logout 登出当前会话
func (a *AuthService) Logout() error {
	resp, err := a.client.post("/auth/logout", nil)
	if err != nil {
		return err
	}
	resp.Body.Close()
	return nil
}

// Verify 验证当前登录状态是否有效
func (a *AuthService) Verify() error {
	return a.client.verifyAuth()
}

// SaveCookies 手动持久化当前 cookie
func (a *AuthService) SaveCookies() error {
	return a.client.saveCookiesToFile()
}

// IsAuthenticated 检查是否已经登录（通过 cookie 判断）
func (a *AuthService) IsAuthenticated() bool {
	return a.client.verifyAuth() == nil
}
```

**修正：** 需要使用标准库。需要在文件头添加缺少的导入：

```go
package luogusdk

import "net/http"
```

以及添加辅助函数 `readAllBody`。实际上这个函数应该放在 `client.go` 里，或者直接在 auth.go 中用 `io.ReadAll`。

修正后的 auth.go:

```go
package luogusdk

import (
	"io"
	"net/http"
)

// AuthService 认证服务
type AuthService struct {
	client *Client
}

// RefreshCSRF 刷新 CSRF token（需要先调用才能登录）
func (a *AuthService) RefreshCSRF() error {
	return a.client.refreshCSRF()
}

// GetCaptcha 获取验证码图片，返回 JPEG 字节
func (a *AuthService) GetCaptcha() ([]byte, error) {
	resp, err := a.client.get("/lg4/captcha")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return io.ReadAll(resp.Body)
}

// Login 使用用户名、密码、验证码登录
func (a *AuthService) Login(username, password, captcha string) (*LoginResponse, error) {
	body := &LoginRequest{
		Username: username,
		Password: password,
		Captcha:  captcha,
	}

	resp, err := a.client.post("/do-auth/password", body)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errResp struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
		}
		if parseErr := parseBody(resp, &errResp); parseErr == nil && errResp.Message != "" {
			return nil, &AuthError{Code: errResp.Code, Message: errResp.Message}
		}
		return nil, &AuthError{Code: resp.StatusCode, Message: "login failed"}
	}

	var result LoginResponse
	if err := parseBody(resp, &result); err != nil {
		return nil, err
	}

	_ = a.client.saveCookiesToFile()

	return &result, nil
}

// LoginWithSolver 使用 CaptchaSolver 自动获取验证码并登录
func (a *AuthService) LoginWithSolver(username, password string, solver CaptchaSolver) (*LoginResponse, error) {
	image, err := a.GetCaptcha()
	if err != nil {
		return nil, err
	}

	captcha, err := solver(image)
	if err != nil {
		return nil, &AuthError{Code: 0, Message: "captcha solve failed: " + err.Error()}
	}

	return a.Login(username, password, captcha)
}

// Logout 登出当前会话
func (a *AuthService) Logout() error {
	resp, err := a.client.post("/auth/logout", nil)
	if err != nil {
		return err
	}
	resp.Body.Close()
	return nil
}

// Verify 验证当前登录状态是否有效
func (a *AuthService) Verify() error {
	return a.client.verifyAuth()
}

// IsAuthenticated 检查是否已经登录
func (a *AuthService) IsAuthenticated() bool {
	return a.client.verifyAuth() == nil
}
```

- [ ] **Step 2: 编译验证**

```bash
cd "c:/Users/29084/Desktop/code/go/luoguSDK" && go build ./...
```

- [ ] **Step 3: 提交**

```bash
git add auth.go && git commit -m "feat: implement AuthService (login, logout, captcha, cookie persistence)"
```

---

### Task 9: ProblemService

**文件：**
- 创建: `problem.go`

- [ ] **Step 1: 创建 problem.go**

```go
package luogusdk

import (
	"fmt"
	"net/http"
)

// ProblemService 题目服务
type ProblemService struct {
	client *Client
}

// apiResponse 洛谷通用 API 响应包装
type apiResponse struct {
	Code int             `json:"code"`
	Data json.RawMessage `json:"data"`
}

func (p *ProblemService) checkAuth() error {
	if err := p.client.verifyAuth(); err != nil {
		return err
	}
	return nil
}

// Get 获取题目详情
func (p *ProblemService) Get(pid string) (*Problem, error) {
	path := fmt.Sprintf("/problem/%s", pid)
	resp, err := p.client.get(path)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("get problem %s: status %d", pid, resp.StatusCode)
	}

	var result struct {
		Data struct {
			Problem Problem `json:"problem"`
		} `json:"data"`
	}
	if err := parseBody(resp, &result); err != nil {
		return nil, err
	}
	return &result.Data.Problem, nil
}

// Search 搜索题目
func (p *ProblemService) Search(params SearchParams) (*SearchResult, error) {
	path := fmt.Sprintf("/problem/list?keyword=%s&page=%d&_contentOnly=1",
		params.Keyword, params.Page)
	resp, err := p.client.get(path)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("search problems: status %d", resp.StatusCode)
	}

	var result struct {
		Data struct {
			Problems struct {
				Result []ProblemSummary `json:"result"`
				Total  int              `json:"total"`
				Page   int              `json:"page"`
			} `json:"problems"`
		} `json:"data"`
	}
	if err := parseBody(resp, &result); err != nil {
		return nil, err
	}
	return &SearchResult{
		Problems: result.Data.Problems.Result,
		Total:    result.Data.Problems.Total,
		Page:     result.Data.Problems.Page,
	}, nil
}

// GetSolutions 获取题解列表
func (p *ProblemService) GetSolutions(pid string, page int) (*SolutionList, error) {
	path := fmt.Sprintf("/problem/solution?pid=%s&page=%d&_contentOnly=1", pid, page)
	resp, err := p.client.get(path)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("get solutions for %s: status %d", pid, resp.StatusCode)
	}

	var result struct {
		Data struct {
			Solutions struct {
				Result []SolutionSummary `json:"result"`
				Total  int               `json:"total"`
				Page   int               `json:"page"`
			} `json:"solutions"`
		} `json:"data"`
	}
	if err := parseBody(resp, &result); err != nil {
		return nil, err
	}
	return &SolutionList{
		Solutions: result.Data.Solutions.Result,
		Total:     result.Data.Solutions.Total,
		Page:      result.Data.Solutions.Page,
	}, nil
}

// GetSolutionDetail 获取题解详情
func (p *ProblemService) GetSolutionDetail(sid string) (*Solution, error) {
	path := fmt.Sprintf("/problem/solution/%s?_contentOnly=1", sid)
	resp, err := p.client.get(path)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("get solution %s: status %d", sid, resp.StatusCode)
	}

	var result struct {
		Data struct {
			Solution Solution `json:"solution"`
		} `json:"data"`
	}
	if err := parseBody(resp, &result); err != nil {
		return nil, err
	}
	return &result.Data.Solution, nil
}

// GetTranslation 获取题目翻译
func (p *ProblemService) GetTranslation(pid string) ([]Translation, error) {
	path := fmt.Sprintf("/problem/translation?pid=%s&_contentOnly=1", pid)
	resp, err := p.client.get(path)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("get translation for %s: status %d", pid, resp.StatusCode)
	}

	var result struct {
		Data struct {
			Translations []Translation `json:"translations"`
		} `json:"data"`
	}
	if err := parseBody(resp, &result); err != nil {
		return nil, err
	}
	return result.Data.Translations, nil
}
```

**修正：** 需要添加 `encoding/json` 导入。

修正后的 problem.go:

```go
package luogusdk

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// ProblemService 题目服务
type ProblemService struct {
	client *Client
}

// Get 获取题目详情
func (p *ProblemService) Get(pid string) (*Problem, error) {
	path := fmt.Sprintf("/problem/%s", pid)
	resp, err := p.client.get(path)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("get problem %s: status %d", pid, resp.StatusCode)
	}

	var result struct {
		Data struct {
			Problem Problem `json:"problem"`
		} `json:"data"`
	}
	if err := parseBody(resp, &result); err != nil {
		return nil, err
	}
	return &result.Data.Problem, nil
}

// Search 搜索题目
func (p *ProblemService) Search(params SearchParams) (*SearchResult, error) {
	query := fmt.Sprintf("keyword=%s&page=%d&_contentOnly=1", params.Keyword, params.Page)
	if len(params.Difficulty) == 2 {
		query += fmt.Sprintf("&difficulty=%d,%d", params.Difficulty[0], params.Difficulty[1])
	}
	for _, tag := range params.Tags {
		query += fmt.Sprintf("&tag=%d", tag)
	}
	path := "/problem/list?" + query
	resp, err := p.client.get(path)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("search problems: status %d", resp.StatusCode)
	}

	var result struct {
		Data struct {
			Problems struct {
				Result []ProblemSummary `json:"result"`
				Total  int              `json:"total"`
				Page   int              `json:"page"`
			} `json:"problems"`
		} `json:"data"`
	}
	if err := parseBody(resp, &result); err != nil {
		return nil, err
	}
	return &SearchResult{
		Problems: result.Data.Problems.Result,
		Total:    result.Data.Problems.Total,
		Page:     result.Data.Problems.Page,
	}, nil
}

// GetSolutions 获取题解列表
func (p *ProblemService) GetSolutions(pid string, page int) (*SolutionList, error) {
	path := fmt.Sprintf("/problem/solution?pid=%s&page=%d&_contentOnly=1", pid, page)
	resp, err := p.client.get(path)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("get solutions for %s: status %d", pid, resp.StatusCode)
	}

	var result struct {
		Data struct {
			Solutions struct {
				Result []SolutionSummary `json:"result"`
				Total  int               `json:"total"`
				Page   int               `json:"page"`
			} `json:"solutions"`
		} `json:"data"`
	}
	if err := parseBody(resp, &result); err != nil {
		return nil, err
	}
	return &SolutionList{
		Solutions: result.Data.Solutions.Result,
		Total:     result.Data.Solutions.Total,
		Page:      result.Data.Solutions.Page,
	}, nil
}

// GetSolutionDetail 获取题解详情
func (p *ProblemService) GetSolutionDetail(sid string) (*Solution, error) {
	path := fmt.Sprintf("/problem/solution/%s?_contentOnly=1", sid)
	resp, err := p.client.get(path)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("get solution %s: status %d", sid, resp.StatusCode)
	}

	var result struct {
		Data struct {
			Solution Solution `json:"solution"`
		} `json:"data"`
	}
	if err := parseBody(resp, &result); err != nil {
		return nil, err
	}
	return &result.Data.Solution, nil
}

// GetTranslation 获取题目翻译
func (p *ProblemService) GetTranslation(pid string) ([]Translation, error) {
	path := fmt.Sprintf("/problem/translation?pid=%s&_contentOnly=1", pid)
	resp, err := p.client.get(path)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("get translation for %s: status %d", pid, resp.StatusCode)
	}

	var result struct {
		Data struct {
			Translations []Translation `json:"translations"`
		} `json:"data"`
	}
	if err := parseBody(resp, &result); err != nil {
		return nil, err
	}
	return result.Data.Translations, nil
}
```

- [ ] **Step 2: 编译验证**

```bash
cd "c:/Users/29084/Desktop/code/go/luoguSDK" && go build ./...
```

- [ ] **Step 3: 提交**

```bash
git add problem.go && git commit -m "feat: implement ProblemService (get, search, solutions, translation)"
```

---

### Task 10: 编译与验证

**文件：** 无

- [ ] **Step 1: 完整编译**

```bash
cd "c:/Users/29084/Desktop/code/go/luoguSDK" && go build ./...
```

期望：所有包编译无错误。

- [ ] **Step 2: 运行 vet 检查**

```bash
cd "c:/Users/29084/Desktop/code/go/luoguSDK" && go vet ./...
```

期望：无警告。

- [ ] **Step 3: 验证 API 端点（手动浏览器抓包）**

对于以下端点，需要通过浏览器开发者工具在洛谷网站上验证实际请求格式：

| 端点 | 需要确认 |
|------|---------|
| `POST /do-auth/password` | 请求体字段名是否确为 `username`/`password`/`captcha` |
| `GET /problem/:pid` | 实际 API 路径和响应结构 |
| `GET /problem/list` | 查询参数格式（keyword, difficulty, tag） |
| `GET /problem/solution` | 查询参数格式 |
| `GET /problem/translation` | 查询参数格式 |

**注意：** 洛谷的 SPA 可能使用 `_contentOnly=1` 参数获取纯 JSON 响应。如果实际 API 路径不同，需要调整 `problem.go` 中的路径。

---

### Task 11: 错误类型单元测试

**文件：**
- 创建: `errors_test.go`

- [ ] **Step 1: 创建 errors_test.go**

```go
package luogusdk

import (
	"errors"
	"testing"
)

func TestAuthError(t *testing.T) {
	err := &AuthError{Code: 403, Message: "wrong password"}
	want := "auth error [403]: wrong password"
	if err.Error() != want {
		t.Errorf("expected %q, got %q", want, err.Error())
	}
}

func TestCSRFErrorUnwrap(t *testing.T) {
	inner := errors.New("timeout")
	err := &CSRFError{Err: inner}
	if !errors.Is(err, inner) {
		t.Error("CSRFError should unwrap to inner error")
	}
}

func TestNetworkErrorUnwrap(t *testing.T) {
	inner := errors.New("connection refused")
	err := &NetworkError{Err: inner}
	if !errors.Is(err, inner) {
		t.Error("NetworkError should unwrap to inner error")
	}
}

func TestUnauthorizedError(t *testing.T) {
	err := &UnauthorizedError{}
	if err.Error() == "" {
		t.Error("UnauthorizedError should have a message")
	}
}
```

- [ ] **Step 2: 运行测试**

```bash
cd "c:/Users/29084/Desktop/code/go/luoguSDK" && go test ./... -v
```

期望: 全部 4 个测试通过。

- [ ] **Step 3: 提交**

```bash
git add errors_test.go && git commit -m "test: add error type unit tests"
```

---

### Task 12: Cookie 持久化测试

**文件：**
- 创建: `cookiestore_test.go`

- [ ] **Step 1: 创建 cookiestore_test.go**

```go
package luogusdk

import (
	"net/http"
	"path/filepath"
	"testing"
)

func TestExportableCookieJarRoundTrip(t *testing.T) {
	jar, err := newExportableCookieJar()
	if err != nil {
		t.Fatalf("create jar: %v", err)
	}

	u, _ := http.NewRequest("GET", "https://www.luogu.com.cn/", nil)
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
	u, _ := http.NewRequest("GET", "https://www.luogu.com.cn/", nil)
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
```

- [ ] **Step 2: 运行测试**

```bash
cd "c:/Users/29084/Desktop/code/go/luoguSDK" && go test ./... -v -run "TestExportable|TestSaveLoad"
```

期望: 2 个测试通过。

- [ ] **Step 3: 提交**

```bash
git add cookiestore_test.go && git commit -m "test: add cookie persistence round-trip tests"
```

---

### Task 13: Client 请求构建与重试测试

**文件：**
- 创建: `client_test.go`

- [ ] **Step 1: 创建 client_test.go**

```go
package luogusdk

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestClientNewRequestHeaders(t *testing.T) {
	c := &Client{
		baseURL:    "https://www.luogu.com.cn",
		csrfToken:  "test-csrf-token",
		httpClient: &http.Client{},
	}

	req, err := c.newRequest("POST", "/test", map[string]string{"key": "value"})
	if err != nil {
		t.Fatalf("newRequest: %v", err)
	}

	if ct := req.Header.Get("Content-Type"); ct != "application/json" {
		t.Errorf("Content-Type = %q, want application/json", ct)
	}
	if ref := req.Header.Get("Referer"); ref != "https://www.luogu.com.cn/" {
		t.Errorf("Referer = %q, want https://www.luogu.com.cn/", ref)
	}
	if csrf := req.Header.Get("X-CSRF-TOKEN"); csrf != "test-csrf-token" {
		t.Errorf("X-CSRF-TOKEN = %q, want test-csrf-token", csrf)
	}
	if ua := req.Header.Get("User-Agent"); ua == "" {
		t.Error("User-Agent should not be empty")
	}
}

func TestClientNewRequestNoCSRFForGET(t *testing.T) {
	c := &Client{
		baseURL:    "https://www.luogu.com.cn",
		csrfToken:  "test-csrf-token",
		httpClient: &http.Client{},
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

	c := &Client{
		baseURL:    server.URL,
		maxRetries: 3,
		backoffFn:  func(int) time.Duration { return 0 },
		httpClient: &http.Client{},
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

	c := &Client{
		baseURL:    server.URL,
		maxRetries: 2,
		backoffFn:  func(int) time.Duration { return 0 },
		httpClient: &http.Client{},
	}

	req, _ := http.NewRequest("GET", server.URL+"/test", nil)
	resp, _ := c.do(req)
	resp.Body.Close()

	if callCount != 3 {
		t.Errorf("5xx should retry, expected 3 calls, got %d", callCount)
	}
}
```

- [ ] **Step 2: 运行测试**

```bash
cd "c:/Users/29084/Desktop/code/go/luoguSDK" && go test ./... -v -run "TestClient"
```

期望: 4 个测试通过。

- [ ] **Step 3: 提交**

```bash
git add client_test.go && git commit -m "test: add Client request building and retry tests"
```

---

### Task 14: 最终全量验证

- [ ] **Step 1: 运行所有测试**

```bash
cd "c:/Users/29084/Desktop/code/go/luoguSDK" && go test ./... -v
```

期望: 全部测试通过。

- [ ] **Step 2: 完整编译和检查**

```bash
cd "c:/Users/29084/Desktop/code/go/luoguSDK" && go build ./... && go vet ./...
```

期望: 零错误、零警告。

- [ ] **Step 3: 提交**

```bash
git commit -m "chore: final compilation and vet check passed" --allow-empty
```
```
