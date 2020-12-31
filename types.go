package ohcya

import (
	"bufio"
	"log"
	"os"
)

type CyaFileType int

const (
	FileTypeUnknown CyaFileType = iota
	FileTypeCyacd   CyaFileType = iota
	FileTypeCyacd2  CyaFileType = iota

	// These are all readable lines, we shouldn't see over a typical MCU page
	// size or things are wacky. Padded for flexibility.
	MaxRecordDataSz = (16 * 1024)

	// Current Cyacd2 File Version
	FileVersionCyacd2 = 1

	// Metadata Tags
	MetaTagAppInfo = "@appinfo"
	MetaTagEIV     = "@eiv"
)

type Cya struct {
	t   CyaFileType
	f   *os.File
	b   *bufio.Scanner
	buf []byte
	Ops
	L *log.Logger
}

type Ops interface {
	Parse(*Cya) bool
	FileType() CyaFileType
	Header() Header
	Metadata() *Metadata
	Records() *Records
	GetData() []byte
}

type Cyacd struct {
	lineno int
	h      Header
	r      []Record1
	L      *log.Logger
}

type Cyacd2 struct {
	lineno int
	h      Header
	m      Metadata
	r      []Record2
	L      *log.Logger
}

// This combines both v1 and v2 file headers.
type Header struct {
	// Both
	silicon_id    uint32
	silicon_rev   uint8
	checksum_type uint8

	// Just v2
	file_version uint8
	app_id       uint8
	product_id   uint32
}

type Metadata struct {
	AppInfo
	EIV []byte
}

type AppInfo struct {
	__cy_app_verify_start        uint32
	__cy_app_verify_start_length uint32
}

type Records struct {
	CyaFileType
	Records interface{}
}

type Record1 struct {
	Array_id   uint8
	Row_number uint16
	Data_len   uint16
	Data       []byte
	Checksum   uint8
}

type Record2 struct {
	Address   uint32
	Data      []byte
	Signature []byte
}
