package endpoints

type Request interface {
	validate() error
}

// SquareRequest collects the request parameters for the Square method.
type SquareRequest struct {
	S int64 `json:"s"`
}

func (r SquareRequest) validate() error {
	return nil // TBA
}