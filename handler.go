// Copyright 2021 Faisal Alam
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package error_handler

import (
	"bytes"
	"context"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/luraproject/lura/config"
	"github.com/luraproject/lura/proxy"
	luragin "github.com/luraproject/lura/router/gin"
)

func New(next luragin.HandlerFactory) luragin.HandlerFactory {
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

		httpResp, e := httpResponseFromErr(t)
		if e != nil {
			return response, err
		}

		return proxy.NoOpHTTPResponseParser(ctx, httpResp)
	}
}

func httpResponseFromErr(respErr responseError) (*http.Response, error) {
	rawBody, err := respErr.Body()
	if err != nil {
		return nil, err
	}

	httpResp := new(http.Response)
	httpResp.StatusCode = respErr.StatusCode()

	contentType := respErr.ContentType()
	if len(contentType) > 0 {
		httpResp.Header = http.Header{
			"Content-Type": contentType,
		}
	}

	if rawBody != nil {
		httpResp.Body = io.NopCloser(bytes.NewReader(rawBody))
	}

	return httpResp, nil
}

type responseError interface {
	Body() ([]byte, error)
	ContentType() []string
	StatusCode() int
}
