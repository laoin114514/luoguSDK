package luogusdk

import (
	"fmt"
	"net/http"
	"net/url"
)

// ProblemService 题目服务
type ProblemService struct {
	client *Client
}

// Get 获取题目详情
func (p *ProblemService) Get(pid string) (*Problem, error) {
	path := fmt.Sprintf("/problem/%s?_contentOnly=1", pid)
	resp, err := p.client.get(path)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("get problem %s: status %d", pid, resp.StatusCode)
	}

	var result struct {
		Problem Problem `json:"problem"`
	}
	if err := parseBody(resp, &result); err != nil {
		return nil, err
	}
	return &result.Problem, nil
}

// Search 搜索题目
func (p *ProblemService) Search(params SearchParams) (*SearchResult, error) {
	q := url.Values{}
	q.Set("_contentOnly", "1")
	if params.Keyword != "" {
		q.Set("keyword", params.Keyword)
	}
	if params.Page > 0 {
		q.Set("page", fmt.Sprintf("%d", params.Page))
	}
	path := "/problem/list?" + q.Encode()
	resp, err := p.client.get(path)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("search problems: status %d", resp.StatusCode)
	}

	var result struct {
		CurrentData struct {
			Problems struct {
				Result    []ProblemSummary `json:"problems"`
				Count     int              `json:"count"`
				TotalPage int              `json:"totalPages"`
				Page      int              `json:"page"`
				PerPage   int              `json:"perPage"`
			} `json:"problems"`
		} `json:"currentData"`
	}
	if err := parseBody(resp, &result); err != nil {
		return nil, err
	}
	return &SearchResult{
		Problems: result.CurrentData.Problems.Result,
		Total:    result.CurrentData.Problems.Count,
		Page:     result.CurrentData.Problems.Page,
		PerPage:  result.CurrentData.Problems.PerPage,
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
		CurrentData struct {
			Solutions struct {
				Result    []SolutionSummary `json:"solutions"`
				Count     int               `json:"count"`
				TotalPage int               `json:"totalPages"`
				Page      int               `json:"page"`
				PerPage   int               `json:"perPage"`
			} `json:"solutions"`
		} `json:"currentData"`
	}
	if err := parseBody(resp, &result); err != nil {
		return nil, err
	}
	return &SolutionList{
		Solutions: result.CurrentData.Solutions.Result,
		Total:     result.CurrentData.Solutions.Count,
		Page:      result.CurrentData.Solutions.Page,
		PerPage:   result.CurrentData.Solutions.PerPage,
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
		CurrentData struct {
			Solution Solution `json:"solution"`
		} `json:"currentData"`
	}
	if err := parseBody(resp, &result); err != nil {
		return nil, err
	}
	return &result.CurrentData.Solution, nil
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
		CurrentData struct {
			Translations []Translation `json:"translations"`
		} `json:"currentData"`
	}
	if err := parseBody(resp, &result); err != nil {
		return nil, err
	}
	return result.CurrentData.Translations, nil
}
