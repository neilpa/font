package main

/*
#cgo LDFLAGS: -L${SRCDIR}/lzcomp -llzcomp

#include <stdlib.h>
#include "mtx.h"
*/
import "C"
import (
	"bytes"
	"io"
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
	if !C.mtx_init(&mtx, (*C.uchar)(unsafe.Pointer(&buf[0])), C.ulong(len(buf))) {
		panic("Failed to init MTX struct")
	}
	defer C.mtx_fini(mtx)

	//C.mtx_dump(mtx)

	var rest, data, code *C.uchar
	var restSize, dataSize, codeSize C.ulong

	if !C.mtx_getRest(mtx, &rest, &restSize) {
		panic("Failed to decompress rest block")
	}
	defer C.free(unsafe.Pointer(rest))

	if !C.mtx_getData(mtx, &data, &dataSize) {
		panic("Failed to decompress data block")
	}
	defer C.free(unsafe.Pointer(code))

	if !C.mtx_getCode(mtx, &code, &codeSize) {
		panic("Failed to decompress code block")
	}
	defer C.free(unsafe.Pointer(data))

	ctf := make([]byte, 0, restSize+dataSize+codeSize)
	ctf = append(ctf, C.GoBytes(unsafe.Pointer(rest), C.int(restSize))...)
	ctf = append(ctf, C.GoBytes(unsafe.Pointer(data), C.int(dataSize))...)
	ctf = append(ctf, C.GoBytes(unsafe.Pointer(code), C.int(codeSize))...)

	io.Copy(os.Stdout, bytes.NewReader(ctf))
}
