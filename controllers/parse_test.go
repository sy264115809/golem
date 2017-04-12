package controllers_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/sy264115809/golem/controllers"

	"github.com/stretchr/testify/assert"
	iris "gopkg.in/kataras/iris.v6"
	"gopkg.in/kataras/iris.v6/adaptors/httprouter"
	"gopkg.in/kataras/iris.v6/httptest"
	"gopkg.in/mgo.v2/bson"
)

func baseController() *controllers.Base {
	return controllers.New()
}

func TestValidate(t *testing.T) {
	t.Run("default validator", func(t *testing.T) {
		type s struct {
			Name string `validate:"required,gt=10"`
		}

		type testcase struct {
			s      s
			hasErr bool
		}

		testcases := []testcase{
			{
				s:      s{},
				hasErr: true,
			},
			{
				s:      s{"short-name"},
				hasErr: true,
			},
			{
				s:      s{"long-long-name"},
				hasErr: false,
			},
		}

		for _, tc := range testcases {
			err := baseController().Validate(tc.s)
			assert.Equal(t, tc.hasErr, err != nil)
		}
	})

	t.Run("custom always error validator", func(t *testing.T) {
		err := errors.New("always error")
		c := baseController().SetValidateFunc(func(interface{}) error {
			return err
		})

		assert.EqualError(t, c.Validate(1), err.Error())
		assert.EqualError(t, c.Validate(-1.0), err.Error())
		assert.EqualError(t, c.Validate(true), err.Error())
		assert.EqualError(t, c.Validate(struct{}{}), err.Error())
		assert.EqualError(t, c.Validate(nil), err.Error())
		assert.EqualError(t, c.Validate(make(map[string]interface{})), err.Error())
	})

	t.Run("custom always non-error validator", func(t *testing.T) {
		c := baseController().SetValidateFunc(func(interface{}) error {
			return nil
		})

		assert.NoError(t, c.Validate(1))
		assert.NoError(t, c.Validate(-1.0))
		assert.NoError(t, c.Validate(true))
		assert.NoError(t, c.Validate(struct{}{}))
		assert.NoError(t, c.Validate(nil))
		assert.NoError(t, c.Validate(make(map[string]interface{})))
	})
}

func TestBindJSON(t *testing.T) {
	type testcase struct {
		json   string
		hasErr bool
	}

	app := iris.New()
	app.Adapt(httprouter.New())
	app.Post("/bind-json", func(ctx *iris.Context) {
		var param struct {
			Name string `json:"name" validate:"required,eq=jacy"`
			Age  int    `json:"age" validate:"gt=16"`
		}
		err := baseController().BindJSON(ctx, &param)
		if err != nil {
			ctx.WriteHeader(iris.StatusBadRequest)
			ctx.WriteString(err.Error())
		} else {
			ctx.JSON(iris.StatusOK, param)
		}
	})

	testcases := []testcase{
		{
			json:   `{"name":"jacy","age":18}`,
			hasErr: false,
		},
		{
			json:   `{"name":"jack","age":18}`,
			hasErr: true,
		},
		{
			json:   `{"name":"jacy","age":15}`,
			hasErr: true,
		},
		{
			json:   `{"name":18,"age":"jacy"}`,
			hasErr: true,
		},
	}

	for _, tc := range testcases {
		var obj interface{}
		json.Unmarshal([]byte(tc.json), &obj)

		req := httptest.New(app, t).POST("/bind-json").WithJSON(obj).Expect()

		if !tc.hasErr {
			req.Status(iris.StatusOK).Body().Equal(tc.json)
		} else {
			req.Status(iris.StatusBadRequest)
		}
	}
}

func TestBindForm(t *testing.T) {
	type testcase struct {
		form   map[string]interface{}
		hasErr bool
	}

	app := iris.New()
	app.Adapt(httprouter.New())
	app.Post("/bind-form", func(ctx *iris.Context) {
		var param struct {
			Name string `json:"name" form:"name" validate:"required,eq=jacy"`
			Age  int    `json:"age" form:"age" validate:"gt=16"`
		}
		err := baseController().BindForm(ctx, &param)
		if err != nil {
			ctx.WriteHeader(iris.StatusBadRequest)
			ctx.WriteString(err.Error())
		} else {
			ctx.JSON(iris.StatusOK, param)
		}
	})

	testcases := []testcase{
		{
			form: map[string]interface{}{
				"name": "jacy",
				"age":  18,
			},
			hasErr: false,
		},
		{
			form: map[string]interface{}{
				"name": "jack",
				"age":  18,
			},
			hasErr: true,
		},
		{
			form: map[string]interface{}{
				"name": "jacy",
				"age":  15,
			},
			hasErr: true,
		},
		{
			form: map[string]interface{}{
				"name": 18,
				"age":  "jacy",
			},
			hasErr: true,
		},
	}

	for _, tc := range testcases {
		req := httptest.New(app, t).POST("/bind-form").WithForm(tc.form).Expect()

		if !tc.hasErr {
			res := req.Status(iris.StatusOK).JSON().Object()
			for k, v := range tc.form {
				res.ValueEqual(k, v)
			}
		} else {
			req.Status(iris.StatusBadRequest)
		}
	}
}

func TestQueryString(t *testing.T) {
	type testcase struct {
		value      string
		defaultVal string
		expected   string
	}

	var (
		defaultVal string
	)

	app := iris.New()
	app.Adapt(httprouter.New())
	app.Get("/parse-string", func(ctx *iris.Context) {
		str := baseController().QueryString(ctx, "key", defaultVal)
		ctx.WriteString(str)
	})

	testcases := []testcase{
		{
			value:    "normal string",
			expected: "normal string",
		},
		{
			defaultVal: "default",
			expected:   "default",
		},
		{
			value:    "1",
			expected: "1",
		},
		{
			value:    "true",
			expected: "true",
		},
		{
			value:    "2.0",
			expected: "2.0",
		},
	}

	for _, tc := range testcases {
		defaultVal = tc.defaultVal
		httptest.New(app, t).GET("/parse-string").WithQuery("key", tc.value).
			Expect().Status(iris.StatusOK).Body().Equal(tc.expected)
	}
}

func TestQueryStrings(t *testing.T) {
	type testcase struct {
		values      []string
		defaultVals []string
		expected    []string
	}

	var (
		defaultVals []string
	)

	app := iris.New()
	app.Adapt(httprouter.New())
	app.Get("/parse-strings", func(ctx *iris.Context) {
		strs := baseController().QueryStrings(ctx, "key", defaultVals)
		ctx.JSON(iris.StatusOK, strs)
	})

	testcases := []testcase{
		{
			values:   []string{"string1", "string2", "string3"},
			expected: []string{"string1", "string2", "string3"},
		},
		{
			defaultVals: []string{"string1", "string2", "string3"},
			expected:    []string{"string1", "string2", "string3"},
		},
		{
			values:   []string{"1", "true", "string", "2.0"},
			expected: []string{"1", "true", "string", "2.0"},
		},
	}

	for _, tc := range testcases {
		defaultVals = tc.defaultVals
		req := httptest.New(app, t).GET("/parse-strings")
		for _, v := range tc.values {
			req.WithQuery("key", v)
		}
		res := req.Expect().Status(iris.StatusOK).JSON()
		if tc.expected != nil {
			res.Array().Equal(tc.expected)
		} else {
			res.Null()
		}
	}
}

func TestQueryInt(t *testing.T) {
	type testcase struct {
		value      string
		defaultVal int
		expected   int
	}

	var (
		defaultVal int
	)

	app := iris.New()
	app.Adapt(httprouter.New())
	app.Get("/parse-int", func(ctx *iris.Context) {
		i := baseController().QueryInt(ctx, "key", defaultVal)
		ctx.JSON(iris.StatusOK, i)
	})

	testcases := []testcase{
		{
			value:    "1",
			expected: 1,
		},
		{
			value:    "-1",
			expected: -1,
		},
		{
			defaultVal: 2,
			expected:   2,
		},
		{
			value:      "string",
			defaultVal: 3,
			expected:   3,
		},
		{
			value:      "-2.0",
			defaultVal: 4,
			expected:   4,
		},
		{
			value:      "false",
			defaultVal: 5,
			expected:   5,
		},
	}

	for _, tc := range testcases {
		defaultVal = tc.defaultVal
		httptest.New(app, t).GET("/parse-int").WithQuery("key", tc.value).
			Expect().Status(iris.StatusOK).JSON().Number().Equal(tc.expected)
	}
}

func TestQueryInts(t *testing.T) {
	type testcase struct {
		values      []string
		defaultVals []int
		expected    []int
	}

	var (
		defaultVals []int
	)

	app := iris.New()
	app.Adapt(httprouter.New())
	app.Get("/parse-ints", func(ctx *iris.Context) {
		ints := baseController().QueryInts(ctx, "key", defaultVals)
		ctx.JSON(iris.StatusOK, ints)
	})

	testcases := []testcase{
		{
			values:   []string{"1", "-2", "3", "4"},
			expected: []int{1, -2, 3, 4},
		},
		{
			defaultVals: []int{5, 6, -7},
			expected:    []int{5, 6, -7},
		},
		{
			values:   []string{"1", "true", "string", "-3", "2.0", "5.1"},
			expected: []int{1, -3},
		},
		{
			values:   []string{"true", "all is invalid"},
			expected: nil,
		},
		{
			values:      []string{"true", "all is invalid"},
			defaultVals: []int{1, -2},
			expected:    []int{1, -2},
		},
	}

	for _, tc := range testcases {
		defaultVals = tc.defaultVals
		req := httptest.New(app, t).GET("/parse-ints")
		for _, v := range tc.values {
			req.WithQuery("key", v)
		}
		res := req.Expect().Status(iris.StatusOK).JSON()
		if tc.expected != nil {
			res.Array().Equal(tc.expected)
		} else {
			res.Null()
		}
	}
}

func TestQueryFloat(t *testing.T) {
	type testcase struct {
		value      string
		defaultVal float64
		expected   float64
	}

	var (
		defaultVal float64
	)

	app := iris.New()
	app.Adapt(httprouter.New())
	app.Get("/parse-float", func(ctx *iris.Context) {
		f := baseController().QueryFloat(ctx, "key", defaultVal)
		ctx.JSON(iris.StatusOK, f)
	})

	testcases := []testcase{
		{
			value:    "1.1",
			expected: 1.1,
		},
		{
			defaultVal: 2.2,
			expected:   2.2,
		},
		{
			value:      "-3.3",
			defaultVal: 3,
			expected:   -3.3,
		},
		{
			value:      "string",
			defaultVal: -5,
			expected:   -5.0,
		},
		{
			value:      "false",
			defaultVal: 6,
			expected:   6.0,
		},
	}

	for _, tc := range testcases {
		defaultVal = tc.defaultVal
		httptest.New(app, t).GET("/parse-float").WithQuery("key", tc.value).
			Expect().Status(iris.StatusOK).JSON().Number().Equal(tc.expected)
	}
}

func TestQueryFloats(t *testing.T) {
	type testcase struct {
		values      []string
		defaultVals []float64
		expected    []float64
	}

	var (
		defaultVals []float64
	)

	app := iris.New()
	app.Adapt(httprouter.New())
	app.Get("/parse-floats", func(ctx *iris.Context) {
		float64s := baseController().QueryFloats(ctx, "key", defaultVals)
		ctx.JSON(iris.StatusOK, float64s)
	})

	testcases := []testcase{
		{
			values:   []string{"1", "2", "3.3", "4.0", "-2.3"},
			expected: []float64{1.0, 2.0, 3.3, 4.0, -2.3},
		},
		{
			defaultVals: []float64{5, 6.66, 7, -1},
			expected:    []float64{5, 6.66, 7, -1},
		},
		{
			values:   []string{"1", "true", "string", "3", "-2.0", "5.1"},
			expected: []float64{1.0, 3.0, -2.0, 5.1},
		},
		{
			values:   []string{"true", "all is invalid"},
			expected: nil,
		},
		{
			values:      []string{"true", "all is invalid"},
			defaultVals: []float64{-1.1, 2},
			expected:    []float64{-1.1, 2},
		},
	}

	for _, tc := range testcases {
		defaultVals = tc.defaultVals
		req := httptest.New(app, t).GET("/parse-floats")
		for _, v := range tc.values {
			req.WithQuery("key", v)
		}
		res := req.Expect().Status(iris.StatusOK).JSON()
		if tc.expected != nil {
			res.Array().Equal(tc.expected)
		} else {
			res.Null()
		}
	}
}

func TestQueryBool(t *testing.T) {
	type testcase struct {
		value      string
		defaultVal bool
		expected   bool
	}

	var (
		defaultVal bool
	)

	app := iris.New()
	app.Adapt(httprouter.New())
	app.Get("/parse-bool", func(ctx *iris.Context) {
		b := baseController().QueryBool(ctx, "key", defaultVal)
		ctx.JSON(iris.StatusOK, b)
	})

	testcases := []testcase{
		{
			value:    "true",
			expected: true,
		},
		{
			defaultVal: true,
			expected:   true,
		},
		{
			value:      "1",
			defaultVal: false,
			expected:   true,
		},
		{
			value:      "0",
			defaultVal: true,
			expected:   false,
		},
		{
			value:      "false",
			defaultVal: true,
			expected:   false,
		},
		{
			value:      "f",
			defaultVal: true,
			expected:   false,
		},
		{
			value:      "F",
			defaultVal: true,
			expected:   false,
		},
		{
			value:      "-0",
			defaultVal: true,
			expected:   true,
		},
		{
			value:      "true",
			defaultVal: false,
			expected:   true,
		},
		{
			value:      "t",
			defaultVal: false,
			expected:   true,
		},
		{
			value:      "T",
			defaultVal: false,
			expected:   true,
		},
	}

	for _, tc := range testcases {
		defaultVal = tc.defaultVal
		httptest.New(app, t).GET("/parse-bool").WithQuery("key", tc.value).
			Expect().Status(iris.StatusOK).JSON().Boolean().Equal(tc.expected)
	}
}

func TestQueryBools(t *testing.T) {
	type testcase struct {
		values      []string
		defaultVals []bool
		expected    []bool
	}

	var (
		defaultVals []bool
	)

	app := iris.New()
	app.Adapt(httprouter.New())
	app.Get("/parse-bools", func(ctx *iris.Context) {
		bools := baseController().QueryBools(ctx, "key", defaultVals)
		ctx.JSON(iris.StatusOK, bools)
	})

	testcases := []testcase{
		{
			values:   []string{"true", "1", "0", "f", "F", "false", "t", "T"},
			expected: []bool{true, true, false, false, false, false, true, true},
		},
		{
			defaultVals: []bool{true, false},
			expected:    []bool{true, false},
		},
		{
			values:      []string{"string", "2", "-1", "-0"},
			defaultVals: []bool{true},
			expected:    []bool{true},
		},
		{
			values:   []string{"string", "2", "-1"},
			expected: nil,
		},
	}

	for _, tc := range testcases {
		defaultVals = tc.defaultVals
		req := httptest.New(app, t).GET("/parse-bools")
		for _, v := range tc.values {
			req.WithQuery("key", v)
		}
		res := req.Expect().Status(iris.StatusOK).JSON()
		if tc.expected != nil {
			res.Array().Equal(tc.expected)
		} else {
			res.Null()
		}
	}
}

func TestParamString(t *testing.T) {
	type testcase struct {
		queryVal   string
		formVal    string
		defaultVal string
		expected   string
	}

	var (
		defaultVal string
	)

	app := iris.New()
	app.Adapt(httprouter.New())
	app.Post("/param-string", func(ctx *iris.Context) {
		str := baseController().ParamString(ctx, "key", defaultVal)
		ctx.WriteString(str)
	})

	testcases := []testcase{
		{
			queryVal: "without form value",
			expected: "without form value",
		},
		{
			formVal:  "without query value",
			expected: "without query value",
		},
		{
			defaultVal: "default",
			expected:   "default",
		},
		{
			queryVal:   "queryVal",
			formVal:    "formVal",
			defaultVal: "default",
			expected:   "formVal",
		},
	}

	for _, tc := range testcases {
		defaultVal = tc.defaultVal
		req := httptest.New(app, t).POST("/param-string")
		if tc.queryVal != "" {
			req.WithQuery("key", tc.queryVal)
		}
		if tc.formVal != "" {
			req.WithFormField("key", tc.formVal)
		}
		req.Expect().Status(iris.StatusOK).Body().Equal(tc.expected)
	}
}

func TestParamStrings(t *testing.T) {
	type testcase struct {
		queryVals   []string
		formVals    []string
		defaultVals []string
		expected    []string
	}

	var (
		defaultVals []string
	)

	app := iris.New()
	app.Adapt(httprouter.New())
	app.Post("/param-strings", func(ctx *iris.Context) {
		res := baseController().ParamStrings(ctx, "key", defaultVals)
		ctx.JSON(iris.StatusOK, res)
	})

	testcases := []testcase{
		{
			queryVals: []string{"without", "form", "value"},
			expected:  []string{"without", "form", "value"},
		},
		{
			formVals: []string{"without", "query", "value"},
			expected: []string{"without", "query", "value"},
		},
		{
			defaultVals: []string{"default", "values"},
			expected:    []string{"default", "values"},
		},
		{
			queryVals:   []string{"last", "query", "values"},
			formVals:    []string{"first", "form", "values"},
			defaultVals: []string{"default"},
			expected:    []string{"first", "form", "values", "last", "query", "values"},
		},
	}

	for _, tc := range testcases {
		defaultVals = tc.defaultVals
		req := httptest.New(app, t).POST("/param-strings")
		for _, val := range tc.queryVals {
			req.WithQuery("key", val)
		}
		for _, val := range tc.formVals {
			req.WithFormField("key", val)
		}
		res := req.Expect().Status(iris.StatusOK).JSON()
		if tc.expected != nil {
			res.Array().Equal(tc.expected)
		} else {
			res.Null()
		}
	}
}

func TestParamInt(t *testing.T) {
	type testcase struct {
		queryVal   string
		formVal    string
		defaultVal int
		expected   int
	}

	var (
		defaultVal int
	)

	app := iris.New()
	app.Adapt(httprouter.New())
	app.Post("/param-int", func(ctx *iris.Context) {
		i := baseController().ParamInt(ctx, "key", defaultVal)
		ctx.JSON(iris.StatusOK, i)
	})

	testcases := []testcase{
		{
			queryVal: "1",
			expected: 1,
		},
		{
			formVal:  "2",
			expected: 2,
		},
		{
			defaultVal: 3,
			expected:   3,
		},
		{
			queryVal:   "1",
			formVal:    "-2",
			defaultVal: 3,
			expected:   -2,
		},
		{
			queryVal:   "1",
			formVal:    "invalid",
			defaultVal: 3,
			expected:   1,
		},
		{
			queryVal:   "invalid too",
			formVal:    "invalid",
			defaultVal: 3,
			expected:   3,
		},
	}

	for _, tc := range testcases {
		defaultVal = tc.defaultVal
		req := httptest.New(app, t).POST("/param-int")
		if tc.queryVal != "" {
			req.WithQuery("key", tc.queryVal)
		}
		if tc.formVal != "" {
			req.WithFormField("key", tc.formVal)
		}
		req.Expect().Status(iris.StatusOK).JSON().Number().Equal(tc.expected)
	}
}

func TestParamInts(t *testing.T) {
	type testcase struct {
		queryVals   []string
		formVals    []string
		defaultVals []int
		expected    []int
	}

	var (
		defaultVals []int
	)

	app := iris.New()
	app.Adapt(httprouter.New())
	app.Post("/param-ints", func(ctx *iris.Context) {
		res := baseController().ParamInts(ctx, "key", defaultVals)
		ctx.JSON(iris.StatusOK, res)
	})

	testcases := []testcase{
		{
			queryVals: []string{"1", "-1", "3.3"},
			expected:  []int{1, -1},
		},
		{
			formVals:    []string{"valid", "true", "0"},
			defaultVals: []int{1},
			expected:    []int{0},
		},
		{
			defaultVals: []int{1},
			expected:    []int{1},
		},
		{
			queryVals:   []string{"last", "query", "values", "-2"},
			formVals:    []string{"first", "form", "values", "1"},
			defaultVals: []int{3},
			expected:    []int{1, -2},
		},
	}

	for _, tc := range testcases {
		defaultVals = tc.defaultVals
		req := httptest.New(app, t).POST("/param-ints")
		for _, val := range tc.queryVals {
			req.WithQuery("key", val)
		}
		for _, val := range tc.formVals {
			req.WithFormField("key", val)
		}
		res := req.Expect().Status(iris.StatusOK).JSON()
		if tc.expected != nil {
			res.Array().Equal(tc.expected)
		} else {
			res.Null()
		}
	}
}

func TestParamFloat(t *testing.T) {
	type testcase struct {
		queryVal   string
		formVal    string
		defaultVal float64
		expected   float64
	}

	var (
		defaultVal float64
	)

	app := iris.New()
	app.Adapt(httprouter.New())
	app.Post("/param-float", func(ctx *iris.Context) {
		f := baseController().ParamFloat(ctx, "key", defaultVal)
		ctx.JSON(iris.StatusOK, f)
	})

	testcases := []testcase{
		{
			queryVal: "1.1",
			expected: 1.1,
		},
		{
			formVal:  "-2.0",
			expected: -2.0,
		},
		{
			defaultVal: 3.45,
			expected:   3.45,
		},
		{
			queryVal:   "1",
			formVal:    "2",
			defaultVal: 3,
			expected:   2,
		},
		{
			queryVal:   "1",
			formVal:    "invalid",
			defaultVal: 3,
			expected:   1,
		},
		{
			queryVal:   "invalid too",
			formVal:    "invalid",
			defaultVal: 3,
			expected:   3,
		},
	}

	for _, tc := range testcases {
		defaultVal = tc.defaultVal
		req := httptest.New(app, t).POST("/param-float")
		if tc.queryVal != "" {
			req.WithQuery("key", tc.queryVal)
		}
		if tc.formVal != "" {
			req.WithFormField("key", tc.formVal)
		}
		req.Expect().Status(iris.StatusOK).JSON().Number().Equal(tc.expected)
	}
}

func TestParamFloats(t *testing.T) {
	type testcase struct {
		queryVals   []string
		formVals    []string
		defaultVals []float64
		expected    []float64
	}

	var (
		defaultVals []float64
	)

	app := iris.New()
	app.Adapt(httprouter.New())
	app.Post("/param-floats", func(ctx *iris.Context) {
		res := baseController().ParamFloats(ctx, "key", defaultVals)
		ctx.JSON(iris.StatusOK, res)
	})

	testcases := []testcase{
		{
			queryVals: []string{"1.3", "-1.0", "3.3"},
			expected:  []float64{1.3, -1.0, 3.3},
		},
		{
			formVals:    []string{"valid", "true", "0"},
			defaultVals: []float64{1},
			expected:    []float64{0},
		},
		{
			defaultVals: []float64{1.5, -2.0},
			expected:    []float64{1.5, -2.0},
		},
		{
			queryVals:   []string{"last", "query", "values", "-2.5"},
			formVals:    []string{"first", "form", "values", "1"},
			defaultVals: []float64{3.3},
			expected:    []float64{1.0, -2.5},
		},
	}

	for _, tc := range testcases {
		defaultVals = tc.defaultVals
		req := httptest.New(app, t).POST("/param-floats")
		for _, val := range tc.queryVals {
			req.WithQuery("key", val)
		}
		for _, val := range tc.formVals {
			req.WithFormField("key", val)
		}
		res := req.Expect().Status(iris.StatusOK).JSON()
		if tc.expected != nil {
			res.Array().Equal(tc.expected)
		} else {
			res.Null()
		}
	}
}

func TestParamBool(t *testing.T) {
	type testcase struct {
		queryVal   string
		formVal    string
		defaultVal bool
		expected   bool
	}

	var (
		defaultVal bool
	)

	app := iris.New()
	app.Adapt(httprouter.New())
	app.Post("/param-bool", func(ctx *iris.Context) {
		b := baseController().ParamBool(ctx, "key", defaultVal)
		ctx.JSON(iris.StatusOK, b)
	})

	testcases := []testcase{
		{
			queryVal: "true",
			expected: true,
		},
		{
			formVal:  "T",
			expected: true,
		},
		{
			defaultVal: true,
			expected:   true,
		},
		{
			queryVal:   "0",
			formVal:    "1",
			defaultVal: false,
			expected:   true,
		},
		{
			queryVal:   "true",
			formVal:    "invalid",
			defaultVal: false,
			expected:   true,
		},
		{
			queryVal:   "invalid too",
			formVal:    "invalid",
			defaultVal: true,
			expected:   true,
		},
		{
			queryVal:   "-0",
			formVal:    "negative zero is invalid",
			defaultVal: true,
			expected:   true,
		},
	}

	for _, tc := range testcases {
		defaultVal = tc.defaultVal
		req := httptest.New(app, t).POST("/param-bool")
		if tc.queryVal != "" {
			req.WithQuery("key", tc.queryVal)
		}
		if tc.formVal != "" {
			req.WithFormField("key", tc.formVal)
		}
		req.Expect().Status(iris.StatusOK).JSON().Boolean().Equal(tc.expected)
	}
}

func TestParamBools(t *testing.T) {
	type testcase struct {
		queryVals   []string
		formVals    []string
		defaultVals []bool
		expected    []bool
	}

	var (
		defaultVals []bool
	)

	app := iris.New()
	app.Adapt(httprouter.New())
	app.Post("/param-bools", func(ctx *iris.Context) {
		res := baseController().ParamBools(ctx, "key", defaultVals)
		ctx.JSON(iris.StatusOK, res)
	})

	testcases := []testcase{
		{
			queryVals: []string{"1.3", "-1.0", "t", "-0"},
			expected:  []bool{true},
		},
		{
			formVals:    []string{"valid", "true", "F", "1", "3"},
			defaultVals: []bool{true, false, false},
			expected:    []bool{true, false, true},
		},
		{
			defaultVals: []bool{true, false},
			expected:    []bool{true, false},
		},
		{
			queryVals:   []string{"last", "query", "values", "false"},
			formVals:    []string{"first", "form", "values", "1"},
			defaultVals: []bool{false, true},
			expected:    []bool{true, false},
		},
	}

	for _, tc := range testcases {
		defaultVals = tc.defaultVals
		req := httptest.New(app, t).POST("/param-bools")
		for _, val := range tc.queryVals {
			req.WithQuery("key", val)
		}
		for _, val := range tc.formVals {
			req.WithFormField("key", val)
		}
		res := req.Expect().Status(iris.StatusOK).JSON()
		if tc.expected != nil {
			res.Array().Equal(tc.expected)
		} else {
			res.Null()
		}
	}
}

func TestParseQueryPagination(t *testing.T) {
	type testcase struct {
		page     int
		limit    int
		expected string
	}

	defaultSettings := []interface{}{controllers.DefaultPaginationPage, controllers.DefaultPaginationLimit, controllers.QueryKeyPage, controllers.QueryKeyLimit}
	paginationString := func(page, limit, skip int) string {
		return fmt.Sprintf("page=%d,limit=%d,skip=%d", page, limit, skip)
	}

	app := iris.New()
	app.Adapt(httprouter.New())
	app.Get("/pagination", func(ctx *iris.Context) {
		page, skip, limit := baseController().Pagination(ctx)
		ctx.WriteString(paginationString(page, limit, skip))
	})

	t.Run("default pagination setting & query key", func(t *testing.T) {
		testcases := []testcase{
			{
				page:     1,
				limit:    10,
				expected: paginationString(1, 10, 0),
			},
			{
				page:     2,
				limit:    20,
				expected: paginationString(2, 20, 20),
			},
			{
				page:     0,
				limit:    20,
				expected: paginationString(1, 20, 0),
			},
			{
				page:     -1,
				limit:    20,
				expected: paginationString(1, 20, 0),
			},
			{
				page:     2,
				limit:    0,
				expected: paginationString(2, 20, 20),
			},
			{
				page:     2,
				limit:    -1,
				expected: paginationString(2, 20, 20),
			},
			{
				page:     0,
				limit:    0,
				expected: paginationString(1, 20, 0),
			},
		}

		for _, tc := range testcases {
			httptest.New(app, t).GET("/pagination").WithQueryObject(map[string]interface{}{
				"_page":  tc.page,
				"_limit": tc.limit,
			}).Expect().Status(iris.StatusOK).Body().Equal(tc.expected)
		}
	})

	t.Run("change default pagination setting & query key", func(t *testing.T) {
		controllers.DefaultPaginationPage = 2
		controllers.DefaultPaginationLimit = 30
		controllers.QueryKeyPage = "page"
		controllers.QueryKeyLimit = "per_page"

		testcases := []testcase{
			{
				page:     1,
				limit:    10,
				expected: paginationString(1, 10, 0),
			},
			{
				page:     -1,
				limit:    10,
				expected: paginationString(2, 10, 10),
			},
			{
				page:     1,
				limit:    -10,
				expected: paginationString(1, 30, 0),
			},
		}

		for _, tc := range testcases {
			httptest.New(app, t).GET("/pagination").WithQueryObject(map[string]interface{}{
				"page":     tc.page,
				"per_page": tc.limit,
			}).Expect().Status(iris.StatusOK).Body().Equal(tc.expected)
		}
	})

	// tear down
	controllers.DefaultPaginationPage = defaultSettings[0].(int)
	controllers.DefaultPaginationLimit = defaultSettings[1].(int)
	controllers.QueryKeyPage = defaultSettings[2].(string)
	controllers.QueryKeyLimit = defaultSettings[3].(string)
}

func TestParseQuerySort(t *testing.T) {
	type testcase struct {
		sort     string
		order    string
		expected string
	}

	defaultSettings := []string{controllers.OrderAscending, controllers.OrderDescending, controllers.QueryKeySort, controllers.QueryKeyOrder}
	sortString := func(sort, order string) string {
		return fmt.Sprintf("sort=%s,order=%s", sort, order)
	}

	app := iris.New()
	app.Adapt(httprouter.New())
	app.Get("/sort", func(ctx *iris.Context) {
		sort, order := baseController().Sort(ctx)
		ctx.WriteString(sortString(sort, order))
	})

	t.Run("json-server style with default sort settings", func(t *testing.T) {
		testcases := []testcase{
			{
				expected: sortString("", ""),
			},
			{
				sort:     "field",
				order:    "ASC",
				expected: sortString("field", controllers.OrderAscending),
			},
			{
				sort:     "field",
				order:    "DESC",
				expected: sortString("field", controllers.OrderDescending),
			},
			{
				sort:     "field",
				order:    "asc",
				expected: sortString("field", controllers.OrderAscending),
			},
			{
				sort:     "field",
				order:    "asc1",
				expected: sortString("field", controllers.OrderAscending),
			},
		}

		for _, tc := range testcases {
			httptest.New(app, t).GET("/sort").WithQueryObject(map[string]interface{}{
				"_sort":  tc.sort,
				"_order": tc.order,
			}).Expect().Status(iris.StatusOK).Body().Equal(tc.expected)
		}
	})

	t.Run("mgo style with default sort settings", func(t *testing.T) {
		testcases := []testcase{
			{
				sort:     "field",
				expected: sortString("field", controllers.OrderAscending),
			},
			{
				sort:     "-field",
				expected: sortString("field", controllers.OrderDescending),
			},
		}

		for _, tc := range testcases {
			httptest.New(app, t).GET("/sort").WithQueryObject(map[string]interface{}{
				"_sort":  tc.sort,
				"_order": tc.order,
			}).Expect().Status(iris.StatusOK).Body().Equal(tc.expected)
		}
	})

	t.Run("change default sort settings", func(t *testing.T) {

		controllers.OrderAscending = "a"
		controllers.OrderDescending = "d"
		controllers.QueryKeySort = "sort"
		controllers.QueryKeyOrder = "order"

		testcases := []testcase{
			{
				expected: sortString("", ""),
			},
			{
				sort:     "field",
				order:    "a",
				expected: sortString("field", controllers.OrderAscending),
			},
			{
				sort:     "field",
				order:    "d",
				expected: sortString("field", controllers.OrderDescending),
			},
			{
				sort:     "field",
				order:    "A",
				expected: sortString("field", controllers.OrderAscending),
			},
			{
				sort:     "field",
				order:    "asc1",
				expected: sortString("field", controllers.OrderAscending),
			},
			{
				sort:     "-field",
				expected: sortString("field", controllers.OrderDescending),
			},
			{
				sort:     "field",
				expected: sortString("field", controllers.OrderAscending),
			},
		}

		for _, tc := range testcases {
			httptest.New(app, t).GET("/sort").WithQueryObject(map[string]interface{}{
				"sort":  tc.sort,
				"order": tc.order,
			}).Expect().Status(iris.StatusOK).Body().Equal(tc.expected)
		}
	})

	// tear down
	controllers.OrderAscending = defaultSettings[0]
	controllers.OrderDescending = defaultSettings[1]
	controllers.QueryKeySort = defaultSettings[2]
	controllers.QueryKeyOrder = defaultSettings[3]
}

func TestParseQuerySortAsMultiple(t *testing.T) {
	type testcase struct {
		sort     string
		order    string
		expected []string
	}
	app := iris.New()
	app.Adapt(httprouter.New())
	app.Get("/sort-multiple", func(ctx *iris.Context) {
		sorts := baseController().SortAsMultiple(ctx)
		ctx.JSON(iris.StatusOK, sorts)
	})

	testcases := []testcase{
		{
			expected: nil,
		},
		{
			sort:     "field",
			expected: []string{"field"},
		},
		{
			order:    "desc",
			expected: nil,
		},
		{
			sort:     "field",
			order:    "desc",
			expected: []string{"-field"},
		},
		{
			sort:     "field",
			order:    "asc",
			expected: []string{"field"},
		},
		{
			sort:     "field",
			order:    "d",
			expected: []string{"field"},
		},
		{
			sort:     "-field",
			expected: []string{"-field"},
		},
		{
			sort:     "-field,field2",
			order:    "desc",
			expected: []string{"-field", "field2"},
		},
		{
			sort:     "field,-field2,-field3",
			expected: []string{"field", "-field2", "-field3"},
		},
	}

	for _, tc := range testcases {
		res := httptest.New(app, t).GET("/sort-multiple").WithQueryObject(map[string]interface{}{
			"_sort":  tc.sort,
			"_order": tc.order,
		}).Expect().Status(iris.StatusOK).JSON()
		if tc.expected != nil {
			res.Array().Equal(tc.expected)
		} else {
			res.Null()
		}
	}
}

func TestQueryBSON(t *testing.T) {
	type testcase struct {
		q        string
		expected bson.M
	}

	testFn := func(t *testing.T, testcases []testcase) {
		app := iris.New()
		app.Adapt(httprouter.New())
		app.Get("/query-bson/:case", func(ctx *iris.Context) {
			idx, err := ctx.ParamInt("case")
			assert.NoError(t, err)

			tc := testcases[idx]
			actual := baseController().QueryBSON(ctx)
			assert.Equal(t, tc.expected, actual)
		})

		for i, tc := range testcases {
			httptest.New(app, t).GET(fmt.Sprintf("/query-bson/%d", i)).WithQueryString(tc.q).Expect()
		}
	}

	t.Run("normal query", func(t *testing.T) {
		testcases := []testcase{
			{
				q:        "",
				expected: bson.M{},
			},
			{
				q: "name=mary",
				expected: bson.M{
					"name": bson.M{"$eq": "mary"},
				},
			},
			{
				q: "age=10",
				expected: bson.M{
					"age": bson.M{"$eq": 10},
				},
			},
			{
				q: "ok=true",
				expected: bson.M{
					"ok": bson.M{"$eq": true},
				},
			},
			{
				q: "date=2017-04-11T15:31:44",
				expected: bson.M{
					"date": bson.M{"$eq": time.Date(2017, 4, 11, 15, 31, 44, 0, time.UTC)},
				},
			},
			{
				q: "id=58db2700cf2f6715b00021a7",
				expected: bson.M{
					"_id": bson.M{"$eq": bson.ObjectIdHex("58db2700cf2f6715b00021a7")},
				},
			},
			{
				q: "user.id=58db2700cf2f6715b00021a7",
				expected: bson.M{
					"user._id": bson.M{"$eq": bson.ObjectIdHex("58db2700cf2f6715b00021a7")},
				},
			},
		}
		testFn(t, testcases)
	})

	t.Run("query with operator", func(t *testing.T) {
		testcases := []testcase{
			{
				q: "age_gt=18",
				expected: bson.M{
					"age": bson.M{"$gt": 18},
				},
			},
			{
				q: "name_gt=black",
				expected: bson.M{
					"name": bson.M{"$gt": "black"},
				},
			},
			{
				q: "born_at_gt=1990-01-01T12:00:00",
				expected: bson.M{
					"born_at": bson.M{"$gt": time.Date(1990, 1, 1, 12, 0, 0, 0, time.UTC)},
				},
			},
			{
				q: "age_gte=18",
				expected: bson.M{
					"age": bson.M{"$gte": 18},
				},
			},
			{
				q: "name_gte=black",
				expected: bson.M{
					"name": bson.M{"$gte": "black"},
				},
			},
			{
				q: "born_at_gte=1990-01-01T12:00:00",
				expected: bson.M{
					"born_at": bson.M{"$gte": time.Date(1990, 1, 1, 12, 0, 0, 0, time.UTC)},
				},
			},
			{
				q: "age_lt=18",
				expected: bson.M{
					"age": bson.M{"$lt": 18},
				},
			},
			{
				q: "name_lt=black",
				expected: bson.M{
					"name": bson.M{"$lt": "black"},
				},
			},
			{
				q: "born_at_lt=1990-01-01T12:00:00",
				expected: bson.M{
					"born_at": bson.M{"$lt": time.Date(1990, 1, 1, 12, 0, 0, 0, time.UTC)},
				},
			},
			{
				q: "age_lte=18",
				expected: bson.M{
					"age": bson.M{"$lte": 18},
				},
			},
			{
				q: "name_lte=black",
				expected: bson.M{
					"name": bson.M{"$lte": "black"},
				},
			},
			{
				q: "born_at_lte=1990-01-01T12:00:00",
				expected: bson.M{
					"born_at": bson.M{"$lte": time.Date(1990, 1, 1, 12, 0, 0, 0, time.UTC)},
				},
			},
			{
				q: "age_ne=18",
				expected: bson.M{
					"age": bson.M{"$ne": 18},
				},
			},
			{
				q: "name_ne=black",
				expected: bson.M{
					"name": bson.M{"$ne": "black"},
				},
			},
			{
				q: "born_at_ne=1990-01-01T12:00:00",
				expected: bson.M{
					"born_at": bson.M{"$ne": time.Date(1990, 1, 1, 12, 0, 0, 0, time.UTC)},
				},
			},
			{
				q: "ok_ne=True",
				expected: bson.M{
					"ok": bson.M{"$ne": true},
				},
			},
			{
				q: "name_like=black",
				expected: bson.M{
					"name": bson.M{"$regex": bson.RegEx{Pattern: "black"}},
				},
			},
			{
				q: "name_exists=false",
				expected: bson.M{
					"name": bson.M{"$exists": false},
				},
			},
		}
		testFn(t, testcases)
	})

	t.Run("complex query", func(t *testing.T) {
		testcases := []testcase{
			{
				q: "name=jack&name=mary&age_gte=18&age_lte=60&born_at_gt=1990-01-01T00:00:00&email_like=@gmail.com&weight_gt=60.5&parent.id=58db2700cf2f6715b00021a7&parent._id_exists=true&parent.id_ne=58db2700cf2f6715b00021a8",
				expected: bson.M{
					"name":       bson.M{"$in": []interface{}{"jack", "mary"}},
					"age":        bson.M{"$gte": 18, "$lte": 60},
					"weight":     bson.M{"$gt": 60.5},
					"born_at":    bson.M{"$gt": time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC)},
					"parent._id": bson.M{"$exists": true, "$eq": bson.ObjectIdHex("58db2700cf2f6715b00021a7"), "$ne": bson.ObjectIdHex("58db2700cf2f6715b00021a8")},
					"email":      bson.M{"$regex": bson.RegEx{Pattern: "@gmail.com"}},
				},
			},
		}
		testFn(t, testcases)
	})
}
