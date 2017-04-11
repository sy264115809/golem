package controllers

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/fatih/structs"

	iris "gopkg.in/kataras/iris.v6"
)

type (
	// Response can make a response with given datas and status.
	Response struct {
		ctx  *iris.Context
		data iris.Map
		code Code
	}

	// Code could be a user defined number with specific meaning.
	Code interface {
		Code() int
		Humanize() string
	}
)

// WithData sets the given key-value pair to the response json body.
// It will overwrites the value with same key.
func (r *Response) WithData(key string, value interface{}) *Response {
	r.data[key] = value
	return r
}

// WithDatas sets multiple key-value pairs to the response json body.
// All values has the same key as the `datas` has will be overwrited.
func (r *Response) WithDatas(datas iris.Map) *Response {
	for k, v := range datas {
		r.data[k] = v
	}
	return r
}

// WithStruct converts the struct `v` to map and does the same things as what `WithDatas` does.
func (r *Response) WithStruct(v interface{}) *Response {
	if structs.IsStruct(v) {
		s := structs.New(v)
		s.TagName = "json"
		s.FillMap(r.data)
	}
	return r
}

// WithCode sets biz code of the response.
func (r *Response) WithCode(code Code) *Response {
	r.code = code
	return r
}

// Ok render json format response with iris.OK and given message.
func (r *Response) Ok(message ...string) {
	r.json(iris.StatusOK, message...)
}

// Created render json format response with iris.Created and given message.
func (r *Response) Created(message ...string) {
	r.json(iris.StatusCreated, message...)
}

// BadRequest render json format response with iris.StatusBadRequest and given message.
func (r *Response) BadRequest(message ...string) {
	r.json(iris.StatusBadRequest, message...)
}

// Unauthorized render json format response with iris.StatusUnauthorized and given message.
func (r *Response) Unauthorized(message ...string) {
	r.json(iris.StatusUnauthorized, message...)
}

// Forbidden render json format response with iris.StatusForbidden and given message.
func (r *Response) Forbidden(message ...string) {
	r.json(iris.StatusForbidden, message...)
}

// NotFound render json format response with iris.StatusNotFound and given message.
func (r *Response) NotFound(message ...string) {
	r.json(iris.StatusNotFound, message...)
}

// Conflict render json format response with iris.StatusConflict and given message.
func (r *Response) Conflict(message ...string) {
	r.json(iris.StatusConflict, message...)
}

// InternalServerError render json format response with iris.InternalServerError and given message.
func (r *Response) InternalServerError(message ...string) {
	r.json(iris.StatusInternalServerError, message...)
}

// CustomCode render json format response with given http code and message.
func (r *Response) CustomCode(code int, message ...string) {
	r.json(code, message...)
}

func (r *Response) json(code int, message ...string) {
	if len(message) > 0 {
		r.data["message"] = message[0]
	}
	r.parseCode()
	r.ctx.JSON(code, r.data)
}

func (r *Response) parseCode() {
	if r.code != nil {
		r.data["code"] = r.code.Code()
		r.data["code_text"] = r.code.Humanize()
	}
}

// Redirect redirects the client to the specific url using 302 http code.
// The data attached will be encoded as url parameters.
func (r *Response) Redirect(urlRedirect string) {
	r.parseCode()
	if len(r.data) > 0 {
		if u, err := url.Parse(urlRedirect); err == nil {
			q := u.Query()

			for k, v := range r.data {
				q.Add(k, r.toString(v))
			}

			u.RawQuery = q.Encode()
			urlRedirect = u.String()
		}
	}
	r.ctx.Redirect(urlRedirect)
}

func (r *Response) toString(v interface{}) string {
	if s, ok := v.(string); ok {
		return s
	}

	byts, err := json.Marshal(v)
	if err == nil && string(byts) != "{}" {
		return string(byts)
	}

	return fmt.Sprintf("%v", v)
}

// Response makes a response instance by context ctx provided.
func (c *Base) Response(ctx *iris.Context) *Response {
	return &Response{
		ctx:  ctx,
		data: make(iris.Map),
	}
}
