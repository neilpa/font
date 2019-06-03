package main

/*
#cgo LDFLAGS: -L${SRCDIR}/lzcomp -llzcomp

#include "mtx.h"
*/
import "C"

import (
	"io/ioutil"
	"os"
	"unsafe"
)

func main() {
	f, err := os.Open(os.Args[1])
	if err != nil {
		panic(err)
	}
	buf, err := ioutil.ReadAll(f)
	if err != nil {
		panic(err)
	}

    var mtx *C.mtx_t
	C.mtx_init(&mtx, (*C.uchar)(unsafe.Pointer(&buf[0])), C.ulong(len(buf)))
	C.mtx_dump(mtx)
}
