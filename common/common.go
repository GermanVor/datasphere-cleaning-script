package common

import "log"

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
