package bsonbuilder_test

import (
	"testing"

	"gopkg.in/mgo.v2/bson"

	"time"

	"github.com/stretchr/testify/assert"
	"github.com/sy264115809/golem/utils/bsonbuilder"
)

func TestQueryToBSON(t *testing.T) {
	q := bsonbuilder.New()
	q.Add("name", bsonbuilder.OperatorEq, "tom")
	q.Add("name", bsonbuilder.OperatorEq, "jery")
	q.Add("name", bsonbuilder.OperatorNe, "mary")
	q.Add("age", bsonbuilder.OperatorGt, 10)
	q.Add("age", bsonbuilder.OperatorLte, 20)
	q.Add("born_at", bsonbuilder.OperatorGte, time.Date(1990, 4, 1, 20, 0, 0, 0, time.Local))
	q.Add("email", bsonbuilder.OperatorLike, "@company.com")
	q.Add("is_active", bsonbuilder.OperatorNe, false)

	expected := bson.M{
		"name":      bson.M{"$in": []interface{}{"tom", "jery"}, "$ne": "mary"},
		"age":       bson.M{"$gt": 10, "$lte": 20},
		"born_at":   bson.M{"$gte": time.Date(1990, 4, 1, 20, 0, 0, 0, time.Local)},
		"email":     bson.M{"$regex": bson.RegEx{Pattern: "@company.com"}},
		"is_active": bson.M{"$ne": false},
	}

	assert.Equal(t, expected, q.ToBSON())
}
