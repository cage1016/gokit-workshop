package main

import (
	"context"
	"net/http"
)

// DecodeRequestFunc extracts a user-domain request object from an HTTP
// request object. It's designed to be used in HTTP servers, for server-side
// endpoints. One straightforward DecodeRequestFunc could be something that
// JSON decodes from the request body to the concrete request type.
type DecodeRequestFunc func(context.Context, *http.Request) (request interface{}, err error)

// EncodeResponseFunc encodes the passed response object to the HTTP response
// writer. It's designed to be used in HTTP servers, for server-side
// endpoints. One straightforward EncodeResponseFunc could be something that
// JSON encodes the object directly to the response body.
type EncodeResponseFunc func(context.Context, http.ResponseWriter, interface{}) error

type Server struct {
	e   Endpoint
	dec DecodeRequestFunc
	enc EncodeResponseFunc
}

func NewServer(
	e Endpoint,
	dec DecodeRequestFunc,
	enc EncodeResponseFunc,
) *Server {
	s := &Server{
		e:   e,
		dec: dec,
		enc: enc,
	}

	return s
}

// ServeHTTP implements http.Handler.
func (s Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	request, err := s.dec(ctx, r)
	if err != nil {
		// error handling
		return
	}

	response, err := s.e(ctx, request)
	if err != nil {
		// error handling
		return
	}

	if err := s.enc(ctx, w, response); err != nil {
		// error handling
		return
	}
}
