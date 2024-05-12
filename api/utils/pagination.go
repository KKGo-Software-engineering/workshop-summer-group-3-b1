package utils

type Pagination struct {
	CurrentPage uint `json:"current_page"`
	TotalPages  uint `json:"total_pages"`
	PerPage     uint `json:"per_page"`
}
