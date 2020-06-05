package endpoints

import (
	"context"

	"github.com/cage1016/gokit-workshop/internal/app/square/service"
	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/tracing/opentracing"
	"github.com/go-kit/kit/tracing/zipkin"
	stdopentracing "github.com/opentracing/opentracing-go"
	stdzipkin "github.com/openzipkin/zipkin-go"
)

// Endpoints collects all of the endpoints that compose the square service. It's
// meant to be used as a helper struct, to collect all of the endpoints into a
// single parameter.
type Endpoints struct {
	SquareEndpoint endpoint.Endpoint `json:""`
}

// New return a new instance of the endpoint that wraps the provided service.
func New(svc service.SquareService, logger log.Logger, otTracer stdopentracing.Tracer, zipkinTracer *stdzipkin.Tracer) (ep Endpoints) {
	var squareEndpoint endpoint.Endpoint
	{
		method := "square"
		squareEndpoint = MakeSquareEndpoint(svc)
		squareEndpoint = opentracing.TraceServer(otTracer, method)(squareEndpoint)
		squareEndpoint = zipkin.TraceEndpoint(zipkinTracer, method)(squareEndpoint)
		squareEndpoint = LoggingMiddleware(log.With(logger, "method", method))(squareEndpoint)
		ep.SquareEndpoint = squareEndpoint
	}

	return ep
}

// MakeSquareEndpoint returns an endpoint that invokes Square on the service.
// Primarily useful in a server.
func MakeSquareEndpoint(svc service.SquareService) (ep endpoint.Endpoint) {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(SquareRequest)
		if err := req.validate(); err != nil {
			return SquareResponse{}, err
		}
		res, err := svc.Square(ctx, req.S)
		return SquareResponse{Res: res}, err
	}
}

// Square implements the service interface, so Endpoints may be used as a service.
// This is primarily useful in the context of a client library.
func (e Endpoints) Square(ctx context.Context, s int64) (res int64, err error) {
	resp, err := e.SquareEndpoint(ctx, SquareRequest{S: s})
	if err != nil {
		return
	}
	response := resp.(SquareResponse)
	return response.Res, nil
}
