package forkjoin

import (
	"errors"
	"fmt"
	"testing"
	"time"
)

func TestForkJoin(t *testing.T) {
	f := NewForkJoin()
	f.Add(func() error {
		return errors.New("error")
	})
	f.Add(func() error {
		time.Sleep(time.Second * 1)
		fmt.Println("slept for 1 sec")
		return nil
	})
	f.Add(func() error {
		time.Sleep(time.Millisecond * 500)
		fmt.Println("slept for 0.5 sec")
		return errors.New("error")
	})
	f.Wait()
}
