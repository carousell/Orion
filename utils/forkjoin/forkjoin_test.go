package forkjoin

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
	f := NewForkJoin()

	for i := 0; i < rand.Intn(runCount)+10; i++ {
		f.Add(func() error {
			time.Sleep(time.Millisecond * time.Duration(rand.Intn(200)))
			return nil
		})
	}
	// wait
	assert.NoError(t, f.Wait(), "There should be no error")
}

func TestAllErrors(t *testing.T) {
	defer leaktest.Check(t)()
	f := NewForkJoin()

	for i := 0; i < rand.Intn(runCount)+10; i++ {
		f.Add(func() error {
			time.Sleep(time.Millisecond * time.Duration(rand.Intn(200)))
			return errors.New("error")
		})
	}
	// wait
	assert.Error(t, f.Wait(), "There should be error")
}

func TestRandomErrors(t *testing.T) {
	defer leaktest.Check(t)()
	f := NewForkJoin()
	errored := false

	for i := 0; i < rand.Intn(runCount)+10; i++ {
		f.Add(func() error {
			time.Sleep(time.Millisecond * time.Duration(rand.Intn(200)))
			if rand.Intn(200) > 400 {
				errored = true
				return errors.New("error")
			} else {
				return nil
			}
		})
	}

	if !errored {
		f.Add(func() error {
			return errors.New("error")
		})
	}
	// wait
	assert.Error(t, f.Wait(), "There should be error")
}

func TestRandomPanic(t *testing.T) {
	defer leaktest.Check(t)()
	f := NewForkJoin()
	paniced := false

	for i := 0; i < rand.Intn(runCount)+10; i++ {
		f.Add(func() error {
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
		f.Add(func() error {
			panic("error")
		})
	}
	// wait
	assert.Error(t, f.Wait(), "There should be error")
}
