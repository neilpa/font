package mtx

/*
#cgo CFLAGS: -Wno-format -Wno-pointer-sign -Wno-implicit-int
#cgo LDFLAGS: -L${SRCDIR}/lzcomp -llzcomp

#include <stdlib.h>
#include "mtx.h"
*/
import "C"
import (
	"fmt"
	"io"
	"io/ioutil"
	"unsafe"
)

func DecodeCTF(r io.Reader) ([]byte, error) {
	buf, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}

    var mtx *C.mtx_t
	if !C.mtx_init(&mtx, (*C.uchar)(unsafe.Pointer(&buf[0])), C.ulong(len(buf))) {
		return nil, fmt.Errorf("Failed to init MTX struct")
	}
	defer C.mtx_fini(mtx)

	var rest, data, code *C.uchar
	var restSize, dataSize, codeSize C.ulong

	if !C.mtx_getRest(mtx, &rest, &restSize) {
		return nil, fmt.Errorf("Failed to decompress rest block")
	}
	defer C.free(unsafe.Pointer(rest))

	if !C.mtx_getData(mtx, &data, &dataSize) {
		return nil, fmt.Errorf("Failed to decompress data block")
	}
	defer C.free(unsafe.Pointer(code))

	if !C.mtx_getCode(mtx, &code, &codeSize) {
		return nil, fmt.Errorf("Failed to decompress code block")
	}
	defer C.free(unsafe.Pointer(data))

	ctf := make([]byte, 0, restSize+dataSize+codeSize)
	ctf = append(ctf, C.GoBytes(unsafe.Pointer(rest), C.int(restSize))...)
	ctf = append(ctf, C.GoBytes(unsafe.Pointer(data), C.int(dataSize))...)
	ctf = append(ctf, C.GoBytes(unsafe.Pointer(code), C.int(codeSize))...)

	return ctf, nil
}
