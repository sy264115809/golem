package mgobase

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

var (
	dsn = "mongodb://localhost:27017"
)

func init() {
	if host, ok := os.LookupEnv("MONGO_TEST_HOST"); ok {
		dsn = host
	}
}

func TestMakerQueryStatement(t *testing.T) {
	var (
		field         = "_id"
		value         = bson.NewObjectId()
		expectGtRange = bson.M{field: bson.M{"$gt": value}}
		expectLtRange = bson.M{field: bson.M{"$lt": value}}
	)

	type testcase struct {
		page     string
		baseQ    interface{}
		expected interface{}
	}

	testcases := []testcase{
		{
			page:     PageFirst,
			baseQ:    bson.M{"field": "value"},
			expected: bson.M{"field": "value"},
		},
		{
			page:     PageFirst,
			baseQ:    nil,
			expected: nil,
		},
		{
			page:     PageLast,
			baseQ:    bson.M{"field": "value"},
			expected: bson.M{"field": "value"},
		},
		{
			page:     PagePrev,
			baseQ:    bson.M{"field": "value"},
			expected: bson.M{"$and": []interface{}{bson.M{"field": "value"}, expectLtRange}},
		},
		{
			page:     PageNext,
			baseQ:    bson.M{"field": "value"},
			expected: bson.M{"$and": []interface{}{bson.M{"field": "value"}, expectGtRange}},
		},
	}

	for _, tc := range testcases {
		marker := NewMarker(field, value, tc.page)
		assert.Equal(t, tc.expected, marker.QueryStatement(tc.baseQ))
	}
}

func TestMakerSortField(t *testing.T) {
	var (
		field = "_id"
		value = bson.NewObjectId()
	)

	type testcase struct {
		page     string
		expected string
	}

	testcases := []testcase{
		{
			page:     PageFirst,
			expected: "_id",
		},
		{
			page:     PageLast,
			expected: "-_id",
		},
		{
			page:     PagePrev,
			expected: "-_id",
		},
		{
			page:     PageNext,
			expected: "_id",
		},
	}

	for _, tc := range testcases {
		marker := NewMarker(field, value, tc.page)
		assert.Equal(t, tc.expected, marker.SortField())
	}
}

func TestMakerPrevNext(t *testing.T) {
	var (
		field = "_id"
	)

	type testcase struct {
		models     interface{}
		expectPrev interface{}
		expectNext interface{}
	}

	type item struct {
		ID string `bson:"_id"`
	}

	testcases := []testcase{
		{
			models: []item{
				{"1"}, {"2"}, {"3"}, {"4"},
			},
			expectPrev: "1",
			expectNext: "4",
		},
		{
			models: &[]item{
				{"1"}, {"2"}, {"3"}, {"4"},
			},
			expectPrev: "1",
			expectNext: "4",
		},
		{
			models:     []item{},
			expectPrev: nil,
			expectNext: nil,
		},
		{
			models:     1,
			expectPrev: nil,
			expectNext: nil,
		},
		{
			models: []map[string]interface{}{
				{"_id": 1}, {"_id": 2}, {"_id": 3}, {"_id": 4},
			},
			expectPrev: 1,
			expectNext: 4,
		},
	}

	for _, tc := range testcases {
		marker := NewMarker(field, "value", PageFirst)
		prev, next := marker.PrevNext(tc.models)
		assert.Equal(t, tc.expectPrev, prev)
		assert.Equal(t, tc.expectNext, next)
	}
}

func TestMarkerList(t *testing.T) {
	sess, err := mgo.DialWithTimeout(dsn, 10*time.Second)
	assert.NoError(t, err)
	defer sess.Close()

	type item struct {
		Mark string `bson:"mark"`
		Type int    `bson:"type"`
	}

	type testcase struct {
		items      []item
		limit      int
		val        string
		page       string
		query      interface{}
		expectPrev string
		expectNext string
	}

	testcases := []testcase{
		{
			items: []item{
				{"1", 1},
				{"2", 1},
				{"3", 1},
				{"4", 1},
			},
			limit:      3,
			val:        "2",
			page:       PageFirst,
			query:      nil,
			expectPrev: "1",
			expectNext: "3",
		},
		{
			items: []item{
				{"1", 1},
				{"2", 1},
				{"3", 1},
				{"4", 1},
			},
			limit:      3,
			val:        "2",
			page:       PageLast,
			query:      nil,
			expectPrev: "2",
			expectNext: "4",
		},
		{
			items: []item{
				{"1", 1},
				{"2", 1},
				{"3", 1},
				{"4", 1},
			},
			limit:      2,
			val:        "2",
			page:       PageNext,
			query:      nil,
			expectPrev: "3",
			expectNext: "4",
		},
		{
			items: []item{
				{"1", 1},
				{"2", 1},
				{"3", 1},
				{"4", 1},
			},
			limit:      2,
			val:        "4",
			page:       PagePrev,
			query:      nil,
			expectPrev: "2",
			expectNext: "3",
		},
		{
			items: []item{
				{"1", 2},
				{"2", 2},
				{"3", 1},
				{"4", 2},
				{"5", 1},
				{"6", 2},
			},
			limit:      2,
			val:        "2",
			page:       PageNext,
			query:      bson.M{"type": 2},
			expectPrev: "4",
			expectNext: "6",
		},
		{
			items: []item{
				{"1", 2},
				{"2", 2},
				{"3", 1},
				{"4", 2},
				{"5", 1},
				{"6", 2},
			},
			limit:      4,
			val:        "2",
			page:       PageFirst,
			query:      bson.M{"type": 2},
			expectPrev: "1",
			expectNext: "6",
		},
		{
			items: []item{
				{"1", 2},
				{"2", 2},
				{"3", 1},
				{"4", 2},
				{"5", 1},
				{"6", 2},
			},
			limit:      3,
			val:        "2",
			page:       PageLast,
			query:      bson.M{"type": 2},
			expectPrev: "2",
			expectNext: "6",
		},
		{
			items: []item{
				{"1", 2},
				{"2", 2},
				{"3", 1},
				{"4", 2},
				{"5", 1},
				{"6", 2},
			},
			limit:      1,
			val:        "4",
			page:       PagePrev,
			query:      bson.M{"type": 2},
			expectPrev: "2",
			expectNext: "2",
		},
	}

	for _, tc := range testcases {
		col := sess.DB("golem_test").C("mgobase")
		for _, item := range tc.items {
			col.Insert(item)
		}

		var items []item
		marker := NewMarker("mark", tc.val, tc.page)
		prev, next, err := marker.List(col, tc.query, nil, &items, tc.limit)

		assert.NoError(t, err)
		assert.Equal(t, tc.expectPrev, prev)
		assert.Equal(t, tc.expectNext, next)

		col.Database.DropDatabase()
	}
}
