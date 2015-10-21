package cog

import (
	"testing"

	"github.com/thatguystone/cog/check"
)

func TestExitBasic(t *testing.T) {
	check.New(t)

	ex := NewExit()

	ex.Add(1)
	go func() {
		defer ex.Done()

		for {
			select {
			case <-ex.C:
				return
			}
		}
	}()

	ex.Exit()
	ex.Wait()
}
