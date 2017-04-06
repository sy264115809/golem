package recovery

import (
	"runtime/debug"

	iris "gopkg.in/kataras/iris.v6"
)

type (
	// Logger interface
	Logger interface {
		Errorf(format string, v ...interface{})
	}
)

// New restores the server on internal server errors (panics)
// receives an Logger logger to print error stack.
func New(logger Logger) iris.HandlerFunc {
	return func(ctx *iris.Context) {
		defer func() {
			if err := recover(); err != nil {
				logger.Errorf("!!!Recovery from panic\n%s\n%s\n", err, debug.Stack())
				//ctx.Panic just sends http status 500 by default, but you can change it by: iris.OnPanic(func(c *iris.Context){})
				ctx.Panic()
			}
		}()
		ctx.Next()
	}
}
