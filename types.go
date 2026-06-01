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
