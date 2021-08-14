package error_handler

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/luraproject/lura/config"
	"github.com/luraproject/lura/proxy"
	luragin "github.com/luraproject/lura/router/gin"
)

func NewHandler(next luragin.HandlerFactory) luragin.HandlerFactory {
	return func(config *config.EndpointConfig, p proxy.Proxy) gin.HandlerFunc {
		return next(config, proxyHandler(p))
	}
}

func proxyHandler(next proxy.Proxy) proxy.Proxy {
	return func(ctx context.Context, request *proxy.Request) (*proxy.Response, error) {
		response, err := next(ctx, request)
		if err == nil {
			return response, err
		}

		t, ok := err.(responseError)
		if !ok {
			return response, err
		}

		if c, ok := ctx.Value(&gin.CtxKey).(*gin.Context); ok {
			c.Render(t.StatusCode(), renderer{err: t})
		}

		return response, err
	}
}

type renderer struct {
	err responseError
}

func (r renderer) Render(w http.ResponseWriter) error {
	r.WriteContentType(w)

	body, e := r.err.Body()
	if e != nil {
		return e
	}

	_, e = w.Write(body)
	return e
}

func (r renderer) WriteContentType(w http.ResponseWriter) {
	header := w.Header()
	if val := header["Content-Type"]; len(val) == 0 {
		header["Content-Type"] = r.err.ContentType()
	}
}

type responseError interface {
	Body() ([]byte, error)
	ContentType() []string
	StatusCode() int
}
