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
	if err := parseLentilleContext(resp, &result); err != nil {
		return nil, err
	}
	return &result.Data.Problem, nil
}

// Search 搜索题目
func (p *ProblemService) Search(params SearchParams) (*SearchResult, error) {
	path := fmt.Sprintf("/problem/list?keyword=%s&page=%d", params.Keyword, params.Page)
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
				Result    []ProblemSummary `json:"result"`
				Count     int              `json:"count"`
				TotalPage int              `json:"totalPages"`
				Page      int              `json:"page"`
				PerPage   int              `json:"perPage"`
			} `json:"problems"`
		} `json:"data"`
	}
	if err := parseLentilleContext(resp, &result); err != nil {
		return nil, err
	}
	return &SearchResult{
		Problems: result.Data.Problems.Result,
		Total:    result.Data.Problems.Count,
		Page:     result.Data.Problems.Page,
		PerPage:  result.Data.Problems.PerPage,
	}, nil
}

// GetSolutions 获取题解列表
func (p *ProblemService) GetSolutions(pid string, page int) (*SolutionList, error) {
	path := fmt.Sprintf("/problem/solution?pid=%s&page=%d", pid, page)
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
				Result    []SolutionSummary `json:"result"`
				Count     int               `json:"count"`
				TotalPage int               `json:"totalPages"`
				Page      int               `json:"page"`
				PerPage   int               `json:"perPage"`
			} `json:"solutions"`
		} `json:"data"`
	}
	if err := parseLentilleContext(resp, &result); err != nil {
		return nil, err
	}
	return &SolutionList{
		Solutions: result.Data.Solutions.Result,
		Total:     result.Data.Solutions.Count,
		Page:      result.Data.Solutions.Page,
		PerPage:   result.Data.Solutions.PerPage,
	}, nil
}

// GetSolutionDetail 获取题解详情
func (p *ProblemService) GetSolutionDetail(sid string) (*Solution, error) {
	path := fmt.Sprintf("/problem/solution/%s", sid)
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
	if err := parseLentilleContext(resp, &result); err != nil {
		return nil, err
	}
	return &result.Data.Solution, nil
}

// GetTranslation 获取题目翻译（翻译数据嵌在题目详情页的 lentille-context 中）
func (p *ProblemService) GetTranslation(pid string) ([]Translation, error) {
	path := fmt.Sprintf("/problem/%s", pid)
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
			Translations map[string]Translation `json:"translations"`
		} `json:"data"`
	}
	if err := parseLentilleContext(resp, &result); err != nil {
		return nil, err
	}

	var out []Translation
	for lang, t := range result.Data.Translations {
		t.Language = lang
		out = append(out, t)
	}
	return out, nil
}
