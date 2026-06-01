package luogusdk

import (
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
