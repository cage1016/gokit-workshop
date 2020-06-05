package transports

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	kitjwt "github.com/go-kit/kit/auth/jwt"
	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/tracing/opentracing"
	"github.com/go-kit/kit/tracing/zipkin"
	httptransport "github.com/go-kit/kit/transport/http"
	"github.com/go-zoo/bone"
	stdopentracing "github.com/opentracing/opentracing-go"
	stdzipkin "github.com/openzipkin/zipkin-go"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/grpc/status"

	"github.com/cage1016/gokit-workshop/internal/app/square/endpoints"
	"github.com/cage1016/gokit-workshop/internal/app/square/service"
	"github.com/cage1016/gokit-workshop/internal/pkg/errors"
	"github.com/cage1016/gokit-workshop/internal/pkg/responses"
)

const (
	contentType string = "application/json"
)

func JSONErrorDecoder(r *http.Response) error {
	contentType := r.Header.Get("Content-Type")
	if !strings.Contains(contentType, "application/json") {
		return fmt.Errorf("expected JSON formatted error, got Content-Type %s", contentType)
	}
	var w responses.ErrorWrapper
	if err := json.NewDecoder(r.Body).Decode(&w); err != nil {
		return err
	}
	return errors.New(w.Error)
}

// NewHTTPHandler returns a handler that makes a set of endpoints available on
// predefined paths.
func NewHTTPHandler(endpoints endpoints.Endpoints, otTracer stdopentracing.Tracer, zipkinTracer *stdzipkin.Tracer, logger log.Logger) http.Handler { // Zipkin HTTP Server Trace can either be instantiated per endpoint with a
	// provided operation name or a global tracing service can be instantiated
	// without an operation name and fed to each Go kit endpoint as ServerOption.
	// In the latter case, the operation name will be the endpoint's http method.
	// We demonstrate a global tracing service here.
	zipkinServer := zipkin.HTTPServerTrace(zipkinTracer)

	options := []httptransport.ServerOption{
		httptransport.ServerErrorEncoder(httpEncodeError),
		httptransport.ServerErrorLogger(logger),
		zipkinServer,
	}

	m := bone.New()
	m.Post("/square", httptransport.NewServer(
		endpoints.SquareEndpoint,
		decodeHTTPSquareRequest,
		encodeJSONResponse,
		append(options, httptransport.ServerBefore(opentracing.HTTPToContext(otTracer, "Square", logger), kitjwt.HTTPToContext()))...,
	))
	m.Get("/metrics", promhttp.Handler())
	return m
}

// decodeHTTPSquareRequest is a transport/http.DecodeRequestFunc that decodes a
// JSON-encoded request from the HTTP request body. Primarily useful in a server.
func decodeHTTPSquareRequest(_ context.Context, r *http.Request) (interface{}, error) {
	var req endpoints.SquareRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	return req, err
}

// NewHTTPClient returns an AddService backed by an HTTP server living at the
// remote instance. We expect instance to come from a service discovery system,
// so likely of the form "host:port". We bake-in certain middlewares,
// implementing the client library pattern.
func NewHTTPClient(instance string, otTracer stdopentracing.Tracer, zipkinTracer *stdzipkin.Tracer, logger log.Logger) (service.SquareService, error) { // Quickly sanitize the instance string.
	if !strings.HasPrefix(instance, "http") {
		instance = "http://" + instance
	}
	u, err := url.Parse(instance)
	if err != nil {
		return nil, err
	}

	// Zipkin HTTP Client Trace can either be instantiated per endpoint with a
	// provided operation name or a global tracing client can be instantiated
	// without an operation name and fed to each Go kit endpoint as ClientOption.
	// In the latter case, the operation name will be the endpoint's http method.
	zipkinClient := zipkin.HTTPClientTrace(zipkinTracer)

	// global client middlewares
	options := []httptransport.ClientOption{
		zipkinClient,
	}

	e := endpoints.Endpoints{}

	// Each individual endpoint is an http/transport.Client (which implements
	// endpoint.Endpoint) that gets wrapped with various middlewares. If you
	// made your own client library, you'd do this work there, so your server
	// could rely on a consistent set of client behavior.
	// The Square endpoint is the same thing, with slightly different
	// middlewares to demonstrate how to specialize per-endpoint.
	var squareEndpoint endpoint.Endpoint
	{
		squareEndpoint = httptransport.NewClient(
			"POST",
			copyURL(u, "/square"),
			encodeHTTPSquareRequest,
			decodeHTTPSquareResponse,
			append(options, httptransport.ClientBefore(opentracing.ContextToHTTP(otTracer, logger)))...,
		).Endpoint()
		squareEndpoint = opentracing.TraceClient(otTracer, "Square")(squareEndpoint)
		squareEndpoint = zipkin.TraceEndpoint(zipkinTracer, "Square")(squareEndpoint)
		e.SquareEndpoint = squareEndpoint
	}

	// Returning the endpoint.Set as a service.Service relies on the
	// endpoint.Set implementing the Service methods. That's just a simple bit
	// of glue code.
	return e, nil
}

//
func copyURL(base *url.URL, path string) *url.URL {
	next := *base
	next.Path = path
	return &next
}

// encodeHTTPSquareRequest is a transport/http.EncodeRequestFunc that
// JSON-encodes any request to the request body. Primarily useful in a client.
func encodeHTTPSquareRequest(_ context.Context, r *http.Request, request interface{}) (err error) {
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(request); err != nil {
		return err
	}
	r.Body = ioutil.NopCloser(&buf)
	return nil
}

// decodeHTTPSquareResponse is a transport/http.DecodeResponseFunc that decodes a
// JSON-encoded sum response from the HTTP response body. If the response has a
// non-200 status code, we will interpret that as an error and attempt to decode
// the specific error message from the response body. Primarily useful in a client.
func decodeHTTPSquareResponse(_ context.Context, r *http.Response) (interface{}, error) {
	if r.StatusCode != http.StatusOK {
		return nil, JSONErrorDecoder(r)
	}
	var resp endpoints.SquareResponse
	err := json.NewDecoder(r.Body).Decode(&resp)
	return resp, err
}

func httpEncodeError(_ context.Context, err error, w http.ResponseWriter) {
	code := http.StatusInternalServerError
	var message string
	var errs []errors.Errors
	w.Header().Set("Content-Type", contentType)
	if s, ok := status.FromError(err); !ok {
		// HTTP
		switch errorVal := err.(type) {
		case errors.Error:
			switch {
			// TODO write your own custom error check here
			}

			if errorVal.Msg() != "" {
				message, errs = errorVal.Msg(), errorVal.Errors()
			}
		default:
			switch err {
			case io.ErrUnexpectedEOF, io.EOF:
				code = http.StatusBadRequest
			case kitjwt.ErrTokenContextMissing:
				code = http.StatusUnauthorized
			default:
				switch err.(type) {
				case *json.SyntaxError, *json.UnmarshalTypeError:
					code = http.StatusBadRequest
				}
			}

			errs = errors.FromError(err.Error())
			message = errs[0].Message
		}
	} else {
		// GRPC
		code = HTTPStatusFromCode(s.Code())
		errs = errors.FromError(s.Message())
		message = errs[0].Message
	}

	w.WriteHeader(code)
	json.NewEncoder(w).Encode(responses.ErrorRes{responses.ErrorResItem{code, message, errs}})
}

func encodeJSONResponse(_ context.Context, w http.ResponseWriter, response interface{}) error {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	if headerer, ok := response.(httptransport.Headerer); ok {
		for k, values := range headerer.Headers() {
			for _, v := range values {
				w.Header().Add(k, v)
			}
		}
	}
	code := http.StatusOK
	if sc, ok := response.(httptransport.StatusCoder); ok {
		code = sc.StatusCode()
	}
	w.WriteHeader(code)
	if code == http.StatusNoContent {
		return nil
	}

	if ar, ok := response.(responses.Responser); ok {
		return json.NewEncoder(w).Encode(ar.Response())
	}

	return json.NewEncoder(w).Encode(response)
}
