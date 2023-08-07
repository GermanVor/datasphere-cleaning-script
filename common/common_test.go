package common

import (
	"sync"
	"testing"
	"time"
)

type TestObj struct {
	V int
}

func TestDebounce(t *testing.T) {
	t.Run("", func(t *testing.T) {
		arr := []TestObj{
			{V: 1},
			{V: 2},
			{V: 3},
			{V: 4},
		}

		wg := sync.WaitGroup{}
		wg.Add(len(arr))

		originF := func(arg TestObj) {
			Log(arg.V)
			wg.Done()
		}

		debouncedF := Debounce(originF, time.Second)

		for _, tObj := range arr {
			go debouncedF(tObj)
			time.Sleep(500 * time.Microsecond)
		}

		wg.Wait()
	})
}

// 36
