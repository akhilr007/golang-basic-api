package utils

import (
	"net/http"
	"strconv"
)

type Pagination struct {
	Limit  int
	Page   int
	Offset int
}

func NewPagination(page, limit int) Pagination {
	const (
		defaultLimit = 10
		maxLimit     = 100
	)

	if limit <= 0 {
		limit = defaultLimit
	}
	if limit > maxLimit {
		limit = maxLimit
	}

	if page <= 0 {
		page = 1
	}

	offset := (page - 1) * limit

	return Pagination{
		Limit:  limit,
		Page:   page,
		Offset: offset,
	}
}

func ParsePagination(r *http.Request) Pagination {
	q := r.URL.Query()

	page, _ := strconv.Atoi(q.Get("page"))
	limit, _ := strconv.Atoi(q.Get("limit"))

	return NewPagination(page, limit)
}

type PageMeta struct {
	Page  int `json:"page"`
	Limit int `json:"limit"`

	NextPage *int `json:"next_page,omitempty"`
	PrevPage *int `json:"prev_page,omitempty"`
}

func BuildMeta(p Pagination, hasMore bool) PageMeta {
	var next *int
	if hasMore {
		n := p.Page + 1
		next = &n
	}

	var prev *int
	if p.Page > 1 {
		pv := p.Page - 1
		prev = &pv
	}

	return PageMeta{
		Page:     p.Page,
		Limit:    p.Limit,
		NextPage: next,
		PrevPage: prev,
	}
}

type PaginatedResponse[T any] struct {
	Data []T      `json:"data"`
	Meta PageMeta `json:"meta"`
}
