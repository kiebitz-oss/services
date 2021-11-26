package metrics

import (
	"errors"
	"net/http"
)

type StatusResponseWriter struct {
	http.ResponseWriter
	status int
}

func MakeStatusResponseWriter(w http.ResponseWriter) *StatusResponseWriter {
	return &StatusResponseWriter{
		w,
		-1,
	}
}

func (r *StatusResponseWriter) Status() (int, error) {

	if r.status != -1 {
		return r.status, nil
	} else {
		return 0, errors.New("status was not set")
	}
}

func (r *StatusResponseWriter) WriteHeader(status int) {
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}
