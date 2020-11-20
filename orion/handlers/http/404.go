package http

import (
	"fmt"
	"github.com/carousell/Orion/utils/log/loggers"
	"net/http"

	"github.com/carousell/Orion/utils/log"
)

type notFoundHandler struct{}

func (n *notFoundHandler) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	writeResp(resp, http.StatusNotFound, []byte("Not Found: "+req.URL.String()))
	log.Info(req.Context(), fmt.Sprintf("%s: 404 not found", req.URL.String()),
		[]loggers.Label{{"path", req.URL.String()}, {"method", req.Method}, {"error", "404 not found"}})
}
