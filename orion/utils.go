package orion

import (
	"github.com/carousell/Orion/orion/handlers"
)

func RegisterEncoder(svr Server, serviceName, method, httpMethod, path string, encoder handlers.Encoder) {
	if e, ok := svr.(handlers.Encodeable); ok {
		e.AddEncoder(serviceName, method, httpMethod, path, encoder)
	}
}
