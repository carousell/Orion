//go test -v -bench=. -run=none .
package loggers_test

import (
	"context"
	"fmt"
	"testing"

	s "github.com/carousell/Orion/utils/log/loggers"
)

func BenchmarkFromContext(b *testing.B) {
	ctx := context.Background()
	for i := 0; i < b.N; i++ {
		s.FromContext(ctx)
	}
}

func BenchmarkFromAddToLogContext(b *testing.B) {
	ctx := context.Background()
	for i := 0; i < b.N; i++ {
		s.AddToLogContext(ctx, fmt.Sprintf("key%d", i), "good value")
	}
}
