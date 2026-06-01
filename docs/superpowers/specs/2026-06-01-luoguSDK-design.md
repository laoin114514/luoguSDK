# LuoguSDK 设计文档

## 概述

洛谷 (luogu.com.cn) 平台的 Go SDK。第一阶段覆盖认证（登录、验证码、Cookie 持久化）和题目只读操作（获取、搜索、题解、翻译）。本阶段不涉及提交/代码执行。

## 架构：Client + 服务分层

```
Client (HTTP 会话、CSRF、Cookie 管理)
  ├── AuthService   (登录、登出、验证码、Cookie 持久化)
  └── ProblemService (题目获取、搜索、题解、翻译)
```

**Client** 持有 `http.Client`（含 CookieJar）、CSRF Token、会话 cookie（`_uid`、`__client_id`）。每个 Service 是轻量封装，持有 Client 指针，通过 Client 发起 HTTP 请求。

所有非 GET 请求自动注入请求头：`X-CSRF-TOKEN`、`Referer: https://www.luogu.com.cn/`、`Content-Type: application/json`。User-Agent 默认使用 Go 标准 UA（规避洛谷反爬规则——不能包含 `python-requests`，不能以 `mozilla/` 开头）。

## 包结构

```
luoguSDK/
├── client.go       # Client 结构体、NewClient、配置项、HTTP 工具方法
├── auth.go         # AuthService: Login、Logout、Lock、Unlock、RefreshCSRF、GetCaptcha、VerifyAuth
├── problem.go      # ProblemService: Get、Search、GetSolutions、GetSolutionDetail、GetTranslation
├── captcha.go      # CaptchaSolver 类型、内置 OCR 构造器
├── errors.go       # 错误类型: AuthError、CSRFError、NetworkError、UnauthorizedError
├── types.go        # 公共类型: LoginRequest、LoginResponse、Problem、Solution 等
├── cookiestore.go  # Cookie 持久化: TryLoadCookies、SaveCookies
├── retry.go        # 重试逻辑、退避策略
├── internal/
│   └── ocr/        # 内置 OCR 工具 (ddddocr 封装)
└── go.mod
```

## 认证流程

```
NewClient(opts...)
    │
    ├── TryLoadCookies() → 从文件反序列化 cookie
    │       │
    │       ├── 加载成功 → VerifyAuth() 轻量校验
    │       │       ├── 200 → 已认证，跳过登录
    │       │       └── 401 → cookie 过期，走登录
    │       └── 加载失败 → 走登录
    │
    └── 登录流程:
            RefreshCSRF()     # GET / → 解析 meta[name=csrf-token]
            GetCaptcha()      # GET /lg4/captcha → []byte (JPEG)
            solver(image)     # CaptchaSolver → string
            Login(user, pwd, captcha)  # POST /do-auth/password
            SaveCookies()     # 持久化到文件
```

## API 端点

| 方法 | 路径 | 用途 |
|--------|------|---------|
| GET | `/` | 从 HTML meta 标签提取 CSRF token |
| GET | `/lg4/captcha` | 获取验证码图片 (JPEG) |
| POST | `/do-auth/password` | 用户名/密码/验证码登录 |
| POST | `/auth/logout` | 登出 |
| GET | `/problem/:pid` | 获取题目详情 (JSON API) |
| GET | `/problem/list` | 搜索题目列表 |
| GET | `/problem/solution` | 获取某题的题解列表 |
| GET | `/problem/solution/:sid` | 获取题解详情 |
| GET | `/problem/translation` | 获取题目翻译 |

**注：** 题目相关端点的查询参数将在实现阶段通过抓包洛谷 Web 应用的实际请求来确定。第三方 API 文档 (`0f-0b/luogu-api-docs`) 提供了类型引用，但未列出全部查询参数细节。

## 核心类型

```go
// --- 验证码 ---
type CaptchaSolver func(image []byte) (string, error)

// --- Client ---
type Client struct {
    httpClient  *http.Client
    csrfToken   string
    baseURL     string
    cookieFile  string
    maxRetries  int
    backoffFn   func(int) time.Duration

    Auth    *AuthService
    Problem *ProblemService
}

type ClientOption func(*Client)

// --- 认证 ---
type LoginRequest struct {
    Username string `json:"username"`
    Password string `json:"password"`
    Captcha  string `json:"captcha"`
}

type LoginResponse struct {
    UID      int    `json:"uid"`
    ClientID string `json:"client_id"`
    // 洛谷响应中的其他字段
}

// --- 题目 ---
type Problem struct {
    PID          string        `json:"pid"`
    Title        string        `json:"title"`
    Difficulty   int           `json:"difficulty"`
    Background   string        `json:"background"`
    Description  string        `json:"description"`
    InputFormat  string        `json:"inputFormat"`
    OutputFormat string        `json:"outputFormat"`
    Samples      []Sample      `json:"samples"`
    Hints        []string      `json:"hints"`
    Tags         []Tag         `json:"tags"`
    TimeLimit    int           `json:"timeLimit"`
    MemoryLimit  int           `json:"memoryLimit"`
}

type Sample struct {
    Input  string `json:"input"`
    Output string `json:"output"`
}

type Tag struct {
    ID   int    `json:"id"`
    Name string `json:"name"`
}

type SearchParams struct {
    Keyword    string  // 搜索关键词
    Difficulty []int   // 难度范围 [min, max]
    Tags       []int   // 标签 ID 列表
    Page       int     // 页码
    PageSize   int     // 每页条数 (默认 20)
}

type SearchResult struct {
    Problems []ProblemSummary
    Total    int
    Page     int
}

type ProblemSummary struct {
    PID        string
    Title      string
    Difficulty int
    Tags       []Tag
}

type Solution struct {
    ID      int
    Author  UserInfo
    Title   string
    Content string // Markdown
    Likes   int
}

type SolutionList struct {
    Solutions []SolutionSummary
    Total     int
    Page      int
}

type SolutionSummary struct {
    ID     int
    Title  string
    Author UserInfo
    Likes  int
}

type Translation struct {
    Language    string
    Title       string
    Description string
}

type UserInfo struct {
    UID    int    `json:"uid"`
    Name   string `json:"name"`
    Avatar string `json:"avatar"`
}
```

## Cookie 持久化

- 默认路径：`~/.luogu/cookies.json`
- 自定义 `ExportableCookieJar` 包装 `net/http/cookiejar.Jar`，新增 `Export()` / `Import()` 方法，将所有 cookie 序列化/反序列化为 JSON
- Cookie 按域名保存，主要包含 `_uid` 和 `__client_id`
- `TryLoadCookies()` 在 Client 初始化时调用；`SaveCookies()` 在登录成功后调用
- 可通过 `WithCookieFile(path string)` 自定义路径

## 重试机制

- 默认：3 次重试，指数退避 (1s → 2s → 4s)
- 仅对网络错误和 5xx 响应重试，4xx 不重试（业务错误重试无意义）
- 可配置：`WithRetry(maxRetries int, backoffFn func(int) time.Duration)`

## CaptchaSolver 设计

- **类型：** `type CaptchaSolver func(image []byte) (string, error)` — 函数类型，由调用方实现
- **内置 OCR：** SDK 提供 `NewDDDDOCR()` 构造器，封装 Python `ddddocr` 进程，接收图片字节并返回识别文本
  - 备选方案：实现阶段测试后，如纯 Go OCR 库准确率满足要求则替换
  - 内置 solver 代码位于 `internal/ocr/`，通过 `captcha.go` 中的公开构造器暴露
  - 用户也可自行接入其他 solver（如云端 OCR API、手动输入）

## 错误处理

```go
type AuthError struct{ Code, Message string }     // 登录失败
type CSRFError struct{ Err error }                 // CSRF 获取/过期
type NetworkError struct{ Err error }              // 网络超时等
type UnauthorizedError struct{}                    // 未登录调用需认证的 API
```

所有错误类型实现 `error` 接口，并通过 `errors.Unwrap` 暴露底层错误。

## 测试策略

- 使用 `net/http/httptest` 对 Client 请求构建（请求头注入、CSRF 注入）进行单元测试
- 使用测试账号对洛谷 API 进行集成测试
- 使用示例验证码图片进行 OCR 测试
- Cookie 持久化往返测试（保存 → 加载 → 验证）

## 依赖

- `github.com/PuerkitoBio/goquery` — HTML 解析（提取 CSRF token）
- `github.com/google/go-cmp` — 测试对比（仅开发环境）
- 标准库：`net/http`、`net/http/cookiejar`、`encoding/json`、`time`、`errors`
