package utils

import (
	"net/http"
	"strconv"
)

type Pagination struct {
	Limit  int
	Offset int
}

func NewPagination(limit, offset int) Pagination {
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

	if offset < 0 {
		offset = 0
	}

	return Pagination{
		Limit:  limit,
		Offset: offset,
	}
}

func ParsePagination(r *http.Request) Pagination {
	q := r.URL.Query()

	limit, _ := strconv.Atoi(q.Get("limit"))
	offset, _ := strconv.Atoi(q.Get("offset"))

	return NewPagination(limit, offset)
}

type PageMeta struct {
	Limit      int  `json:"limit"`
	Offset     int  `json:"offset"`
	NextOffset int  `json:"next_offset"`
	PrevOffset *int `json:"prev_offset,omitempty"`
}

func BuildMeta(p Pagination, hasMore bool) PageMeta {
	next := p.Offset + p.Limit

	var prev *int
	if p.Offset > 0 {
		pv := p.Offset - p.Limit
		if pv < 0 {
			pv = 0
		}
		prev = &pv
	}

	if !hasMore {
		next = -1
	}

	return PageMeta{
		Limit:      p.Limit,
		Offset:     p.Offset,
		NextOffset: next,
		PrevOffset: prev,
	}
}

type PaginatedResponse[T any] struct {
	Data []T      `json:"data"`
	Meta PageMeta `json:"meta"`
}
