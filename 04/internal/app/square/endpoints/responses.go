package endpoints

import (
	"net/http"

	httptransport "github.com/go-kit/kit/transport/http"

	"github.com/cage1016/gokit-workshop/internal/app/square/service"
	"github.com/cage1016/gokit-workshop/internal/pkg/responses"
)

var (
	_ httptransport.Headerer = (*SquareResponse)(nil)

	_ httptransport.StatusCoder = (*SquareResponse)(nil)
)

// SquareResponse collects the response values for the Square method.
type SquareResponse struct {
	Res int64 `json:"res"`
	Err error `json:"-"`
}

func (r SquareResponse) StatusCode() int {
	return http.StatusOK // TBA
}

func (r SquareResponse) Headers() http.Header {
	return http.Header{}
}

func (r SquareResponse) Response() interface{} {
	return responses.DataRes{APIVersion: service.Version, Data: r}
}
