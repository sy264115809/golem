package controllers

import (
	"time"

	"github.com/sy264115809/logrush"
	iris "gopkg.in/kataras/iris.v6"
)

// LogWithCtx returns *log.Entry with fields converted from context json format body.
func (c *Base) LogWithCtx(ctx *iris.Context) *logrush.Entry {
	var reqCtx interface{}
	if err := ctx.ReadJSON(&reqCtx); err != nil {
		return c.L.WithField("req_ctx", err.Error())
	}
	return c.L.WithField("req_ctx", reqCtx)
}

// LogRunningTime logs the running time from paramter now to the time this function called.
// So this function should always be called as defer form, like:
// defer c.LogRunningTime(time.Now(), message)
func (c *Base) LogRunningTime(from time.Time, message string) {
	c.L.WithField("elapse", time.Now().Sub(from)).Debug(message)
}
