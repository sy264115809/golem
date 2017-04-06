package bsonbuilder

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"gopkg.in/mgo.v2/bson"
)

// IsNumeric returns true if a val is numeric.
func IsNumeric(val interface{}) bool {
	switch val.(type) {
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64:
		return true
	}
	return false
}

// IsString returns true if a val is string.
func IsString(val interface{}) bool {
	switch val.(type) {
	case string:
		return true
	}
	return false
}

// IsDateTime returns true if a val is datetime.
func IsDateTime(val interface{}) bool {
	switch val.(type) {
	case time.Time, *time.Time:
		return true
	}
	return false
}

// IsObjectID returns true if a val is ObjectID.
func IsObjectID(val interface{}) bool {
	switch val.(type) {
	case bson.ObjectId:
		return true
	}
	return false
}

// IsComparable returns true if a val is comparable.
func IsComparable(val interface{}) bool {
	return IsNumeric(val) || IsString(val) || IsDateTime(val) || IsObjectID(val)
}

// Max returns the maximum value in vals if each elements in vals is comparable, otherwise returns nil.
func Max(vals []interface{}) interface{} {
	vals = Sort(vals)
	if len(vals) == 0 {
		return nil
	}
	return vals[len(vals)-1]
}

// Min returns the minimum value in vals if each elements in vals is comparable, otherwise returns nil.
func Min(vals []interface{}) interface{} {
	vals = Sort(vals)
	if len(vals) == 0 {
		return nil
	}
	return vals[0]
}

// Sort sorts returns a sorted vals in ascending order if comparable, otherwise returns nil.
func Sort(vals []interface{}) []interface{} {
	hasSameType := true
	sort.Slice(vals, func(i, j int) bool {
		vi, vj := vals[i], vals[j]
		switch {
		case IsNumeric(vi) && IsNumeric(vj):
			fi, _ := strconv.ParseFloat(fmt.Sprint(vi), 64)
			fj, _ := strconv.ParseFloat(fmt.Sprint(vj), 64)
			return fi < fj
		case IsString(vi) && IsString(vj):
			return strings.Compare(vi.(string), vj.(string)) == -1
		case IsDateTime(vi) && IsDateTime(vj):
			return (vi.(time.Time)).Before(vj.(time.Time))
		case IsObjectID(vi) && IsObjectID(vj):
			return strings.Compare(vi.(bson.ObjectId).Hex(), vj.(bson.ObjectId).Hex()) == -1
		default:
			hasSameType = false
		}
		return false
	})

	if !hasSameType {
		return nil
	}

	return vals
}
