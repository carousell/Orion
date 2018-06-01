package handlers

import (
	"context"
	"log"
	"reflect"

	"github.com/carousell/Orion/interceptors"
	"github.com/carousell/Orion/orion/modifiers"
	"github.com/carousell/Orion/utils/errors"
	"github.com/carousell/Orion/utils/errors/notifier"
	"github.com/carousell/Orion/utils/options"
	"google.golang.org/grpc"
)

// ChainUnaryServer creates a single interceptor out of a chain of many interceptors.
//
// Execution is done in left-to-right order, including passing of context.
// For example ChainUnaryServer(one, two, three) will execute one before two before three, and three
// will see context changes of one and two.
func chainUnaryServer(interceptors ...grpc.UnaryServerInterceptor) grpc.UnaryServerInterceptor {
	n := len(interceptors)

	if n > 1 {
		lastI := n - 1
		return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
			var (
				chainHandler grpc.UnaryHandler
				curI         int
			)

			chainHandler = func(currentCtx context.Context, currentReq interface{}) (interface{}, error) {
				if curI == lastI {
					return handler(currentCtx, currentReq)
				}
				curI++
				return interceptors[curI](currentCtx, currentReq, info, chainHandler)
			}

			return interceptors[0](ctx, req, info, chainHandler)
		}
	}

	if n == 1 {
		return interceptors[0]
	}

	// n == 0; Dummy interceptor maintained for backward compatibility to avoid returning nil.
	return func(ctx context.Context, req interface{}, _ *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		return handler(ctx, req)
	}
}

//GetInterceptors fetches interceptors from a given GRPC service
func GetInterceptors(svc interface{}, config CommonConfig) grpc.UnaryServerInterceptor {
	return chainUnaryServer(getInterceptors(svc, config, []string{})...)
}

func GetInterceptorsWithMethodMiddlewares(svc interface{}, config CommonConfig, middlewares []string) grpc.UnaryServerInterceptor {
	return chainUnaryServer(getInterceptors(svc, config, middlewares)...)
}

func getInterceptors(svc interface{}, config CommonConfig, middlewares []string) []grpc.UnaryServerInterceptor {
	opts := []grpc.UnaryServerInterceptor{optionsInterceptor}

	// check and add default interceptors
	if !config.NoDefaultInterceptors {
		// Add default interceptors
		opts = append(opts, interceptors.DefaultInterceptors()...)
	}

	// check and add service interceptors
	interceptor, ok := svc.(Interceptor)
	if ok {
		opts = append(opts, interceptor.GetInterceptors()...)
	}

	// check and add method interceptors
	opts = append(opts, GetMethodInterceptors(svc, config, middlewares)...)

	return opts
}

func getMiddleware(svc interface{}, middleware string) (grpc.UnaryServerInterceptor, error) {
	r := reflect.TypeOf(svc)
	if m, ok := r.MethodByName(middleware); ok {
		if m.Type.NumIn() == 1 && m.Type.NumOut() == 1 && !m.Type.IsVariadic() {
			t := reflect.TypeOf(grpc.UnaryServerInterceptor(nil))
			if r.ConvertibleTo(m.Type.In(0)) && m.Type.Out(0).ConvertibleTo(t) {
				v := m.Func.Call([]reflect.Value{reflect.ValueOf(svc)})
				return v[0].Interface().(grpc.UnaryServerInterceptor), nil
			}
		}
		return nil, errors.New("middleware should be defined as 'func (" + r.String() + ") " + middleware + "() grpc.UnaryServerInterceptor'")
	}
	return nil, errors.New("could not find middleware " + middleware)
}

func GetMethodInterceptors(svc interface{}, config CommonConfig, middlewares []string) []grpc.UnaryServerInterceptor {
	interceptors := make([]grpc.UnaryServerInterceptor, 0)
	for _, middleware := range middlewares {
		interceptor, err := getMiddleware(svc, middleware)
		if err != nil {
			log.Println(err)
			notifier.NotifyWithLevel(err, "critical")
		} else {
			if interceptor != nil {
				interceptors = append(interceptors, interceptor)
			}
		}
	}
	return interceptors
}

func optionsInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	ctx = options.AddToOptions(ctx, "", "")
	if !modifiers.IsHTTPRequest(ctx) {
		options.AddToOptions(ctx, modifiers.RequestGRPC, true)
	}
	return handler(ctx, req)
}
