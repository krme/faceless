package helper

import "time"

func Throttle() {
	time.Sleep(time.Duration(time.Second * 1))
}
