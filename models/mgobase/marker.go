package mgobase

import (
	"errors"
	"reflect"

	"github.com/fatih/structs"
	"github.com/sy264115809/golem/utils/slice"

	"strings"

	"fmt"

	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

const (
	// PageFirst is the first page.
	PageFirst = "first"
	// PageLast is the last page.
	PageLast = "last"
	// PageNext is the next page.
	PageNext = "next"
	// PagePrev is the previous page.
	PagePrev = "previous"
)

// Marker is the field which is the query condition in range-based paging query.
// Normally, a collection which hopes to use marker pagination should always have a default marker field.
type Marker interface {
	List(col *mgo.Collection, query, selector, models interface{}, limit int) (prev, next interface{}, err error)

	QueryStatement(baseQuery interface{}) interface{}
	SortField() string
	PrevNext(result interface{}) (prev, next interface{})
}

type marker struct {
	field string
	value interface{}
	page  string
}

// NewMarker returns a marker.
func NewMarker(field string, value interface{}, page string) Marker {
	return &marker{
		field: field,
		value: value,
		page:  page,
	}
}

func (m *marker) validate() error {
	if m.field == "" {
		return errors.New("marker's field can't be empty")
	}

	switch m.page {
	case PageFirst, PageLast: // no more validation
	case PageNext, PagePrev:
		if m.value == nil {
			return fmt.Errorf("marker's value can't be nil when page equals to %s", m.page)
		}
	default:
		return errors.New("invalid page type")
	}

	return nil
}

// QueryStatement makes a range-based query based on the `baseQuery` according to the target page type.
func (m *marker) QueryStatement(baseQuery interface{}) interface{} {
	switch m.page {
	case PageNext:
		rangeQuery := bson.M{m.field: bson.M{"$gt": m.value}}
		if baseQuery == nil {
			return rangeQuery
		}
		return bson.M{"$and": []interface{}{baseQuery, rangeQuery}}
	case PagePrev:
		rangeQuery := bson.M{m.field: bson.M{"$lt": m.value}}
		if baseQuery == nil {
			return rangeQuery
		}
		return bson.M{"$and": []interface{}{baseQuery, rangeQuery}}
	}
	return baseQuery
}

func (m *marker) SortField() string {
	if m.isReverse() {
		return "-" + m.field
	}
	return m.field
}

func (m *marker) PrevNext(result interface{}) (prev, next interface{}) {
	v := reflect.ValueOf(result)
	for v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	if v.Kind() != reflect.Slice || v.Len() == 0 {
		return
	}

	elemt := v.Type().Elem()
	switch {
	case elemt.Kind() == reflect.Struct || elemt.Elem().Kind() == reflect.Struct:
		prev = field(v.Index(0).Interface(), m.field)
		next = field(v.Index(v.Len()-1).Interface(), m.field)
	case elemt.Kind() == reflect.Map:
		prev = v.Index(0).MapIndex(reflect.ValueOf(m.field)).Interface()
		next = v.Index(v.Len() - 1).MapIndex(reflect.ValueOf(m.field)).Interface()
	}
	return
}

func (m *marker) List(col *mgo.Collection, query, selector, models interface{}, limit int) (prev, next interface{}, err error) {
	if err = m.validate(); err != nil {
		return
	}

	err = col.Find(m.QueryStatement(query)).Select(selector).Sort(m.SortField()).Limit(limit).All(models)
	if err != nil {
		return
	}

	if m.isReverse() {
		slice.Reverse(models)
	}

	prev, next = m.PrevNext(models)
	return
}

func (m *marker) isReverse() bool {
	return m.page == PageLast || m.page == PagePrev
}

func field(s interface{}, alias string) interface{} {
	if structs.IsStruct(s) {
		for _, f := range structs.Fields(s) {
			if f.Name() == alias {
				return f.Value()
			}

			if tag := f.Tag("bson"); tag != "" && alias == strings.Split(tag, ",")[0] {
				return f.Value()
			}
		}
	}
	return nil
}
