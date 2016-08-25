package cog

import (
	"fmt"
	"testing"

	"github.com/iheartradio/cog/check"
)

type testExiter struct {
	i     int
	exits *[]int
}

func (t testExiter) Exit() {
	*t.exits = append(*t.exits, t.i)
}

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

func TestExitExiters(t *testing.T) {
	const total = 10

	c := check.New(t)

	ex := NewExit()

	exits := []int{}
	for i := 0; i < total; i++ {
		ex.AddExiter(testExiter{
			i:     i + 1,
			exits: &exits,
		})
	}

	ex.Exit()
	ex.Wait()

	// Exiters should be called in reverse-order
	for i, ei := range exits {
		c.Equal(total-i, ei)
	}
}

func ExampleExit() {
	e := NewExit()

	run := func(i int) {
		e.Add(1)
		fmt.Println(i, "started")

		go func() {
			defer e.Done()
			defer fmt.Println("exited")

			for {
				select {
				case <-e.C:
					return
				}
			}
		}()
	}

	for i := 0; i < 10; i++ {
		run(i)
	}

	// Wait for all goroutines to exit
	e.Exit()

	// Output:
	// 0 started
	// 1 started
	// 2 started
	// 3 started
	// 4 started
	// 5 started
	// 6 started
	// 7 started
	// 8 started
	// 9 started
	// exited
	// exited
	// exited
	// exited
	// exited
	// exited
	// exited
	// exited
	// exited
	// exited
}
