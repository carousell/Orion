package log

import (
	"context"
	"fmt"
	"google.golang.org/grpc/metadata"
	"testing"
)

func TestName(t *testing.T) {
	ctx := context.Background()
	ctx = metadata.AppendToOutgoingContext(ctx, "a", "b")
	ctx = metadata.AppendToOutgoingContext(ctx, "c", "d")
	ctx = metadata.AppendToOutgoingContext(ctx, "a", "b")
	ctx = metadata.AppendToOutgoingContext(ctx, "a", "c")

	md, ok := metadata.FromOutgoingContext(ctx)
	if !ok {
		fmt.Println("error")
		return
	}
	for k, v := range md {
		fmt.Println(k, v)
	}

}

func TestAppendTracingInfoToLoggingContext(t *testing.T) {

	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			AppendTracingInfoToLoggingContext(tt.args.ctx)
		})
	}
}
