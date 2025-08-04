package filters

import "math"

type FilterResponse[T any] struct {
	Items      []T   `json:"items" yaml:"items"`
	Page       int   `json:"page" yaml:"page"`
	PageSize   int   `json:"page_size" yaml:"page_size"`
	TotalPages int   `json:"total_pages" yaml:"total_pages"`
	Total      int64 `json:"total" yaml:"total"`
}

type Pagination struct {
	Page     int   `json:"page" yaml:"page"`
	PageSize int   `json:"page_size" yaml:"page_size"`
	Total    int64 `json:"total" yaml:"total"`
}

func (p *Pagination) GetTotalPages() int {
	if p.Total == 0 {
		return 0
	}
	return int(math.Ceil(float64(p.Total) / float64(p.PageSize)))
}

func (p *Pagination) GetOffset() int {
	if p.Total == 0 {
		return 0
	}
	offset := p.GetPageIndex() * p.PageSize
	if offset > int(p.Total) {
		p.Page = 1
		offset = int(p.Total) - p.PageSize
	}

	return offset
}

func (p *Pagination) GetPageIndex() int {
	return p.Page - 1
}

type PaginationResponse[T any] struct {
	Items      []T   `json:"items" yaml:"items"`
	Page       int   `json:"page" yaml:"page"`
	PageSize   int   `json:"page_size" yaml:"page_size"`
	TotalPages int   `json:"total_pages" yaml:"total_pages"`
	Total      int64 `json:"total" yaml:"total"`
}
