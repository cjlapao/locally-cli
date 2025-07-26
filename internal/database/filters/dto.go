package filters

type FilterResponse[T any] struct {
	Items      []T   `json:"items" yaml:"items"`
	Page       int   `json:"page" yaml:"page"`
	PageSize   int   `json:"page_size" yaml:"page_size"`
	TotalPages int   `json:"total_pages" yaml:"total_pages"`
	Total      int64 `json:"total" yaml:"total"`
}
