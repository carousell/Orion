package helpers

import (
	"math/rand"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/carousell/Orion/orion"
	"github.com/stretchr/testify/assert"
)

type svc struct {
	version uint64
}

var counter sync.Map

type testFactory struct {
	test *testing.T
}

func (t *testFactory) NewService(svr orion.Server, params orion.FactoryParams) interface{} {
	initVal := int32(0)
	counter, _ := counter.LoadOrStore(params.Version, &initVal)
	c, _ := counter.(*int32)
	val := atomic.AddInt32(c, 1)
	assert.Equal(t.test, int32(1), val, "create service mismatch")
	return svc{params.Version}
}

func (t *testFactory) DisposeService(svc interface{}, param orion.FactoryParams) {
	counter, found := counter.Load(param.Version)
	if !found {
		assert.Fail(t.test, "version not found in dispose ")
		return
	}
	c, _ := counter.(*int32)
	val := atomic.AddInt32(c, -1)
	assert.Equal(t.test, int32(0), val, "dispose mismatch for version %d", param.Version)
}

func TestSingleServiceFactory(t *testing.T) {
	tf := &testFactory{t}
	sf, err := NewSingleServiceFactory(tf)
	assert.NoError(t, err, "Error during NewSingleServiceFactory")
	assert.NotNil(t, sf, "SingleServiceFactory is nil")
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		param := orion.FactoryParams{
			Version: uint64(i),
		}
		go func() {
			defer wg.Done()
			for i := int32(0); i < rand.Int31n(10)+1; i++ {
				svc := sf.NewService(nil, param)
				wg.Add(1)
				go func() {
					defer wg.Done()
					sf.DisposeService(svc, param)
				}()
			}
		}()
	}
	wg.Wait()
}
