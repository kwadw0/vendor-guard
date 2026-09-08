package utils

import (
	"net/url"
	"strconv"
)

const (
	DefaultPage  = 1
	DefaultLimit = 20
	MaxLimit     = 50
)

// PaginationQuery documents standard ?page=&limit= params for Swagger.
// Handlers declare @Param page/limit explicitly; this struct is for parsing.
type PaginationQuery struct {
	Page  int `json:"page" validate:"omitempty,min=1"`
	Limit int `json:"limit" validate:"omitempty,min=1,max=50"`
}

// ParsePagination reads ?page= (1-based) & ?limit= with defaults + clamp.
// defaultLimit <=0 uses DefaultLimit, maxLimit <=0 uses MaxLimit.
func ParsePagination(q url.Values, defaultLimit, maxLimit int) (page, limit, offset int) {
	if defaultLimit <= 0 {
		defaultLimit = DefaultLimit
	}
	if maxLimit <= 0 {
		maxLimit = MaxLimit
	}
	page, _ = strconv.Atoi(q.Get("page"))
	if page < 1 {
		page = DefaultPage
	}
	limit, _ = strconv.Atoi(q.Get("limit"))
	if limit < 1 {
		limit = defaultLimit
	}
	if limit > maxLimit {
		limit = maxLimit
	}
	offset = (page - 1) * limit
	return page, limit, offset
}

// NewPaginationMeta builds meta envelope with total_pages derived.
func NewPaginationMeta(page, limit, total int) PaginationMeta {
	totalPages := 0
	if total > 0 && limit > 0 {
		totalPages = (total + limit - 1) / limit
	}
	return PaginationMeta{
		Page:       page,
		Limit:      limit,
		Total:      total,
		TotalPages: totalPages,
	}
}

// PaginateSlice slices in-memory items for small lists (forms/partners/orgs/templates).
// For large tables prefer DB-level LIMIT/OFFSET; this keeps API shape consistent.
func PaginateSlice[T any](items []T, page, limit int) ([]T, PaginationMeta) {
	total := len(items)
	if total == 0 {
		return []T{}, NewPaginationMeta(page, limit, 0)
	}
	start := (page - 1) * limit
	if start >= total {
		return []T{}, NewPaginationMeta(page, limit, total)
	}
	end := start + limit
	if end > total {
		end = total
	}
	return items[start:end], NewPaginationMeta(page, limit, total)
}
