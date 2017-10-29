package utils

import (
	"log"
	"net/http"
	"net/http/pprof"
	"strings"
)

func GetPprofHandler(base string) http.Handler {
	h := func(w http.ResponseWriter, r *http.Request) {
		name := strings.TrimPrefix(r.URL.Path, base)
		log.Println("name is ", name)
		switch name {
		case "":
			pprof.Index(w, r)
		case "cmdline":
			pprof.Cmdline(w, r)
		case "profile":
			pprof.Profile(w, r)
		case "trace":
			pprof.Trace(w, r)
		case "symbol":
			pprof.Symbol(w, r)
		default:
			// Provides access to all profiles under runtime/pprof
			pprof.Handler(name).ServeHTTP(w, r)
		}
	}
	return http.HandlerFunc(h)
}
