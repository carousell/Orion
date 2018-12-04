package http

import (
	"net/http"

	"github.com/carousell/Orion/utils/log"
)

type notFoundHandler struct{}

func (n *notFoundHandler) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	writeResp(resp, http.StatusNotFound, []byte("Not Found: "+req.URL.String()))
	log.Info(req.Context(), "path", req.URL.String(), "method", req.Method, "error", "404 not found")
}
