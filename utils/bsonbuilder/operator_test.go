package bsonbuilder_test

import (
	"testing"
	"time"

	"gopkg.in/mgo.v2/bson"

	"github.com/stretchr/testify/assert"
	"github.com/sy264115809/golem/utils/bsonbuilder"
)

func TestOperatorToBSON(t *testing.T) {
	type testcase struct {
		desc     string
		op       bsonbuilder.Operator
		vals     []interface{}
		expected bson.M
	}

	var (
		date1, date2 = time.Now(), time.Now().Add(10 * time.Second)
		oid1, oid2   = bson.NewObjectIdWithTime(date1), bson.NewObjectIdWithTime(date2)
	)

	testcases := []testcase{
		{
			desc:     "OperatorEq with single value",
			op:       bsonbuilder.OperatorEq,
			vals:     []interface{}{"value"},
			expected: bson.M{"$eq": "value"},
		},
		{
			desc:     "OperatorEq with multiple values",
			op:       bsonbuilder.OperatorEq,
			vals:     []interface{}{"value1", "value2", 2, true},
			expected: bson.M{"$in": []interface{}{"value1", "value2", 2, true}},
		},
		{
			desc:     "OperatorNe with single value",
			op:       bsonbuilder.OperatorNe,
			vals:     []interface{}{"value"},
			expected: bson.M{"$ne": "value"},
		},
		{
			desc:     "OperatorNe with multiple values",
			op:       bsonbuilder.OperatorNe,
			vals:     []interface{}{"value1", "value2", 1, false},
			expected: bson.M{"$nin": []interface{}{"value1", "value2", 1, false}},
		},
		{
			desc:     "OperatorGt with single value",
			op:       bsonbuilder.OperatorGt,
			vals:     []interface{}{date1},
			expected: bson.M{"$gt": date1},
		},
		{
			desc:     "OperatorGt with integers",
			op:       bsonbuilder.OperatorGt,
			vals:     []interface{}{3, 2, 4, 1, 5},
			expected: bson.M{"$gt": 5},
		},
		{
			desc:     "OperatorGt with strings",
			op:       bsonbuilder.OperatorGt,
			vals:     []interface{}{"a", "b", "c", "d", "dd"},
			expected: bson.M{"$gt": "dd"},
		},
		{
			desc:     "OperatorGt with object ids",
			op:       bsonbuilder.OperatorGt,
			vals:     []interface{}{oid1, oid2},
			expected: bson.M{"$gt": oid2},
		},
		{
			desc:     "OperatorGt with dates",
			op:       bsonbuilder.OperatorGt,
			vals:     []interface{}{date1, date2},
			expected: bson.M{"$gt": date2},
		},
		{
			desc:     "OperatorGt with mixed type",
			op:       bsonbuilder.OperatorGt,
			vals:     []interface{}{1, 2, "2", oid1, date1, date2},
			expected: nil,
		},
		{
			desc:     "OperatorGte with single value",
			op:       bsonbuilder.OperatorGte,
			vals:     []interface{}{date1},
			expected: bson.M{"$gte": date1},
		},
		{
			desc:     "OperatorGte with integers",
			op:       bsonbuilder.OperatorGte,
			vals:     []interface{}{3, 2, 4, 1, 5},
			expected: bson.M{"$gte": 5},
		},
		{
			desc:     "OperatorGte with strings",
			op:       bsonbuilder.OperatorGte,
			vals:     []interface{}{"a", "b", "c", "d", "dd"},
			expected: bson.M{"$gte": "dd"},
		},
		{
			desc:     "OperatorGte with object ids",
			op:       bsonbuilder.OperatorGte,
			vals:     []interface{}{oid1, oid2},
			expected: bson.M{"$gte": oid2},
		},
		{
			desc:     "OperatorGte with dates",
			op:       bsonbuilder.OperatorGte,
			vals:     []interface{}{date1, date2},
			expected: bson.M{"$gte": date2},
		},
		{
			desc:     "OperatorGte with mixed type",
			op:       bsonbuilder.OperatorGte,
			vals:     []interface{}{1, 2, "2", oid1, date1, date2},
			expected: nil,
		},
		{
			desc:     "OperatorLt with single value",
			op:       bsonbuilder.OperatorLt,
			vals:     []interface{}{date1},
			expected: bson.M{"$lt": date1},
		},
		{
			desc:     "OperatorLt with integers",
			op:       bsonbuilder.OperatorLt,
			vals:     []interface{}{3, 2, 4, 1, 5},
			expected: bson.M{"$lt": 1},
		},
		{
			desc:     "OperatorLt with strings",
			op:       bsonbuilder.OperatorLt,
			vals:     []interface{}{"a", "b", "c", "d", "dd"},
			expected: bson.M{"$lt": "a"},
		},
		{
			desc:     "OperatorLt with object ids",
			op:       bsonbuilder.OperatorLt,
			vals:     []interface{}{oid1, oid2},
			expected: bson.M{"$lt": oid1},
		},
		{
			desc:     "OperatorLt with dates",
			op:       bsonbuilder.OperatorLt,
			vals:     []interface{}{date1, date2},
			expected: bson.M{"$lt": date1},
		},
		{
			desc:     "OperatorLt with mixed type",
			op:       bsonbuilder.OperatorLt,
			vals:     []interface{}{1, 2, "2", oid1, date1, date2},
			expected: nil,
		},
		{
			desc:     "OperatorLte with single value",
			op:       bsonbuilder.OperatorLte,
			vals:     []interface{}{date1},
			expected: bson.M{"$lte": date1},
		},
		{
			desc:     "OperatorLte with integers",
			op:       bsonbuilder.OperatorLte,
			vals:     []interface{}{3, 2, 4, 1, 5},
			expected: bson.M{"$lte": 1},
		},
		{
			desc:     "OperatorLte with strings",
			op:       bsonbuilder.OperatorLte,
			vals:     []interface{}{"a", "b", "c", "d", "dd"},
			expected: bson.M{"$lte": "a"},
		},
		{
			desc:     "OperatorLte with object ids",
			op:       bsonbuilder.OperatorLte,
			vals:     []interface{}{oid1, oid2},
			expected: bson.M{"$lte": oid1},
		},
		{
			desc:     "OperatorLte with dates",
			op:       bsonbuilder.OperatorLte,
			vals:     []interface{}{date1, date2},
			expected: bson.M{"$lte": date1},
		},
		{
			desc:     "OperatorLte with mixed type",
			op:       bsonbuilder.OperatorLte,
			vals:     []interface{}{1, 2, "2", oid1, date1, date2},
			expected: nil,
		},
		{
			desc:     "OperatorLike with single numeric value",
			op:       bsonbuilder.OperatorLike,
			vals:     []interface{}{1},
			expected: bson.M{"$regex": bson.RegEx{Pattern: "1"}},
		},
		{
			desc:     "OperatorLike with single string value",
			op:       bsonbuilder.OperatorLike,
			vals:     []interface{}{"abc"},
			expected: bson.M{"$regex": bson.RegEx{Pattern: "abc"}},
		},
		{
			desc:     "OperatorLike with multiple numeric value",
			op:       bsonbuilder.OperatorLike,
			vals:     []interface{}{1, 2, 3},
			expected: bson.M{"$regex": bson.RegEx{Pattern: "1"}},
		},
		{
			desc:     "OperatorLike with multiple string value",
			op:       bsonbuilder.OperatorLike,
			vals:     []interface{}{"a", "b", "c"},
			expected: bson.M{"$regex": bson.RegEx{Pattern: "a"}},
		},
	}

	for _, tc := range testcases {
		assert.Equal(t, tc.expected, tc.op.ToBSON(tc.vals...), tc.desc)
	}
}
