package http

import (
	"net/http"

	"github.com/carousell/Orion/v2/utils/log"
)

type notFoundHandler struct{}

func (n *notFoundHandler) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	writeResp(resp, http.StatusNotFound, []byte("Not found"))
	log.Info(req.Context(), "path", req.URL.String(), "method", req.Method, "error", "404 not found")
}
