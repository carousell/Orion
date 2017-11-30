package executor

import (
	"errors"
	"math/rand"
	"testing"
	"time"

	"github.com/fortytw2/leaktest"
	"github.com/stretchr/testify/assert"
)

/*
no errors
all errors
random errors
test for panic
*/

const (
	runCount = 100
)

func TestNoErrors(t *testing.T) {
	defer leaktest.Check(t)()
	e := NewExecutor(WithFailOnError(true))

	for i := 0; i < rand.Intn(runCount)+10; i++ {
		e.Add(func() error {
			time.Sleep(time.Millisecond * time.Duration(rand.Intn(200)))
			return nil
		})
	}
	// wait
	assert.NoError(t, e.Wait(), "There should be no error")
}

func TestNoErrorsNoFail(t *testing.T) {
	defer leaktest.Check(t)()
	e := NewExecutor(WithFailOnError(false))

	for i := 0; i < rand.Intn(runCount)+10; i++ {
		e.Add(func() error {
			time.Sleep(time.Millisecond * time.Duration(rand.Intn(200)))
			return nil
		})
	}
	// wait
	assert.NoError(t, e.Wait(), "There should be no error")
}

func TestAllErrors(t *testing.T) {
	defer leaktest.Check(t)()
	e := NewExecutor(WithFailOnError(true))

	for i := 0; i < rand.Intn(runCount)+10; i++ {
		e.Add(func() error {
			time.Sleep(time.Millisecond * time.Duration(rand.Intn(200)))
			return errors.New("error")
		})
	}
	// wait
	assert.Error(t, e.Wait(), "There should be error")
}

func TestAllErrorsNoFail(t *testing.T) {
	defer leaktest.Check(t)()
	e := NewExecutor(WithFailOnError(false))

	for i := 0; i < rand.Intn(runCount)+10; i++ {
		e.Add(func() error {
			time.Sleep(time.Millisecond * time.Duration(rand.Intn(200)))
			return errors.New("error")
		})
	}
	// wait
	assert.NoError(t, e.Wait(), "There should be no error")
}

func TestRandomErrors(t *testing.T) {
	defer leaktest.Check(t)()
	e := NewExecutor()
	errored := false

	for i := 0; i < rand.Intn(runCount)+10; i++ {
		e.Add(func() error {
			time.Sleep(time.Millisecond * time.Duration(rand.Intn(200)))
			if rand.Intn(200) > 400 {
				errored = true
				return errors.New("error")
			}
			return nil
		})
	}

	if !errored {
		e.Add(func() error {
			return errors.New("error")
		})
	}
	// wait
	assert.Error(t, e.Wait(), "There should be error")
}
func TestRandomErrorsNoFail(t *testing.T) {
	defer leaktest.Check(t)()
	e := NewExecutor(WithFailOnError(false))
	errored := false

	for i := 0; i < rand.Intn(runCount)+10; i++ {
		e.Add(func() error {
			time.Sleep(time.Millisecond * time.Duration(rand.Intn(200)))
			if rand.Intn(200) > 400 {
				errored = true
				return errors.New("error")
			}
			return nil
		})
	}

	if !errored {
		e.Add(func() error {
			return errors.New("error")
		})
	}
	// wait
	assert.NoError(t, e.Wait(), "There should be error")
}

func TestRandomPanic(t *testing.T) {
	defer leaktest.Check(t)()
	e := NewExecutor()
	paniced := false

	for i := 0; i < rand.Intn(runCount)+10; i++ {
		e.Add(func() error {
			time.Sleep(time.Millisecond * time.Duration(rand.Intn(200)))
			if rand.Intn(500) > 300 {
				paniced = true
				panic("error")
			} else {
				return nil
			}
		})
	}

	if !paniced {
		e.Add(func() error {
			panic("error")
		})
	}
	// wait
	assert.Error(t, e.Wait(), "There should be error")
}
