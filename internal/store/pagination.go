package store

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type PaginatedFeedQuery struct {
	Limit  int      `json:"limit" validate:"gte=1,lte=20"`
	Offset int      `json:"offset" validate:"gte=0"`
	Sort   string   `json:"sort" validate:"oneof=asc desc"`
	Tags   []string `json:"tags" validate:"max=5"`
	Search string   `json:"search" validate:"max=100"`
	Since  string   `json:"since"`
	Until  string   `json:"until"`
}

func (fq PaginatedFeedQuery) Parse(r *http.Request) (PaginatedFeedQuery, error) {
	qs := r.URL.Query()

	limit := qs.Get("limit")
	if limit != "" {
		l, err := strconv.Atoi(limit)
		if err != nil {
			return fq, err
		}
		fq.Limit = l
	}
	offset := qs.Get("offset")
	if offset != "" {
		o, err := strconv.Atoi(offset)
		if err != nil {
			return fq, err
		}
		fq.Offset = o
	}
	sort := qs.Get("sort")
	if sort != "" {
		fq.Sort = sort
	}

	search := qs.Get("search")
	if search != "" {
		fq.Search = search
	}

	since := qs.Get("since")
	if since != "" {
		s, err := parseTime(since)
		if err != nil {
			return fq, err
		}
		fq.Since = s
	}
	until := qs.Get("until")
	if until != "" {
		u, err := parseTime(until)
		if err != nil {
			return fq, err
		}
		fq.Until = u
	}
	tags := qs.Get("tags")
	if tags != "" {
		fq.Tags = strings.Split(tags, ",")
	}

	return fq, nil
}

func parseTime(s string) (string, error) {
	t, err := time.Parse(time.DateTime, s)
	if err != nil {
		return "", fmt.Errorf("invalid time format, expected %q: %w", time.DateTime, err)
	}
	return t.Format(time.DateTime), nil
}
