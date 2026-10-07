package api

import (
	"net/http"
	"strconv"
)

const (
	defaultLimit = 20
	maxLimit     = 100
)

// PaginationParams holds sanitized pagination values.
type PaginationParams struct {
	Limit  int
	Offset int
}

// PaginationMeta provides pagination information in API responses.
type PaginationMeta struct {
	Total  int64 `json:"total"`
	Limit  int   `json:"limit"`
	Offset int   `json:"offset"`
}

// parsePagination extracts and clamps limit and offset query parameters.
func parsePagination(r *http.Request) PaginationParams {
	limit := defaultLimit
	offset := 0

	q := r.URL.Query()
	if lStr := q.Get("limit"); lStr != "" {
		if val, err := strconv.Atoi(lStr); err == nil {
			if val > 0 {
				limit = val
			}
			if limit > maxLimit {
				limit = maxLimit
			}
		}
	}

	if oStr := q.Get("offset"); oStr != "" {
		if val, err := strconv.Atoi(oStr); err == nil && val >= 0 {
			offset = val
		}
	}

	return PaginationParams{
		Limit:  limit,
		Offset: offset,
	}
}
