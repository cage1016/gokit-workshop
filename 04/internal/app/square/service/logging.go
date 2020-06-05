package service

import (
	"context"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
)

type loggingMiddleware struct {
	logger log.Logger    `json:""`
	next   SquareService `json:""`
}

// LoggingMiddleware takes a logger as a dependency
// and returns a ServiceMiddleware.
func LoggingMiddleware(logger log.Logger) Middleware {
	return func(next SquareService) SquareService {
		return loggingMiddleware{level.Info(logger), next}
	}
}

func (lm loggingMiddleware) Square(ctx context.Context, s int64) (res int64, err error) {
	defer func() {
		lm.logger.Log("method", "Square", "s", s, "err", err)
	}()

	return lm.next.Square(ctx, s)
}
