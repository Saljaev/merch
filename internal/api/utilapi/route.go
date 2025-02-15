package utilapi

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
)

var (
	ErrNoSupported = errors.New("method not supported")
	ErrTimeout     = errors.New("context timeout")
)

type HandlerFunc func(ctx *APIContext)

func (r *Router) Handle(pattern, method string, handlerFuncs ...HandlerFunc) {
	r.mux.HandleFunc(pattern, func(w http.ResponseWriter, req *http.Request) {
		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "Streaming not supported", http.StatusInternalServerError)
			return
		}

		done := make(chan struct{})

		go func() {
			defer close(done)

			ctx := newAPIContext(w, req, r.log, r.sli)
			ctx.log = ctx.log.With(slog.String("pattern", fmt.Sprintf("%s", pattern)))
			ctx.w.Header().Set("Content-Type", "application/json; charset=utf-8")

			if req.Method != method {
				ctx.Error("unsupported method", ErrNoSupported)
				ctx.WriteFailure(http.StatusMethodNotAllowed, "internal error")
				return
			}

			for _, h := range handlerFuncs {
				select {
				case <-ctx.Done():
					ctx.Error("context timeout", ErrTimeout)
					ctx.WriteFailure(http.StatusInternalServerError, "internal error")
					return
				default:
					h(ctx)
				}
			}

			flusher.Flush()
		}()

		<-done
	})
}
