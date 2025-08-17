package util

import "fmt"

type HttpError struct {
	StatusCode int
	Msg        string
}

func (e *HttpError) Error() string {
	return fmt.Sprintf("riot API error: %d - %s", e.StatusCode, e.Msg)
}