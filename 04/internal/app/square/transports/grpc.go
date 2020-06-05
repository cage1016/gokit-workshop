package transports

import (
	"context"

	"github.com/cage1016/gokit-workshop/internal/app/square/endpoints"
	"github.com/cage1016/gokit-workshop/internal/app/square/service"
	"github.com/cage1016/gokit-workshop/internal/pkg/errors"
	pb "github.com/cage1016/gokit-workshop/pb/square"
	kitjwt "github.com/go-kit/kit/auth/jwt"
	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/tracing/opentracing"
	"github.com/go-kit/kit/tracing/zipkin"
	grpctransport "github.com/go-kit/kit/transport/grpc"
	stdopentracing "github.com/opentracing/opentracing-go"
	stdzipkin "github.com/openzipkin/zipkin-go"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type grpcServer struct {
	square grpctransport.Handler `json:""`
}

func (s *grpcServer) Square(ctx context.Context, req *pb.SquareRequest) (rep *pb.SquareResponse, err error) {
	_, rp, err := s.square.ServeGRPC(ctx, req)
	if err != nil {
		return nil, grpcEncodeError(errors.Cast(err))
	}
	rep = rp.(*pb.SquareResponse)
	return rep, nil
}

// MakeGRPCServer makes a set of endpoints available as a gRPC server.
func MakeGRPCServer(endpoints endpoints.Endpoints, otTracer stdopentracing.Tracer, zipkinTracer *stdzipkin.Tracer, logger log.Logger) (req pb.SquareServer) { // Zipkin GRPC Server Trace can either be instantiated per gRPC method with a
	// provided operation name or a global tracing service can be instantiated
	// without an operation name and fed to each Go kit gRPC server as a
	// ServerOption.
	// In the latter case, the operation name will be the endpoint's grpc method
	// path if used in combination with the Go kit gRPC Interceptor.
	//
	// In this example, we demonstrate a global Zipkin tracing service with
	// Go kit gRPC Interceptor.
	zipkinServer := zipkin.GRPCServerTrace(zipkinTracer)

	options := []grpctransport.ServerOption{
		grpctransport.ServerErrorLogger(logger),
		zipkinServer,
	}

	return &grpcServer{
		square: grpctransport.NewServer(
			endpoints.SquareEndpoint,
			decodeGRPCSquareRequest,
			encodeGRPCSquareResponse,
			append(options, grpctransport.ServerBefore(opentracing.GRPCToContext(otTracer, "Square", logger), kitjwt.GRPCToContext()))...,
		),
	}
}

// decodeGRPCSquareRequest is a transport/grpc.DecodeRequestFunc that converts a
// gRPC request to a user-domain request. Primarily useful in a server.
func decodeGRPCSquareRequest(_ context.Context, grpcReq interface{}) (interface{}, error) {
	req := grpcReq.(*pb.SquareRequest)
	return endpoints.SquareRequest{S: req.S}, nil
}

// encodeGRPCSquareResponse is a transport/grpc.EncodeResponseFunc that converts a
// user-domain response to a gRPC reply. Primarily useful in a server.
func encodeGRPCSquareResponse(_ context.Context, grpcReply interface{}) (res interface{}, err error) {
	reply := grpcReply.(endpoints.SquareResponse)
	return &pb.SquareResponse{Res: reply.Res}, grpcEncodeError(errors.Cast(reply.Err))
}

// NewGRPCClient returns an AddService backed by a gRPC server at the other end
// of the conn. The caller is responsible for constructing the conn, and
// eventually closing the underlying transport. We bake-in certain middlewares,
// implementing the client library pattern.
func NewGRPCClient(conn *grpc.ClientConn, otTracer stdopentracing.Tracer, zipkinTracer *stdzipkin.Tracer, logger log.Logger) service.SquareService { // Zipkin GRPC Client Trace can either be instantiated per gRPC method with a
	// provided operation name or a global tracing client can be instantiated
	// without an operation name and fed to each Go kit client as ClientOption.
	// In the latter case, the operation name will be the endpoint's grpc method
	// path.
	//
	// In this example, we demonstrace a global tracing client.
	zipkinClient := zipkin.GRPCClientTrace(zipkinTracer)

	// global client middlewares
	options := []grpctransport.ClientOption{
		zipkinClient,
	}

	// The Square endpoint is the same thing, with slightly different
	// middlewares to demonstrate how to specialize per-endpoint.
	var squareEndpoint endpoint.Endpoint
	{
		squareEndpoint = grpctransport.NewClient(
			conn,
			"pb.Square",
			"Square",
			encodeGRPCSquareRequest,
			decodeGRPCSquareResponse,
			pb.SquareResponse{},
			append(options, grpctransport.ClientBefore(opentracing.ContextToGRPC(otTracer, logger), kitjwt.ContextToGRPC()))...,
		).Endpoint()
		squareEndpoint = opentracing.TraceClient(otTracer, "Square")(squareEndpoint)
	}

	return endpoints.Endpoints{
		SquareEndpoint: squareEndpoint,
	}
}

// encodeGRPCSquareRequest is a transport/grpc.EncodeRequestFunc that converts a
// user-domain Square request to a gRPC Square request. Primarily useful in a client.
func encodeGRPCSquareRequest(_ context.Context, request interface{}) (interface{}, error) {
	req := request.(endpoints.SquareRequest)
	return &pb.SquareRequest{S: req.S}, nil
}

// decodeGRPCSquareResponse is a transport/grpc.DecodeResponseFunc that converts a
// gRPC Square reply to a user-domain Square response. Primarily useful in a client.
func decodeGRPCSquareResponse(_ context.Context, grpcReply interface{}) (interface{}, error) {
	reply := grpcReply.(*pb.SquareResponse)
	return endpoints.SquareResponse{Res: reply.Res}, nil
}

func grpcEncodeError(err errors.Error) error {
	if err == nil {
		return nil
	}

	st, ok := status.FromError(err)
	if ok {
		return status.Error(st.Code(), st.Message())
	}

	switch {
	// TODO write your own custom error check here
	case errors.Contains(err, kitjwt.ErrTokenContextMissing):
		return status.Error(codes.Unauthenticated, err.Error())
	default:
		return status.Error(codes.Internal, "internal server error")
	}
}
