package mgobase

import (
	"math"
	"sort"
	"strconv"
)

type (
	// Paginater interface
	Paginater interface {
		Page() int
		Limit() int
		TotalItems() int
		TotalPages() int

		HasPrev() bool
		PrevPage() int
		HasNext() bool
		NextPage() int
		IterRange(leftEdge, leftCurrent, rightCurrent, RightEdge int) []PageItem
	}

	// PageItem type
	PageItem struct {
		PageNum       int
		IsPlaceHolder bool
	}
)

type (
	paginater struct {
		page       int
		limit      int
		count      int
		totalPages int
	}
)

var _ Paginater = paginater{}

// NewPaginater instances a paginater implements Paginater interface.
func NewPaginater(skip, limit, count int) Paginater {
	if skip < 0 {
		skip = 0
	}

	if limit < 0 {
		limit = -limit
	}

	page := 1
	if limit > 0 {
		page = (skip / limit) + 1
	}

	if count < 0 {
		count = 0
	}

	totalPages := 1
	if count != 0 {
		totalPages = int(math.Ceil(float64(count) / float64(limit)))
	}
	return &paginater{
		page:       page,
		limit:      limit,
		count:      count,
		totalPages: totalPages,
	}
}

func (p paginater) Page() int {
	return p.page
}

func (p paginater) Limit() int {
	return p.limit
}

func (p paginater) TotalItems() int {
	return p.count
}

func (p paginater) TotalPages() int {
	return p.totalPages
}

func (p paginater) HasPrev() bool {
	return p.page > 1
}

func (p paginater) PrevPage() int {
	if p.page > 1 {
		return p.page - 1
	}
	return p.page
}

func (p paginater) HasNext() bool {
	return p.page < p.totalPages
}

func (p paginater) NextPage() int {
	if p.page < p.totalPages {
		return p.page + 1
	}
	return p.page
}

func (p paginater) IterRange(leftEdge, leftCurrent, rightCurrent, rightEdge int) (items []PageItem) {
	var pages = make(map[int]struct{})

	for cursor := 1; cursor < 1+leftEdge && cursor <= p.totalPages; cursor++ {
		pages[cursor] = struct{}{}
	}

	start, end := math.Max(1.0, float64(p.page-leftCurrent)), math.Min(float64(p.page+rightCurrent), float64(p.totalPages))
	for cursor := start; cursor <= end; cursor++ {
		pages[int(cursor)] = struct{}{}
	}
	for cursor := p.totalPages; cursor > p.totalPages-rightEdge && cursor >= 1; cursor-- {
		pages[cursor] = struct{}{}
	}

	var keys []int
	for page := range pages {
		keys = append(keys, page)
	}
	sort.Ints(keys)

	lastKey := 0
	for _, key := range keys {
		if lastKey+1 != key {
			items = append(items, PageItem{IsPlaceHolder: true})
		}
		items = append(items, PageItem{PageNum: key})
		lastKey = key
	}

	return
}

func (i PageItem) String() string {
	if i.IsPlaceHolder {
		return "..."
	}
	return strconv.Itoa(i.PageNum)
}
