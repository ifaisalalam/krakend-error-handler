## KrakenD Error Handler

### Description

KrakenD Error Handler allows returning custom response from your KrakenD Plugin.

When you return an error from your Plugin code, KrakenD's default GIN Endpoint handler 
checks if your error implements `StatusCode() int` method and sets the response http status
code to this returned value. There is no way to return a custom body in the response.

By extending KrakenD Error Handler in KrakenD CE, you can return custom response from
your KrakenD plugin if an error occurs.

#### How to add KrakenD Error Handler in KrakenD CE?

In the [`handler_factory.go`](https://github.com/devopsfaith/krakend-ce/blob/9b45e9c3c515f53b624c229f434520d51b6ca456/handler_factory.go) 
file in KrakenD CE, simply add KrakenD Error Handler.

```diff
package krakend

import (
	botdetector "github.com/devopsfaith/krakend-botdetector/gin"
	jose "github.com/devopsfaith/krakend-jose"
	ginjose "github.com/devopsfaith/krakend-jose/gin"
	lua "github.com/devopsfaith/krakend-lua/router/gin"
	metrics "github.com/devopsfaith/krakend-metrics/gin"
	opencensus "github.com/devopsfaith/krakend-opencensus/router/gin"
	juju "github.com/devopsfaith/krakend-ratelimit/juju/router/gin"
	"github.com/luraproject/lura/logging"
	router "github.com/luraproject/lura/router/gin"

+	"github.com/ifaisalalam/krakend-error-handler"
)

// NewHandlerFactory returns a HandlerFactory with a rate-limit and a metrics collector middleware injected
func NewHandlerFactory(logger logging.Logger, metricCollector *metrics.Metrics, rejecter jose.RejecterFactory) router.HandlerFactory {
	handlerFactory := juju.HandlerFactory
+	handlerFactory = error_handler.New(handlerFactory)
	handlerFactory = lua.HandlerFactory(logger, handlerFactory)
	handlerFactory = ginjose.HandlerFactory(handlerFactory, logger, rejecter)
	handlerFactory = metricCollector.NewHTTPHandlerFactory(handlerFactory)
	handlerFactory = opencensus.New(handlerFactory)
	handlerFactory = botdetector.New(handlerFactory, logger)
	return handlerFactory
}
```

### How to use?

To return the custom response to the user from the plugin, the returned error must implement the following
interface.

```go
type responseError interface {
    Body() ([]byte, error)
    ContentType() []string
    StatusCode() int
}
```

### Example request-modifier plugin code

```go
package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

func main() {}

func init() {
	fmt.Println(string(ModifierRegisterer), "loaded!!!")
}

const (
	pluginName = "example-plugin"
)

var ModifierRegisterer = registerer(pluginName)

type registerer string

func (r registerer) RegisterModifiers(f func(
	name string,
	modifierFactory func(map[string]interface{}) func(interface{}) (interface{}, error),
	appliesToRequest bool,
	appliesToResponse bool,
)) {
	f(string(r), r.modifierFactory, true, false)
}

func (r registerer) modifierFactory(map[string]interface{}) func(interface{}) (interface{}, error) {
	return func(input interface{}) (interface{}, error) {
		err := errors.New("something went wrong")
		return nil, &pluginError{err}
	}
}

type pluginError struct {
	err error
}

func (e *pluginError) Error() string {
	if e.err != nil {
		return e.err.Error()
	}
	return ""
}

func (e *pluginError) Unwrap() error {
	return e.err
}

func (e *pluginError) ContentType() []string {
	return []string{"application/json"}
}

func (e *pluginError) Body() ([]byte, error) {
	body := map[string]interface{}{
		"message": e.Error(),
	}

	b := new(bytes.Buffer)
	err := json.NewEncoder(b).Encode(body)

	return b.Bytes(), err
}

func (e *pluginError) StatusCode() int {
	return http.StatusInternalServerError
}
```
