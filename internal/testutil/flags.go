package testutil

import (
	"os"
	"strconv"
)

func IsDisableInlining() bool {
	flag := os.Getenv("DISABLE_INLINING")
	ret, _ := strconv.ParseBool(flag)
	return ret
}
