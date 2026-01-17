package logging

import (
	"log/slog"
	"net/http"
)

type loggingWriter struct {
	http.ResponseWriter
	request    *http.Request
	statusCode int
	size       int
}

func (w *loggingWriter) WriteHeader(statusCode int) {
	// NOTE: Prevents accidental rewrites
	if w.statusCode == 0 {
		w.statusCode = statusCode
	}
	w.ResponseWriter.WriteHeader(statusCode)
}

func (w *loggingWriter) Write(b []byte) (int, error) {
	// NOTE: Edge case: WriteHeader not called. i.e. implicit 200 OK
	if w.statusCode == 0 {
		w.statusCode = http.StatusOK
	}

	n, err := w.ResponseWriter.Write(b)
	if err != nil {
		logger := GetLogger(w.request)
		logger.Error("Failed to write response body", slog.String("err", err.Error()))
	}

	w.size += n

	return n, err
}
