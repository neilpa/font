package sfnt

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"unicode/utf16"

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

type eotDynamicHeader struct {
	// Version 0x0001000 or greater
	//

	Padding1 uint16
	FamilyNameSize uint16
	FamilyName []uint16 // TODO UTF-16 string

	Padding2 uint16
	StyleNameSize uint16
	StyleName []uint16 // TODO UTF-16 string

	Padding3 uint16
	VersionNameSize uint16
	VersionName []uint16 //TODO UTF-16 string

	Padding4 uint16
	FullNameSize uint16
	FullName []uint16 // TODO UTF-16 string

	// Version 0x00020001 or greater
	//

	Padding5 uint16
	RootStringSize uint16
	RootString []uint16 // TODO UTF-16 string

	// Version 0x00020002 or greater
	//

	RootStringCheckSum uint32
	EUDCCodePage uint32

	Padding6 uint16
	SignatureSize uint16
	Signature []byte

	EUDCFlags uint32
	EUDCFontSize uint32

	// TODO EUDCFontData [/*EUDCFontSize*/]byte
	// TODO FontData [/*FontDataSize*/]byte
}

// checkEOT detects if the File is an valid Embedded OpenType (.eot) file.
// If so, returns the parsed header, else returns nil. If reading fails
// returns the corresponding error.
func checkEOT(file File) (*eotHeader, error) {
	var header eotHeader
	file.Seek(0, 0)
	switch f := file.(type) {
		case *os.File:
		fi, _ := f.Stat()
		fmt.Println(fi.Size())
	}

	err := binary.Read(file, binary.LittleEndian, &header)
	if err == io.EOF {
		return nil, nil
	}
	if  err != nil {
		return nil, err
	}

	// Validate the EOT magic number
	fmt.Printf("Magic 0x%X\n", header.MagicNumber)
	if header.MagicNumber != 0x504c {
		fmt.Println("Not right")
		return nil, nil
	}
	return &header, nil
}

// parseEOT reads an Embedded OpenType (.eot) and returns the decode
// MTX data that's either an OpenType or TrueType file. Expects that the
// file is seeked at the end of the provided header.
// If parsing fails, then an error is returned and the file is nil.
func parseEOT(file File, header *eotHeader) (File, error) {
	var dynHeader eotDynamicHeader
	var err error
	switch header.Version {
		case 0x00010000:
			err = readEOTHeaderV10(file, &dynHeader)
		case 0x00020001:
			err = readEOTHeaderV21(file, &dynHeader)
		case 0x00020002:
			err = readEOTHeaderV22(file, &dynHeader)
		default:
			err = fmt.Errorf("Invalid EOT Version (0x%X)", header.Version)
	}
	if err != nil {
		return nil, err
	}

	fmt.Printf("%+v\n", dynHeader)

	ctf, err := mtx.DecodeCTF(file)
	if err != nil {
		return nil, err
	}
	return bytes.NewReader(ctf), nil
}

func readEOTHeaderV10(r io.Reader, dynHeader *eotDynamicHeader) error {
	err := readPaddedUTF16(r, &dynHeader.Padding1, &dynHeader.FamilyNameSize, &dynHeader.FamilyName)
	if err != nil {
		return err
	}
	err = readPaddedUTF16(r, &dynHeader.Padding2, &dynHeader.StyleNameSize, &dynHeader.StyleName)
	if err != nil {
		return err
	}
	err = readPaddedUTF16(r, &dynHeader.Padding3, &dynHeader.VersionNameSize, &dynHeader.VersionName)
	if err != nil {
		return err
	}
	return readPaddedUTF16(r, &dynHeader.Padding4, &dynHeader.FullNameSize, &dynHeader.FullName)
}

func readEOTHeaderV21(r io.Reader, dynHeader *eotDynamicHeader) error {
	err := readEOTHeaderV10(r, dynHeader)
	if err != nil {
		return err
	}
	return readPaddedUTF16(r, &dynHeader.Padding5, &dynHeader.RootStringSize, &dynHeader.RootString)
}

func readEOTHeaderV22(r io.ReadSeeker, dynHeader *eotDynamicHeader) error {
	err := readEOTHeaderV21(r, dynHeader)
	if err != nil {
		return err
	}
	err = binary.Read(r, binary.LittleEndian, &dynHeader.RootStringCheckSum)
	if err != nil {
		return err
	}
	err = binary.Read(r, binary.LittleEndian, &dynHeader.EUDCCodePage)
	if err != nil {
		return err
	}
	err = readPaddedSize(r, &dynHeader.Padding6, &dynHeader.SignatureSize)
	if err != nil {
		return err
	}
	dynHeader.Signature = make([]byte, dynHeader.SignatureSize)
	err = binary.Read(r, binary.LittleEndian, dynHeader.Signature)
	if err != nil {
		return err
	}
	err = binary.Read(r, binary.LittleEndian, dynHeader.EUDCFlags)
	if err != nil {
		return err
	}
	err = binary.Read(r, binary.LittleEndian, dynHeader.EUDCFontSize)
	if err != nil {
		return err
	}
	// TODO Skip EUDCFontData for now
	_ , err = r.Seek(int64(dynHeader.EUDCFontSize), io.SeekCurrent)
	return err
}

func readPaddedUTF16(r io.Reader, padding, bytes *uint16, buffer *[]uint16) error {
	err := readPaddedSize(r, padding, bytes)
	if err != nil {
		return err
	}
	// TODO Can bytes by 0? What about non-even?
	*buffer = make([]uint16, *bytes/2)
	err = binary.Read(r, binary.LittleEndian, *buffer)
	fmt.Println(string(utf16.Decode(*buffer)))
	return err
}

func readPaddedSize(r io.Reader, padding, bytes *uint16) error {
	err := binary.Read(r, binary.LittleEndian, padding)
	if err != nil {
		return err
	}
	return binary.Read(r, binary.LittleEndian, bytes)
}
