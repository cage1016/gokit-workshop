package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
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

func (s *addService) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		var req SumRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if res, err := s.sum(req.A, req.B); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		} else {
			fmt.Fprint(w, res)
		}
	}
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
		log.Printf("Defaulting to port %s", port)
	}

	log.Printf("Listening on port %s", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", port), &addService{}))
}
