package common

import (
	"log"
	"os"
	"sync"
	"time"
)

func Fatalln(err error) {
	if err != nil {
		log.Fatalln(err)
	}
}

func Filter[T any](ss []T, test func(T) bool) (ret []T) {
	for _, s := range ss {
		if test(s) {
			ret = append(ret, s)
		}
	}
	return
}

type Retry struct {
	OnErrDuration time.Duration
	RetryLimit    int
}

func (r *Retry) Call(f func() error) error {
	errCount := 0

	for {
		err := f()

		if err == nil {
			return err
		}

		errCount++

		if errCount == r.RetryLimit {
			return err
		}

		time.Sleep(r.OnErrDuration)
	}
}

type Debouncer struct {
	Wait time.Duration
}

func Debounce[T any](f func(arg T), t time.Duration) func(arg T) {
	m := sync.RWMutex{}
	stepTime := t.Nanoseconds()
	lastTime := int64(0)

	return func(arg T) {
		m.Lock()
		defer m.Unlock()

		currentTime := int64(time.Now().Nanosecond())

		expectedNextCallTime := lastTime + stepTime

		if lastTime == 0 || expectedNextCallTime <= currentTime {
			expectedNextCallTime = currentTime
		}

		lastTime = expectedNextCallTime
		time.AfterFunc(time.Duration(expectedNextCallTime-currentTime), func() {
			f(arg)
		})
	}
}

var l = log.New(os.Stdout, "", 0)

func Log(v ...any) {
	l.SetPrefix(time.Now().Format("2006-01-02 15:04:05") + " [AAA] ")
	l.Print(v...)
}
