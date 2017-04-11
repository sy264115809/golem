package controllers

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/sy264115809/golem/utils/bsonbuilder"

	iris "gopkg.in/kataras/iris.v6"
	"gopkg.in/mgo.v2/bson"
)

/***********************************************/
/*************** CONTEXT BINDER ****************/
/***********************************************/

// Validate validates struct type obj.
func (c *Base) Validate(obj interface{}) (err error) {
	if c.validateFunc != nil {
		return c.validateFunc(obj)
	}
	return nil
}

// BindJSON binds application/json content-type body and validates it
func (c *Base) BindJSON(ctx *iris.Context, obj interface{}) (err error) {
	if err = ctx.ReadJSON(&obj); err != nil {
		return
	}

	return c.Validate(obj)
}

// BindXML binds application/xml content-type body and validates it
func (c *Base) BindXML(ctx *iris.Context, obj interface{}) (err error) {
	if err = ctx.ReadXML(&obj); err != nil {
		return
	}

	return c.Validate(obj)
}

// BindForm binds application/x-www-form-urlencode content-type body and validates it
func (c *Base) BindForm(ctx *iris.Context, obj interface{}) (err error) {
	if err = ctx.ReadForm(obj); err != nil {
		return
	}
	return c.Validate(obj)
}

/*********************************************/
/*************** QUERY PARSER ****************/
/*********************************************/

// QueryString parses query parameter as string if key exist, or returns defaultVal.
func (c *Base) QueryString(ctx *iris.Context, key string, defaultVal string) string {
	if str := ctx.URLParam(key); str != "" {
		return str
	}
	return defaultVal
}

// QueryStrings parses query parameter as string array if key exist, or returns defaultVal.
func (c *Base) QueryStrings(ctx *iris.Context, key string, defaultVal []string) []string {
	if strs, ok := ctx.URLParamsAsMulti()[key]; ok {
		return strs
	}
	return defaultVal
}

// QueryInt parses query parameter as int if key exist, or returns defaultVal.
func (c *Base) QueryInt(ctx *iris.Context, key string, defaultVal int) int {
	if val, err := ctx.URLParamInt64(key); err == nil {
		return int(val)
	}
	return defaultVal
}

// QueryInts parses query parameter as int array if key exist, or returns defaultVal.
func (c *Base) QueryInts(ctx *iris.Context, key string, defaultVal []int) []int {
	if strs := c.QueryStrings(ctx, key, nil); strs != nil {
		var ints []int
		for _, str := range strs {
			if i, err := strconv.ParseInt(str, 10, 0); err == nil {
				ints = append(ints, int(i))
			}
		}
		if ints != nil {
			return ints
		}
	}
	return defaultVal
}

// QueryFloat parses query parameter as float64 if exist, or returns defaultVal.
func (c *Base) QueryFloat(ctx *iris.Context, key string, defaultVal float64) float64 {
	if str := ctx.URLParam(key); str != "" {
		if f, err := strconv.ParseFloat(str, 64); err == nil {
			return f
		}
	}
	return defaultVal
}

// QueryFloats parses query parameter as float64 array if exist, or returns defaultVal.
func (c *Base) QueryFloats(ctx *iris.Context, key string, defaultVal []float64) []float64 {
	if strs := c.QueryStrings(ctx, key, nil); strs != nil {
		var floats []float64
		for _, str := range strs {
			if f, err := strconv.ParseFloat(str, 64); err == nil {
				floats = append(floats, f)
			}
		}
		if floats != nil {
			return floats
		}
	}
	return defaultVal
}

// QueryBool parses query parameter as boolean if exist, or returns defaultVal.
func (c *Base) QueryBool(ctx *iris.Context, key string, defaultVal bool) bool {
	if str := ctx.URLParam(key); str != "" {
		if b, err := strconv.ParseBool(str); err == nil {
			return b
		}
	}
	return defaultVal
}

// QueryBools parses query parameter as boolean array if exist, or returns defaultVal.
func (c *Base) QueryBools(ctx *iris.Context, key string, defaultVal []bool) []bool {
	if strs := c.QueryStrings(ctx, key, nil); strs != nil {
		var bools []bool
		for _, str := range strs {
			if b, err := strconv.ParseBool(str); err == nil {
				bools = append(bools, b)
			}
		}
		if bools != nil {
			return bools
		}
	}
	return defaultVal
}

/************************************************/
/************* QUERY & FORM PARSER **************/
/************************************************/

// ParamString takes first value for the named component of the query.
// POST, PUT and PATCH body parameters take precedence over URL query string values.
// returns the default value if key is not present.
func (c *Base) ParamString(ctx *iris.Context, key string, defaultVal string) string {
	if strs := c.ParamStrings(ctx, key, nil); len(strs) > 0 {
		return strs[0]
	}
	return defaultVal
}

// ParamStrings takes values for the named component of the query and body (if POST, PUT or PATCH).
// returns the default value if key is not present.
func (c *Base) ParamStrings(ctx *iris.Context, key string, defaultVal []string) []string {
	if strs, ok := ctx.FormValues()[key]; ok {
		return strs
	}
	return defaultVal
}

// ParamInt takes first value for the named component of the query and parses it to int.
// POST, PUT and PATCH body parameters take precedence over URL query string values.
// returns the default value if key is not present.
func (c *Base) ParamInt(ctx *iris.Context, key string, defaultVal int) int {
	if ints := c.ParamInts(ctx, key, nil); len(ints) > 0 {
		return ints[0]
	}
	return defaultVal
}

// ParamInts takes values for the named component of the query and body (if POST, PUT or PATCH).
// returns the default value if key is not present.
func (c *Base) ParamInts(ctx *iris.Context, key string, defaultVal []int) []int {
	if strs, ok := ctx.FormValues()[key]; ok {
		var ints []int
		for _, str := range strs {
			if val, err := strconv.ParseInt(str, 10, 0); err == nil {
				ints = append(ints, int(val))
			}
		}
		if ints != nil {
			return ints
		}
	}
	return defaultVal
}

// ParamFloat takes first value for the named component of the query and parses it to float64.
// POST, PUT and PATCH body parameters take precedence over URL query string values.
// returns the default value if key is not present.
func (c *Base) ParamFloat(ctx *iris.Context, key string, defaultVal float64) float64 {
	if floats := c.ParamFloats(ctx, key, nil); len(floats) > 0 {
		return floats[0]
	}
	return defaultVal
}

// ParamFloats takes values for the named component of the query and body (if POST, PUT or PATCH).
// returns the default value if key is not present.
func (c *Base) ParamFloats(ctx *iris.Context, key string, defaultVal []float64) []float64 {
	if strs, ok := ctx.FormValues()[key]; ok {
		var floats []float64
		for _, str := range strs {
			if val, err := strconv.ParseFloat(str, 0); err == nil {
				floats = append(floats, val)
			}
		}
		if floats != nil {
			return floats
		}
	}
	return defaultVal
}

// ParamBool takes first value for the named component of the query and parses it to boolean.
// POST, PUT and PATCH body parameters take precedence over URL query string values.
// returns the default value if key is not present.
func (c *Base) ParamBool(ctx *iris.Context, key string, defaultVal bool) bool {
	if bools := c.ParamBools(ctx, key, nil); len(bools) > 0 {
		return bools[0]
	}
	return defaultVal
}

// ParamBools takes values for the named component of the query and body (if POST, PUT or PATCH).
// returns the default value if key is not present.
func (c *Base) ParamBools(ctx *iris.Context, key string, defaultVal []bool) []bool {
	if strs, ok := ctx.FormValues()[key]; ok {
		var bools []bool
		for _, str := range strs {
			if val, err := strconv.ParseBool(str); err == nil {
				bools = append(bools, val)
			}
		}
		if bools != nil {
			return bools
		}
	}
	return defaultVal
}

/************************************************/
/************* CONVENTIONAL PARSE ***************/
/************************************************/

var (
	// QueryKeyPage is the key of pagination parameter ‘page’.
	QueryKeyPage = "_page"
	// QueryKeyLimit is the key of pagination paramter 'limit' or 'per page'.
	QueryKeyLimit = "_limit"
	// DefaultPaginationPage is the default page if not provide.
	DefaultPaginationPage = 1
	// DefaultPaginationLimit is the default page if not provide.
	DefaultPaginationLimit = 20
)

var (
	// QueryKeySort is the key of sort parameter which defines what field should be sorted on.
	QueryKeySort = "_sort"
	// QueryKeyOrder is the key of order paramter which affects the order of query result.
	QueryKeyOrder = "_order"
	// OrderAscending represents ascending order.
	OrderAscending = "asc"
	// OrderDescending represents descending order.
	OrderDescending = "desc"
)

var (
	// QueryFormatDateTime is the formatter to parse the datetime parameter.
	QueryFormatDateTime = "2006-01-02T15:04:05"
)

// Pagination parses the pagination info from query params.
func (c *Base) Pagination(ctx *iris.Context) (page, skip, limit int) {
	page = c.QueryInt(ctx, QueryKeyPage, DefaultPaginationPage)
	if page <= 0 {
		page = DefaultPaginationPage
	}

	limit = c.QueryInt(ctx, QueryKeyLimit, DefaultPaginationLimit)
	if limit <= 0 {
		limit = DefaultPaginationLimit
	}

	skip = (page - 1) * limit
	return
}

// Sort parses the sort info from query params.
// Suppose the sort settings are all defaults, there are two valid forms:
// - json-server style: ?_sort=filed&order=DESC
// - mgo style: ?_sort=-field
// The order parameter is case-insensitive.
func (c *Base) Sort(ctx *iris.Context) (sort, order string) {
	sort = c.QueryString(ctx, QueryKeySort, "")
	order = strings.ToLower(c.QueryString(ctx, QueryKeyOrder, ""))
	if sort != "" {
		if strings.HasPrefix(sort, "-") {
			sort = sort[1:]
			order = OrderDescending
			return
		}

		if order != OrderAscending && order != OrderDescending {
			order = OrderAscending
		}
	}
	return
}

// SortAsMultiple parses the sort info from query params and convert to mgo style.
// The rules:
// 1. If the sort parameter is empty, return an empty array.
// 2. Elseif the sort parameter can be splited by comma, split it and return.
// 3. Else the order will be considered:
//   3.1 If the order is `DESC`, prepend a minus to the sort parameter.
//   3.2 return an array contains the (may be changed)sort parameter.
func (c *Base) SortAsMultiple(ctx *iris.Context) (sorts []string) {
	sort := c.QueryString(ctx, QueryKeySort, "")
	if sort != "" {
		if s := strings.Split(sort, ","); len(s) > 1 {
			return s
		}

		order := strings.ToLower(c.QueryString(ctx, QueryKeyOrder, ""))
		if !strings.HasPrefix(sort, "-") && order == OrderDescending {
			sort = fmt.Sprintf("-%s", sort)
		}
		sorts = append(sorts, sort)
	}
	return
}

// QueryBSON parses query paramter as bson.M follow the same practice as the JSON-Server is.
// JSON-Server's Home: https://github.com/typicode/json-server
// Glance:
// - Add `_gt`, `_gte` or `_lt`, `_lte` for getting a range
// - Add `_ne` to exclude a value
// - Add `_like` to filter (RegExp supported)
// - Use `.` to access deep properties
func (c *Base) QueryBSON(ctx *iris.Context, convertors ...func(key, value string) interface{}) bson.M {
	var (
		params = ctx.Request.URL.Query()
		query  = bsonbuilder.New()
	)

	for key := range params {
		// skip pagination & sort keys
		switch key {
		case QueryKeyPage, QueryKeyLimit, QueryKeySort, QueryKeyOrder:
			continue
		}

		for _, val := range params[key] {

			// handle suffix _gt, _gte, _lt, _lte, _ne, _like, decides what operator should be used
			actualKey, op := key, bsonbuilder.OperatorEq
			for suffix, o := range map[string]bsonbuilder.Operator{
				"_gt":   bsonbuilder.OperatorGt,
				"_gte":  bsonbuilder.OperatorGte,
				"_lt":   bsonbuilder.OperatorLt,
				"_lte":  bsonbuilder.OperatorLte,
				"_ne":   bsonbuilder.OperatorNe,
				"_like": bsonbuilder.OperatorLike,
			} {
				if strings.HasSuffix(actualKey, suffix) {
					op = o
					actualKey = strings.TrimSuffix(actualKey, suffix)
					break
				}
			}
			switch {
			// "id" => "_id"
			case actualKey == "id":
				if bson.IsObjectIdHex(val) {
					query.Add("_id", op, bson.ObjectIdHex(val))
					continue
				}

			// "user.id" => "user_id"
			case strings.Contains(actualKey, ".id"):
				if bson.IsObjectIdHex(val) {
					actualKey = strings.Replace(actualKey, ".id", "._id", -1)
					query.Add(actualKey, op, bson.ObjectIdHex(val))
					continue
				}
			}

			// try to convert value to int
			if i, err := strconv.ParseInt(val, 10, 0); err == nil {
				query.Add(actualKey, op, int(i))
				continue
			}

			// try to convert value to float64
			if f64, err := strconv.ParseFloat(val, 64); err == nil {
				query.Add(actualKey, op, f64)
				continue
			}

			// try to convert value to boolean
			if b, err := strconv.ParseBool(val); err == nil {
				query.Add(actualKey, op, b)
				continue
			}

			// try to convert value to time.Time
			if date, err := time.Parse(QueryFormatDateTime, val); err == nil {
				query.Add(actualKey, op, date)
				continue
			}

			// try to convert value using custom convertor(s)
			skip := false
			for _, c := range convertors {
				if res := c(actualKey, val); res != nil {
					query.Add(actualKey, op, res)
					skip = true
					break
				}
			}

			if !skip {
				query.Add(actualKey, op, val)
			}
		}

	}

	return query.ToBSON()
}
