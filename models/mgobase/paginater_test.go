package mgobase

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPaginater(t *testing.T) {
	type testcase struct {
		desc                         string
		skip, limit, count           int
		page, totalItems, totalPages int
		hasPrev, hasNext             bool
		prevPage, nextPage           int
		iterItems                    []PageItem
	}

	var (
		leftedge     = 2
		leftcurrent  = 5
		rightcurrent = 5
		rightedge    = 2
	)

	testcases := []testcase{
		{
			desc: "no item paginater",
			skip: 0, limit: 10, count: 0,
			page: 1, totalItems: 0, totalPages: 1,
			hasPrev: false, hasNext: false,
			prevPage: 1, nextPage: 1,
			iterItems: []PageItem{
				{IsPlaceHolder: false, PageNum: 1},
			},
		},
		{
			desc: "normal paginater",
			skip: 80, limit: 10, count: 166,
			page: 9, totalItems: 166, totalPages: 17,
			hasPrev: true, hasNext: true,
			prevPage: 8, nextPage: 10,
			iterItems: func() (expect []PageItem) {
				for i := 1; i <= 17; i++ {
					if i == 3 || i == 15 {
						expect = append(expect, PageItem{IsPlaceHolder: true})
					} else {
						expect = append(expect, PageItem{PageNum: i})
					}
				}
				return
			}(),
		},
	}

	for _, tc := range testcases {
		t.Run(tc.desc, func(t *testing.T) {
			p := NewPaginater(tc.skip, tc.limit, tc.count)

			assert.Equal(t, tc.page, p.Page())
			assert.Equal(t, tc.limit, p.Limit())
			assert.Equal(t, tc.totalItems, p.TotalItems())
			assert.Equal(t, tc.totalPages, p.TotalPages())
			assert.Equal(t, tc.hasPrev, p.HasPrev())
			assert.Equal(t, tc.prevPage, p.PrevPage())
			assert.Equal(t, tc.hasNext, p.HasNext())
			assert.Equal(t, tc.nextPage, p.NextPage())

			items := p.IterRange(leftedge, leftcurrent, rightcurrent, rightedge)
			assert.Len(t, items, len(tc.iterItems))
			assert.Equal(t, tc.iterItems, items)
		})
	}

}
