package sfnt

import (
	"bytes"
	"encoding/binary"
	"io"

	"github.com/ConradIrwin/font/sfnt/mtx"
)

// https://www.w3.org/Submission/EOT/#FileFormat
type eotHeader struct {
	EOTSize			uint32
	FontDataSize	uint32
	Version			uint32
	Flags			uint32
	FontPanose		[10]byte
	Charset			byte
	Italic			byte
	Weight			uint32
	FSType			uint16
	MagicNumber		uint16

	UnicodeRange1 uint32
	UnicodeRange2 uint32
	UnicodeRange3 uint32
	UnicodeRange4 uint32

	CodePageRange1	uint32
	CodePageRange2	uint32
	CheckSumAdjustment uint32

	Reserved1 uint32
	Reserved2 uint32
	Reserved3 uint32
	Reserved4 uint32
}

// checkEOT detects if the File is an valid Embedded OpenType (.eot) file.
// If so, returns the parsed header, else returns nil. If reading fails
// returns the corresponding error.
func checkEOT(file File) (*eotHeader, error) {
	var header eotHeader
	err := binary.Read(file, binary.LittleEndian, &header)
	if err == io.EOF {
		return nil, nil
	}
	if  err != nil {
		return nil, err
	}

	// Validate the EOT magic number
	if header.MagicNumber != 0x504c {
		return nil, nil
	}
	return &header, nil
}

// parseEOT reads an Embedded OpenType (.eot) and returns the decode
// MTX data that's either an OpenType or TrueType file.
// If parsing fails, then an error is returned and the file is nil.
func parseEOT(file File, header *eotHeader) (File, error) {
	// Skip decoding the dynamic header data and seek to the FontData
	// that's appended at the end of the file
	_, err := file.Seek(-int64(header.FontDataSize), io.SeekEnd)
	if err != nil {
		return nil, err
	}
	ctf, err := mtx.DecodeCTF(file)
	if err != nil {
		return nil, err
	}
	return bytes.NewReader(ctf), nil
}
