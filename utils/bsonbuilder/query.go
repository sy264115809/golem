package bsonbuilder

import (
	"github.com/imdario/mergo"

	"gopkg.in/mgo.v2/bson"
)

// Query can describe a query through the fields, operators and conditions.
type Query interface {
	Add(key string, op Operator, val ...interface{})
	ToBSON() bson.M
}

type query map[string]map[Operator][]interface{}

var _ Query = make(query)

// New instances a new Query object.
func New() Query {
	q := make(query)
	q.init()
	return q
}

func (q *query) init(keys ...string) {
	if *q == nil {
		*q = make(query)
	}
	for _, key := range keys {
		if (*q)[key] == nil {
			(*q)[key] = make(map[Operator][]interface{})
		}
	}
}

// Add adds a val with given key.
func (q query) Add(key string, op Operator, val ...interface{}) {
	q.init(key)
	q[key][op] = append(q[key][op], val...)
}

// ToBSON converts Query q to bson query format.
func (q query) ToBSON() bson.M {
	q.init()
	bm := make(bson.M)
	for key, ops := range q {
		qs := make([]bson.M, 0)
		for op, conditions := range ops {
			qs = append(qs, op.ToBSON(conditions...))
		}
		bm[key] = merge(qs)
	}
	return bm
}

func merge(qs []bson.M) bson.M {
	bm := make(bson.M)
	for _, q := range qs {
		mergo.MergeWithOverwrite(&bm, q)
	}
	return bm
}
