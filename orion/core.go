package orion

import (
	"fmt"
	"strings"

	"google.golang.org/grpc"
)

func generateUrl(serviceName, method string) string {
	parts := strings.Split(serviceName, ".")
	if len(parts) > 1 {
		serviceName = strings.ToLower(parts[1])
	}
	method = strings.ToLower(method)
	return "/" + serviceName + "/" + method
}

func RegisterService(sd *grpc.ServiceDesc, ss interface{}, svr Server) {
	fmt.Println("Processing Service: ", sd.ServiceName)
	fmt.Println("Mapped URLs: ")
	for _, m := range sd.Methods {
		fmt.Println("\tPOST ", generateUrl(sd.ServiceName, m.MethodName))
	}
}

type svr struct {
}

func GetServer() Server {
	return &svr{}
}
