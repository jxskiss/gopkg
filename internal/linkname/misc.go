package linkname

import (
	"log"
	"unsafe"
)

var _ unsafe.Pointer

//go:linkname LogStd log.std
var LogStd *log.Logger
