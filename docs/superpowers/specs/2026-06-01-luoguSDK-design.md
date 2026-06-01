# LuoguSDK Design Document

## Overview

A Go SDK for the Luogu (洛谷) platform API. Phase 1 covers authentication (login, captcha, cookie persistence) and read-only problem operations (fetch, search, solutions, translations). No submission/code-execution support in this phase.

## Architecture: Client + Service Layers

```
Client (HTTP session, CSRF, cookie management)
  ├── AuthService   (login, logout, captcha, cookie persistence)
  └── ProblemService (fetch, search, solutions, translations)
```

**Client** owns the `http.Client` (with cookie jar), the CSRF token, and the session cookies (`_uid`, `__client_id`). Services are lightweight wrappers that hold a pointer back to Client and make API calls through it.

All non-GET requests automatically inject headers: `X-CSRF-TOKEN`, `Referer: https://www.luogu.com.cn/`, `Content-Type: application/json`. User-Agent defaults to Go's standard UA (avoids Luogu anti-bot rules — must not contain `python-requests` or start with `mozilla/`).

## Package Structure

```
luoguSDK/
├── client.go       # Client struct, NewClient, configuration, HTTP helpers
├── auth.go         # AuthService: Login, Logout, Lock, Unlock, RefreshCSRF, GetCaptcha, VerifyAuth
├── problem.go      # ProblemService: Get, Search, GetSolutions, GetSolutionDetail, GetTranslation
├── captcha.go      # CaptchaSolver type, CaptchaSolverFunc, internal OCR adapter
├── errors.go       # Error types: AuthError, CSRFError, NetworkError, UnauthorizedError
├── types.go        # Shared types: LoginRequest, LoginResponse, Problem, Solution, etc.
├── cookiestore.go  # Cookie persist/load to file, TryLoadCookies, SaveCookies
├── retry.go        # Retry logic, backoff strategies
├── internal/
│   └── ocr/        # Built-in OCR tool (see Captcha Design section)
└── go.mod
```

## Authentication Flow

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

## API Endpoints

| Method | Path | Purpose |
|--------|------|---------|
| GET | `/` | Extract CSRF token from HTML meta tag |
| GET | `/lg4/captcha` | Get captcha image (JPEG) |
| POST | `/do-auth/password` | Login with username/password/captcha |
| POST | `/auth/logout` | Logout |
| GET | `/problem/:pid` | Get problem detail (JSON API) |
| GET | `/problem/list` | Search problems |
| GET | `/problem/solution` | List solutions for a problem |
| GET | `/problem/solution/:sid` | Get solution detail |
| GET | `/problem/translation` | Get problem translation |

**Note:** Exact query parameters for problem endpoints will be determined during implementation by inspecting the Luogu web app's network requests. The API docs (`0f-0b/luogu-api-docs`) provide type references but not all query parameter details.

## Core Types

```go
// --- Captcha ---
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

// --- Auth ---
type LoginRequest struct {
    Username string `json:"username"`
    Password string `json:"password"`
    Captcha  string `json:"captcha"`
}

type LoginResponse struct {
    UID      int    `json:"uid"`
    ClientID string `json:"client_id"`
    // Additional fields from Luogu response
}

// --- Problem ---
type Problem struct {
    PID         string        `json:"pid"`
    Title       string        `json:"title"`
    Difficulty  int           `json:"difficulty"`
    Background  string        `json:"background"`
    Description string        `json:"description"`
    InputFormat string        `json:"inputFormat"`
    OutputFormat string       `json:"outputFormat"`
    Samples     []Sample      `json:"samples"`
    Hints       []string      `json:"hints"`
    Tags        []Tag         `json:"tags"`
    TimeLimit   int           `json:"timeLimit"`
    MemoryLimit int           `json:"memoryLimit"`
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
    Keyword    string
    Difficulty []int
    Tags       []int
    Page       int
    PageSize   int
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
    ID       int
    Author   UserInfo
    Title    string
    Content  string
    Likes    int
}

type SolutionList struct {
    Solutions []SolutionSummary
    Total     int
    Page      int
}

type SolutionSummary struct {
    ID       int
    Title    string
    Author   UserInfo
    Likes    int
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

## Cookie Persistence

- Default path: `~/.luogu/cookies.json`
- A custom `ExportableCookieJar` wraps `net/http/cookiejar.Jar` and adds `Export()` / `Import()` methods to serialize/deserialize all stored cookies to/from JSON
- Cookies are saved keyed by domain, containing `_uid` and `__client_id` values
- `TryLoadCookies()` at client init; `SaveCookies()` after successful login
- Configurable via `WithCookieFile(path string)`

## Retry Mechanism

- Default: 3 retries with exponential backoff (1s → 2s → 4s)
- Only retries on network errors and 5xx responses, never on 4xx
- Configurable: `WithRetry(maxRetries int, backoffFn func(int) time.Duration)`

## CaptchaSolver Design

- **Type:** `type CaptchaSolver func(image []byte) (string, error)` — simple function type, consumers implement their own
- **Built-in OCR:** SDK provides a `NewDDDDOCR()` constructor that wraps an external `ddddocr` process (Python), accepting the image bytes and returning the recognized text
  - Alternative: pure Go OCR using a lightweight library (TBD during implementation based on accuracy testing)
  - The built-in solver lives in `internal/ocr/` and is exposed via a public constructor in `captcha.go`
  - Users can also plug in their own solver (e.g., cloud OCR API, manual input)

## Error Handling

```go
type AuthError struct { Code, Message string }     // 登录失败
type CSRFError struct{ Err error }                  // CSRF 获取/过期
type NetworkError struct{ Err error }               // 网络超时等
type UnauthorizedError struct{}                     // 未登录调用需认证的 API
```

All errors implement the `error` interface and wrap the underlying cause (via `errors.Unwrap`).

## Testing Strategy

- Unit tests for `Client` request building (headers, CSRF injection) with `net/http/httptest`
- Integration tests against Luogu API with a test account
- OCR tests with sample captcha images
- Cookie persistence round-trip tests

## Dependencies

- `github.com/PuerkitoBio/goquery` — HTML parsing (CSRF token extraction)
- `github.com/google/go-cmp` — test comparisons (dev only)
- Standard library: `net/http`, `net/http/cookiejar`, `encoding/json`, `time`, `errors`
