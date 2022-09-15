//go test -race
package loggers_test

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"testing"
	"time"

	s "github.com/carousell/Orion/utils/log/loggers"
	"github.com/stretchr/testify/assert"
)

const readWorkerCount = 50
const writeWorkerCount = 50

func readWorker(idx int, ctx context.Context) {
	s.FromContext(ctx)
	// simulate reading task
	time.Sleep(time.Millisecond * 250)
}

func writeWorker(idx int, ctx context.Context) context.Context {
	key := fmt.Sprintf("key%d", idx)
	val := fmt.Sprintf("val%d", rand.Intn(10000))
	ctx = s.AddToLogContext(ctx, key, val)
	time.Sleep(time.Millisecond * 250)
	return ctx
}

func TestParallelRead(t *testing.T) {
	// LogContext init, non-paralel
	ctx := context.Background()
	ctx = s.AddToLogContext(ctx, "k1", "v1")
	ctx = s.AddToLogContext(ctx, "k2", "v2")

	var wg sync.WaitGroup
	for i := 1; i <= readWorkerCount; i++ {
		wg.Add(1)
		go func(j int) {
			defer wg.Done()
			readWorker(j, ctx)
		}(i)
	}
	wg.Wait()
}

func TestParallelWrite(t *testing.T) {
	ctx := context.Background()
	ctx = s.AddToLogContext(ctx, "test-key", "test-value")

	var wg sync.WaitGroup
	for i := 1; i <= writeWorkerCount; i++ {
		wg.Add(1)
		go func(j int) {
			defer wg.Done()
			writeWorker(j, ctx)
		}(i)
	}
	wg.Wait()

	lf := s.FromContext(ctx)

	assert.Contains(t, lf, "test-key")
	for i := 1; i <= writeWorkerCount; i++ {
		key := fmt.Sprintf("key%d", i)
		assert.Contains(t, lf, key)
	}
}

func TestParallelReadAndWrite(t *testing.T) {
	ctx := context.Background()
	ctx = s.AddToLogContext(ctx, "test-key", "test-value")

	var wg sync.WaitGroup
	for i := 1; i <= readWorkerCount; i++ {
		wg.Add(1)
		go func(j int) {
			defer wg.Done()
			readWorker(j, ctx)
		}(i)
	}
	for i := 1; i <= writeWorkerCount; i++ {
		wg.Add(1)
		go func(j int) {
			defer wg.Done()
			writeWorker(j, ctx)
		}(i)
	}
	wg.Wait()

	lf := s.FromContext(ctx)

	assert.Contains(t, lf, "test-key")
	for i := 1; i <= writeWorkerCount; i++ {
		key := fmt.Sprintf("key%d", i)
		assert.Contains(t, lf, key)
	}
}