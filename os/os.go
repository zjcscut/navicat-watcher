package os

import "runtime"

const WT = "windows"

// 只支持windows
func init() {
	st := runtime.GOOS
	if WT != st {
		panic("unsupported system type " + st)
	}
}
