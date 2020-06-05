package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
)

type Addsvc interface {
	sum(a, b int64) (int64, error)
}

type SumRequest struct {
	A int64 `json:"a"`
	B int64 `json:"b"`
}

type addService struct{}

func (s *addService) sum(a, b int64) (int64, error) {
	return a + b, nil
}

type Endpoint func(ctx context.Context, request interface{}) (response interface{}, err error)

func MakePostSumEndpoint(s addService) Endpoint {
	return func(_ context.Context, request interface{}) (response interface{}, err error) {
		p := request.(SumRequest)
		return s.sum(p.A, p.B)
	}
}

type Middleware func(Endpoint) Endpoint

func loggingMiddleware(next Endpoint) Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		begin := time.Now()
		defer func() {
			log.Printf("request took %s", time.Since(begin))
		}()
		return next(ctx, request)
	}
}

func decodeHTTPSquareRequest(_ context.Context, r *http.Request) (interface{}, error) {
	var req SumRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	return req, err
}

func encodeJSONResponse(_ context.Context, w http.ResponseWriter, response interface{}) error {
	return json.NewEncoder(w).Encode(response)
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
		log.Printf("Defaulting to port %s", port)
	}

	svc := addService{}
	e := MakePostSumEndpoint(svc)
	e = loggingMiddleware(e)

	p := NewServer(e, decodeHTTPSquareRequest, encodeJSONResponse)
	r := mux.NewRouter()
	r.Methods(http.MethodPost).Handler(p)

	http.ListenAndServe(fmt.Sprintf(":%s", port), r)
}
