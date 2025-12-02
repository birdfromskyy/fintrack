package utils

import (
	"strconv"

	"github.com/gin-gonic/gin"
)

type Pagination struct {
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
	Total  int `json:"total"`
	Page   int `json:"page"`
	Pages  int `json:"pages"`
}

func GetPaginationParams(c *gin.Context) *Pagination {
	limit := 20
	page := 1

	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 100 {
			limit = parsed
		}
	}

	if p := c.Query("page"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil && parsed > 0 {
			page = parsed
		}
	}

	offset := (page - 1) * limit

	return &Pagination{
		Limit:  limit,
		Offset: offset,
		Page:   page,
	}
}

func (p *Pagination) Calculate(total int) {
	p.Total = total
	if p.Limit > 0 {
		p.Pages = (total + p.Limit - 1) / p.Limit
	}
}
