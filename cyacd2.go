package ohcya

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"
)

func (c *Cyacd2) GetData() []byte {
	r := c.Records().Records.([]Record2)

	// First, collect all addresses
	var I []int
	for i := range r {
		I = append(I, int(r[i].Address))
	}

	sort.Ints(I)

	c.L.Printf("base address=%#x\n", I[0])

	d := make([]byte, 0)
	var t int
	a := I[0]
	l := a

	for i := 0; i < len(I); i += 1 {
		_i := 0
		for ; _i < len(r); _i += 1 {
			if I[i] == int(r[_i].Address) {
				break
			}
		}

		a = int(r[_i].Address)
		if t != 0 {
			p := a - (l + t)
			if p > 0 {
				c.L.Printf("padding %v bytes of space...\n", p)
				pad := make([]byte, p)
				d = append(d, pad...)
			}
		}

		// Skip the second half of the data which is an embedded 256 bit
		// RSA signature
		d = append(d, r[_i].Data...)
		t = len(r[_i].Data)
		l = a
		c.L.Printf("addr=%#x datalen=%v\n", a, t)
	}

	return d
}

func (c *Cyacd2) Parse(x *Cya) bool {
	c.lineno = 1

	if !c.parseHeader(x) {
		return false
	}

	if !c.parseData(x) {
		return false
	}

	c.L.Printf("found cyacd type 2 file\n")

	return true
}

func (c *Cyacd2) Metadata() *Metadata {
	return &c.m
}

func (c *Cyacd2) Records() *Records {
	r := new(Records)
	r.CyaFileType = FileTypeCyacd2
	r.Records = c.r
	return r
}

func (c *Cyacd2) parseData(x *Cya) bool {
	var r bool
	for {
		c.lineno += 1
		r = x.b.Scan()
		if !r {
			if x.b.Err() != nil {
				c.L.Printf("error: parseData scan error (line=%d): %v\n",
					c.lineno,
					x.b.Err())
				return false
			} else {
				break
			}
		} else {
			if x.b.Text()[0] == []byte(":")[0] {
				r = c.parseDataLine(x)
				if !r {
					return false
				}
			} else if x.b.Text()[0] == []byte("@")[0] {
				r = c.parseMetaLine(x)
				if !r {
					return false
				}
			} else {
				c.L.Printf("error: unknown start-of-line character on line %d\n", c.lineno)
				return false
			}
		}
	}

	return true
}

func (c *Cyacd2) parseMetaLine(x *Cya) bool {
	var e error

	// Format should be "@<META_NAME>:<DATA>\n"
	s := strings.Split(x.b.Text(), ":")
	if len(s) != 2 {
		c.L.Printf("error: unexpected metadata line on line=%d\n", c.lineno)
		return false
	}

	if strings.ToLower(s[0]) == MetaTagAppInfo {
		// Handle @AppID
		n, e := fmt.Sscanf(s[1], "0x%x,0x%x",
			&c.m.AppInfo.__cy_app_verify_start,
			&c.m.AppInfo.__cy_app_verify_start_length)
		if e != nil || n != 2 {
			c.L.Printf("error: invalid @appinfo format on line %d: %v\n", c.lineno, e)
			return false
		}

		c.L.Printf("found @appinfo: %#v\n", c.m.AppInfo)
	} else if strings.ToLower(s[0]) == MetaTagEIV {
		// Handle EIV
		c.m.EIV, e = hex.DecodeString(s[1])
		if e != nil {
			c.L.Printf("error: can't decode EIV data (line=%d): %v\n",
				c.lineno,
				e)
			return false
		}

		c.L.Printf("found @eiv: bitlen=%d, EIV=%#v\n", len(c.m.EIV)*8, c.m.EIV)
	} else {
		c.L.Printf("error: unhandled meta tag: %v\n", s[0])
		return false
	}

	return true
}

func (c *Cyacd2) parseDataLine(x *Cya) bool {
	L := x.b.Text()

	if L[0] != []byte(":")[0] {
		c.L.Printf("error: invalid start-of-line (line=%d): %v\n", c.lineno, L[0])
		return false
	}

	var R Record2
	var e error

	// Address
	D, e := hex.DecodeString(L[1:9])
	if e != nil {
		c.L.Printf("error: can't decode address (line=%d)\n", c.lineno)
		return false
	}
	R.Address = binary.BigEndian.Uint32(D)

	// Encrypted(?)/Signed Data
	b, e := hex.DecodeString(L[9:])
	if e != nil {
		c.L.Printf("error: can't decode embedded data (line=%d): %v\n",
			c.lineno,
			e)
		return false
	}

	// XXX Warning: This order is a presumption
	R.Data = b[0:256]
	R.Signature = b[256:]

	c.r = append(c.r, R)

	return true
}

func (c *Cyacd2) parseHeader(x *Cya) bool {
	if len(x.b.Text()) != 24 {
		c.L.Printf("error: header line length invalid: %d\n", len(x.b.Text()))
		return false
	}

	// File Version
	L, e := hex.DecodeString(x.b.Text()[0:2])
	if e != nil {
		c.L.Printf("error: can't decode File Version from header: %v\n", e)
		return false
	}
	c.h.file_version = uint8(L[0])

	if c.h.file_version != FileVersionCyacd2 {
		c.L.Printf("error: invalid file version for cyacd2: %d != 1\n",
			c.h.file_version)
		return false
	}

	// Silicon ID is first; Cypress also abbreviates this Si Id
	L, e = hex.DecodeString(x.b.Text()[2:10])
	if e != nil {
		c.L.Printf("error: can't decode Silicon ID from header: %v\n", e)
		return false
	}
	c.h.silicon_id = binary.BigEndian.Uint32(L)

	// Silicon Revision
	L, e = hex.DecodeString(x.b.Text()[10:12])
	if e != nil {
		c.L.Printf("error: can't decode Silicon Revision from header: %v\n", e)
		return false
	}
	c.h.silicon_rev = uint8(L[0])

	// Checksum Type
	L, e = hex.DecodeString(x.b.Text()[12:14])
	if e != nil {
		c.L.Printf("error: can't decode Silicon Revision from header: %v\n", e)
		return false
	}
	c.h.checksum_type = uint8(L[0])

	// Application ID (see Fig 6 on Page 9 of "PSoC 6 MCU Device Firmware
	// Update Software Development Kit Guide"
	L, e = hex.DecodeString(x.b.Text()[14:16])
	if e != nil {
		c.L.Printf("error: can't decode App ID from header: %v\n", e)
		return false
	}
	c.h.app_id = uint8(L[0])

	// Product ID; arbitrary by manufacturer
	L, e = hex.DecodeString(x.b.Text()[16:24])
	if e != nil {
		c.L.Printf("error: can't decode Product ID from header: %v\n", e)
		return false
	}
	c.h.product_id = binary.BigEndian.Uint32(L)

	return true
}

func (c *Cyacd2) Header() Header {
	return c.h
}

func (c *Cyacd2) FileType() CyaFileType {
	return FileTypeCyacd2
}
