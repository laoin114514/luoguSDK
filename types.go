package luogusdk

// CaptchaSolver 验证码求解器，接收 JPEG 图片字节，返回识别结果
type CaptchaSolver func(image []byte) (string, error)

// LoginRequest 登录请求体
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Captcha  string `json:"captcha"`
}

// LoginResponse 登录响应（会话由 HTTP cookie 维护，此处字段供参考）
type LoginResponse struct {
	UID      int    `json:"uid"`
	ClientID string `json:"client_id"`
}

// Problem 题目详情（匹配洛谷 SSR 页面中 lentille-context 的实际结构）
type Problem struct {
	PID        string         `json:"pid"`
	Title      string         `json:"name"`
	Difficulty int            `json:"difficulty"`
	Tags       []int          `json:"tags"`
	Samples    [][]string     `json:"samples"`
	Limits     ProblemLimits  `json:"limits"`
	Provider   UserInfo       `json:"provider"`
	Content    ProblemContent `json:"contenu"`
}

// DescText 返回题目描述的 Markdown 文本
func (p *Problem) DescText() string { return p.Content.Description }

// InputText 返回输入格式描述
func (p *Problem) InputText() string { return p.Content.InputFormat }

// OutputText 返回输出格式描述
func (p *Problem) OutputText() string { return p.Content.OutputFormat }

// HintText 返回说明/提示
func (p *Problem) HintText() string { return p.Content.Hint }

// TimeLimit 返回时间限制（ms），取第一个值
func (p *Problem) TimeLimit() int {
	if len(p.Limits.Time) > 0 {
		return p.Limits.Time[0]
	}
	return 0
}

// MemoryLimit 返回内存限制（KB），取第一个值
func (p *Problem) MemoryLimit() int {
	if len(p.Limits.Memory) > 0 {
		return p.Limits.Memory[0]
	}
	return 0
}

// ProblemContent 题目内容（从 contenu/content 字段提取）
type ProblemContent struct {
	Description string `json:"description"`
	InputFormat string `json:"formatI"`
	OutputFormat string `json:"formatO"`
	Hint        string `json:"hint"`
	Background  string `json:"background"`
}

// ProblemLimits 时空限制
type ProblemLimits struct {
	Time   []int `json:"time"`
	Memory []int `json:"memory"`
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
	Title      string `json:"name"`
	Difficulty int    `json:"difficulty"`
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
	Language     string `json:"-"`
	Title        string `json:"name"`
	Description  string `json:"description"`
	InputFormat  string `json:"formatI"`
	OutputFormat string `json:"formatO"`
	Hint         string `json:"hint"`
	Background   string `json:"background"`
}

// UserInfo 用户信息
type UserInfo struct {
	UID    int    `json:"uid"`
	Name   string `json:"name"`
	Avatar string `json:"avatar"`
}
