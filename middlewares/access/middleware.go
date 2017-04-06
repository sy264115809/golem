package access

import (
	"fmt"
	"io"
	"time"

	iris "gopkg.in/kataras/iris.v6"

	uuid "github.com/satori/go.uuid"
)

var (
	green   = string([]byte{27, 91, 57, 55, 59, 52, 50, 109})
	white   = string([]byte{27, 91, 57, 48, 59, 52, 55, 109})
	yellow  = string([]byte{27, 91, 57, 55, 59, 52, 51, 109})
	red     = string([]byte{27, 91, 57, 55, 59, 52, 49, 109})
	blue    = string([]byte{27, 91, 57, 55, 59, 52, 52, 109})
	magenta = string([]byte{27, 91, 57, 55, 59, 52, 53, 109})
	cyan    = string([]byte{27, 91, 57, 55, 59, 52, 54, 109})
	reset   = string([]byte{27, 91, 48, 109})
)

// New instantiates a Logger middleware with the specified writter buffer.
// Example: os.Stdout, a file opened in write mode, a socket...
func New(out io.Writer, prefix string, notlogged ...string) iris.HandlerFunc {
	var skip map[string]struct{}

	if length := len(notlogged); length > 0 {
		skip = make(map[string]struct{}, length)

		for _, path := range notlogged {
			skip[path] = struct{}{}
		}
	}

	return func(ctx *iris.Context) {
		// Start timer
		start := time.Now()
		path := ctx.Request.RequestURI
		reqID := uuid.NewV1().String()

		// Process request
		ctx.Next()

		// Log only when path is not being skipped
		if _, ok := skip[path]; !ok {
			// Stop timer
			end := time.Now()
			latency := end.Sub(start)

			remoteAddr := ctx.RemoteAddr()
			method := ctx.Method()
			statusCode := ctx.StatusCode()
			statusColor := colorForStatus(statusCode)
			methodColor := colorForMethod(method)

			fmt.Fprintf(out, "[%s] %v | %s |%s %3d %s| %13v | %s |%s  %s %-7s %s\n",
				prefix,
				end.Format("2006/01/02 - 15:04:05"),
				reqID,
				statusColor, statusCode, reset,
				latency,
				remoteAddr,
				methodColor, reset, method,
				path,
			)
		}

		// set X-REQID header
		ctx.SetHeader("X-REQID", reqID)
	}
}

func colorForStatus(code int) string {
	switch {
	case code >= 200 && code < 300:
		return green
	case code >= 300 && code < 400:
		return white
	case code >= 400 && code < 500:
		return yellow
	default:
		return red
	}
}

func colorForMethod(method string) string {
	switch method {
	case "GET":
		return blue
	case "POST":
		return cyan
	case "PUT":
		return yellow
	case "DELETE":
		return red
	case "PATCH":
		return green
	case "HEAD":
		return magenta
	case "OPTIONS":
		return white
	default:
		return reset
	}
}
