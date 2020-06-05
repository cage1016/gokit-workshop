package service

import (
	"context"

	"github.com/go-kit/kit/log"
)

// Middleware describes a service (as opposed to endpoint) middleware.
type Middleware func(SquareService) SquareService

// Service describes a service that adds things together
// Implement yor service methods methods.
// e.x: Foo(ctx context.Context, s string)(rs string, err error)
type SquareService interface {
	// [method=post,expose=true]
	Square(ctx context.Context, s int64) (res int64, err error)
}

// the concrete implementation of service interface
type stubSquareService struct {
	logger log.Logger `json:"logger"`
}

// New return a new instance of the service.
// If you want to add service middleware this is the place to put them.
func New(logger log.Logger) (s SquareService) {
	var svc SquareService
	{
		svc = &stubSquareService{logger: logger}
		svc = LoggingMiddleware(logger)(svc)
	}
	return svc
}

// Implement the business logic of Square
func (sq *stubSquareService) Square(ctx context.Context, s int64) (res int64, err error) {
	return s * s, err
}
