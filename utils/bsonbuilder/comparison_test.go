package bsonbuilder_test

import (
	"testing"
	"time"

	"gopkg.in/mgo.v2/bson"

	"github.com/stretchr/testify/assert"
	"github.com/sy264115809/golem/utils/bsonbuilder"
)

func TestIsNumeric(t *testing.T) {
	numbers := []interface{}{1, int8(2), int16(3), int32(4), int64(-5), uint(6), uint8(7), uint16(8), uint32(9), uint64(10), float32(11.0), float64(12.0)}
	for _, number := range numbers {
		assert.Equal(t, true, bsonbuilder.IsNumeric(number))
	}

	nonNumbers := []interface{}{"string", true, false, make(chan int), struct{}{}, time.Now(), bson.NewObjectId()}
	for _, nn := range nonNumbers {
		assert.Equal(t, false, bsonbuilder.IsNumeric(nn))
	}
}

func TestIsString(t *testing.T) {
	strings := []interface{}{"a", "11111", "@@", "S22!"}
	for _, s := range strings {
		assert.Equal(t, true, bsonbuilder.IsString(s))
	}

	nonStrings := []interface{}{1, false, true, &struct{}{}, time.Now(), bson.NewObjectId()}
	for _, ns := range nonStrings {
		assert.Equal(t, false, bsonbuilder.IsString(ns))
	}
}

func TestIsDateTime(t *testing.T) {
	assert.Equal(t, true, bsonbuilder.IsDateTime(time.Now()))

	nonDateTimes := []interface{}{1, false, true, "ss", bson.NewObjectId()}
	for _, ndt := range nonDateTimes {
		assert.Equal(t, false, bsonbuilder.IsDateTime(ndt))
	}
}

func TestIsObjectID(t *testing.T) {
	assert.Equal(t, true, bsonbuilder.IsObjectID(bson.NewObjectId()))

	nonObjectIDs := []interface{}{1, "ss", time.Now()}
	for _, noid := range nonObjectIDs {
		assert.Equal(t, false, bsonbuilder.IsObjectID(noid))
	}
}

func TestIsComparable(t *testing.T) {
	comparables := []interface{}{1, "s", time.Now(), bson.NewObjectId()}
	for _, c := range comparables {
		assert.Equal(t, true, bsonbuilder.IsComparable(c))
	}

	nonComparables := []interface{}{true, make(chan int), make([]string, 0), make(map[string]bool), new(struct{})}
	for _, nc := range nonComparables {
		assert.Equal(t, false, bsonbuilder.IsComparable(nc))
	}
}

func TestSort(t *testing.T) {
	type testcase struct {
		desc     string
		val      []interface{}
		expected []interface{}
	}

	var (
		d1, d2, d3 = time.Now(), time.Now().Add(10 * time.Second), time.Now().Add(10 * time.Minute)
		o1, o2, o3 = bson.NewObjectIdWithTime(d1), bson.NewObjectIdWithTime(d2), bson.NewObjectIdWithTime(d3)
	)

	testcases := []testcase{
		{
			desc:     "Sort with integers",
			val:      []interface{}{3, 2, 6, 1, 0},
			expected: []interface{}{0, 1, 2, 3, 6},
		},
		{
			desc:     "Sort with integers & uints",
			val:      []interface{}{-3, 2, 6, 1, 0, -2},
			expected: []interface{}{-3, -2, 0, 1, 2, 6},
		},
		{
			desc:     "Sort with integers & floats",
			val:      []interface{}{3, 2, 6, 1, 0, 2.1, 2.1111},
			expected: []interface{}{0, 1, 2, 2.1, 2.1111, 3, 6},
		},
		{
			desc:     "Sort with strings",
			val:      []interface{}{"c", "a", "c1", "b", "cd", "d"},
			expected: []interface{}{"a", "b", "c", "c1", "cd", "d"},
		},
		{
			desc:     "Sort with datetimes",
			val:      []interface{}{d2, d1, d3},
			expected: []interface{}{d1, d2, d3},
		},
		{
			desc:     "Sort with object ids",
			val:      []interface{}{o3, o1, o2},
			expected: []interface{}{o1, o2, o3},
		},
		{
			desc:     "Sort with miexd type values",
			val:      []interface{}{1, 2, "b", 3, "c", "a", d1, d3, o2, o1},
			expected: nil,
		},
	}

	for _, tc := range testcases {
		assert.Equal(t, tc.expected, bsonbuilder.Sort(tc.val), tc.desc)
	}
}

func TestMax(t *testing.T) {
	type testcase struct {
		desc     string
		val      []interface{}
		expected interface{}
	}

	var (
		d1, d2, d3 = time.Now(), time.Now().Add(10 * time.Second), time.Now().Add(10 * time.Minute)
		o1, o2, o3 = bson.NewObjectIdWithTime(d1), bson.NewObjectIdWithTime(d2), bson.NewObjectIdWithTime(d3)
	)

	testcases := []testcase{
		{
			desc:     "Max number",
			val:      []interface{}{-3, 2, 6.1, 6.66, 1, 0},
			expected: 6.66,
		},
		{
			desc:     "Max string",
			val:      []interface{}{"aaaa", "bbbbb", "abbbbb", "bbbbbc"},
			expected: "bbbbbc",
		},
		{
			desc:     "Max datetime",
			val:      []interface{}{d2, d1, d3},
			expected: d3,
		},
		{
			desc:     "Max object id",
			val:      []interface{}{o3, o2, o1},
			expected: o3,
		},
	}

	for _, tc := range testcases {
		assert.Equal(t, tc.expected, bsonbuilder.Max(tc.val), tc.desc)
	}
}

func TestMin(t *testing.T) {
	type testcase struct {
		desc     string
		val      []interface{}
		expected interface{}
	}

	var (
		d1, d2, d3 = time.Now(), time.Now().Add(10 * time.Second), time.Now().Add(10 * time.Minute)
		o1, o2, o3 = bson.NewObjectIdWithTime(d1), bson.NewObjectIdWithTime(d2), bson.NewObjectIdWithTime(d3)
	)

	testcases := []testcase{
		{
			desc:     "Min number",
			val:      []interface{}{-3, 2, 6.1, 6.66, 1, 0},
			expected: -3,
		},
		{
			desc:     "Min string",
			val:      []interface{}{"aaaa", "bbbbb", "abbbbb", "bbbbbc"},
			expected: "aaaa",
		},
		{
			desc:     "Min datetime",
			val:      []interface{}{d2, d1, d3},
			expected: d1,
		},
		{
			desc:     "Min object id",
			val:      []interface{}{o3, o2, o1},
			expected: o1,
		},
	}

	for _, tc := range testcases {
		assert.Equal(t, tc.expected, bsonbuilder.Min(tc.val), tc.desc)
	}
}
