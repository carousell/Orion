package loggers_test

import (
	"context"
	"fmt"

	s "github.com/carousell/Orion/utils/log/loggers"
)

func ExampleFromContext() {
	ctx := context.Background()
	ctx = s.AddToLogContext(ctx, "indespensable", "amazing data")
	ctx = s.AddToLogContext(ctx, "preciousData", "valuable key")
	lf := s.FromContext(ctx)
	fmt.Println(lf)

	// Output:
	// map[indespensable:amazing data preciousData:valuable key]
}
