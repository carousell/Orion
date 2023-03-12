package http

import (
	"github.com/carousell/logging"
	"net/http"
)

type notFoundHandler struct{}

func (n *notFoundHandler) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	writeResp(resp, http.StatusNotFound, []byte("Not found"))
	logging.Info(req.Context(), "path", req.URL.String(), "method", req.Method, "error", "404 not found")
}
