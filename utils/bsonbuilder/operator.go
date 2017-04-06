package bsonbuilder

import (
	"fmt"

	"gopkg.in/mgo.v2/bson"
)

// Operator represents the query operation.
type Operator int

const (
	OperatorEq Operator = iota
	OperatorNe
	OperatorGt
	OperatorGte
	OperatorLt
	OperatorLte
	OperatorLike
)

// ToBSON converts operator to bson.
func (op Operator) ToBSON(vals ...interface{}) bson.M {
	if len(vals) == 0 {
		return nil
	}

	isSingle := len(vals) == 1
	switch op {
	case OperatorEq:
		if isSingle {
			return bson.M{"$eq": vals[0]}
		}
		return bson.M{"$in": vals}

	case OperatorNe:
		if isSingle {
			return bson.M{"$ne": vals[0]}
		}
		return bson.M{"$nin": vals}

	case OperatorGt, OperatorGte:
		max := Max(vals)
		if max == nil {
			return nil
		}
		if op == OperatorGt {
			return bson.M{"$gt": max}
		}
		return bson.M{"$gte": max}

	case OperatorLt, OperatorLte:
		min := Min(vals)
		if min == nil {
			return nil
		}
		if op == OperatorLt {
			return bson.M{"$lt": min}
		}
		return bson.M{"$lte": min}

	case OperatorLike:
		return bson.M{
			"$regex": bson.RegEx{
				Pattern: fmt.Sprint(vals[0]),
			},
		}

	default:
		return nil
	}
}
