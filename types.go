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

// Problem 题目详情（字段名匹配洛谷 API 实际响应）
type Problem struct {
	PID          string          `json:"pid"`
	Title        string          `json:"title"`
	Difficulty   *int            `json:"difficulty"`
	Tags         []int           `json:"tags"`
	Description  ProblemDesc     `json:"description"`
	InputFormat  ProblemIOFormat `json:"inputFormat"`
	OutputFormat ProblemIOFormat `json:"outputFormat"`
	Samples      []ProblemSample `json:"sampleTestcases"`
	Limits       *ProblemLimits  `json:"limits"`
}

// DescText 返回题目描述的纯文本
func (p *Problem) DescText() string {
	return p.Description.Text
}

// InputText 返回输入格式描述
func (p *Problem) InputText() string {
	return p.InputFormat.Description
}

// OutputText 返回输出格式描述
func (p *Problem) OutputText() string {
	return p.OutputFormat.Description
}

// TimeLimit 返回时间限制（毫秒），若 API 未提供则返回 0
func (p *Problem) TimeLimit() int {
	if p.Limits == nil {
		return 0
	}
	return p.Limits.Time
}

// MemoryLimit 返回内存限制（KB），若 API 未提供则返回 0
func (p *Problem) MemoryLimit() int {
	if p.Limits == nil {
		return 0
	}
	return p.Limits.Memory
}

// ProblemDesc 题目描述
type ProblemDesc struct {
	Text         string   `json:"text"`
	Notes        []string `json:"notes"`
	ClosingQuote string   `json:"closingQuote"`
}

// ProblemIOFormat 输入/输出格式
type ProblemIOFormat struct {
	Description string `json:"description"`
}

// ProblemSample 输入输出样例
type ProblemSample struct {
	Input  string `json:"input"`
	Output string `json:"output"`
}

// ProblemLimits 时空限制
type ProblemLimits struct {
	Time   int `json:"time"`
	Memory int `json:"memory"`
}

// SearchParams 题目搜索参数
type SearchParams struct {
	Keyword  string
	Page     int
	PageSize int
}

// SearchResult 搜索结果
type SearchResult struct {
	Problems []ProblemSummary
	Total    int
	Page     int
	PerPage  int
}

// ProblemSummary 题目摘要
type ProblemSummary struct {
	PID        string `json:"pid"`
	Title      string `json:"title"`
	Difficulty *int   `json:"difficulty"`
	Tags       []int  `json:"tags"`
}

// Solution 题解详情
type Solution struct {
	ID      int      `json:"id"`
	Author  UserInfo `json:"author"`
	Title   string   `json:"title"`
	Content string   `json:"content"`
	Likes   int      `json:"likes"`
}

// SolutionList 题解列表
type SolutionList struct {
	Solutions []SolutionSummary
	Total     int
	Page      int
	PerPage   int
}

// SolutionSummary 题解摘要
type SolutionSummary struct {
	ID     int      `json:"id"`
	Title  string   `json:"title"`
	Author UserInfo `json:"author"`
	Likes  int      `json:"likes"`
}

// Translation 题目翻译
type Translation struct {
	Language    string `json:"language"`
	Title       string `json:"title"`
	Description string `json:"description"`
}

// UserInfo 用户信息
type UserInfo struct {
	UID    int    `json:"uid"`
	Name   string `json:"name"`
	Avatar string `json:"avatar"`
}
