package util

import (
	"errors"
	"log"
	"os"
	"runtime/debug"
)

func RecoverWrapFunc(h func()) func() {
	return func() {
		var err error
		defer func() {
			r := recover()
			if r != nil {
				switch t := r.(type) {
				case string:
					err = errors.New(t)
				case error:
					err = t
				default:
					err = errors.New("unknown error")
				}
				log.Println(err.Error())
				log.Printf("stack: %s", debug.Stack())
			}
		}()
		h()
	}
}

func Exists(path string) bool {
	_, err := os.Stat(path)
	if err != nil {
		if os.IsExist(err) {
			return true
		}
		return false
	}
	return true
}
