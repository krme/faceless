package helper

import (
	"fmt"
	"hash/fnv"
)

func Hash32(s string) string {
	h := fnv.New32a()
	h.Write([]byte(s))
	return fmt.Sprint(h.Sum32())
}
