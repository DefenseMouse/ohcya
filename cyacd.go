package ohcya

import (
	"encoding/binary"
	"encoding/hex"
)

func (c *Cyacd) GetData() []byte {
	// XXX
	// The value of this is questionable, but actually seeing the data is
	// useful enough. If you want a real "image" of how the data would be
	// loaded into flash, this is not it. This should really only be used for
	// analyzing the data/instructions.
	d := make([]byte, 0)

	for i := 0; i < len(c.r); i += 1 {
		d = append(d, c.r[i].Data...)
	}

	return d
}

func (c *Cyacd) Parse(x *Cya) bool {
	c.lineno = 1

	if !c.parseHeader(x) {
		return false
	}

	if !c.parseData(x) {
		return false
	}

	x.L.Printf("found cyacd type 1 file\n")

	return true
}

func (c *Cyacd) Metadata() *Metadata {
	return nil
}

func (c *Cyacd) Records() *Records {
	r := new(Records)
	r.CyaFileType = FileTypeCyacd
	r.Records = c.r
	return r
}

func (c *Cyacd) parseData(x *Cya) bool {
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
			r = c.parseDataLine(x)
			if !r {
				return false
			}
		}
	}

	return true
}

func (c *Cyacd) parseDataLine(x *Cya) bool {
	L := x.b.Text()

	if L[0] != []byte(":")[0] {
		c.L.Printf("error: invalid start-of-line (line=%d): %v\n", c.lineno, L[0])
		return false
	}

	var R Record1
	var e error

	// Array ID
	D, e := hex.DecodeString(L[1:3])
	if e != nil {
		c.L.Printf("error: can't decode array Id (line=%d)\n", c.lineno)
		return false
	}
	R.Array_id = uint8(D[0])

	// Row Number
	D, e = hex.DecodeString(L[3:7])
	if e != nil {
		c.L.Printf("error: can't decode array Id (line=%d)\n", c.lineno)
		return false
	}
	R.Row_number = binary.BigEndian.Uint16(D)

	// Data Len
	D, e = hex.DecodeString(L[7:11])
	if e != nil {
		c.L.Printf("error: can't decode array Id (line=%d)\n", c.lineno)
		return false
	}
	R.Data_len = binary.BigEndian.Uint16(D)

	// Check sanity
	if R.Data_len == 0 || R.Data_len > MaxRecordDataSz {
		c.L.Printf("error: data exceeds maximum data size (line=%d)\n",
			c.lineno)
		return false
	}

	// Check fit, but allow for trailing data so we can warn
	if (len(L) - 13) < int((R.Data_len * 2)) {
		c.L.Printf("error: Data_len %v doesn't match actual line (line=%d)\n",
			R.Data_len,
			c.lineno)
		return false
	}

	R.Data, e = hex.DecodeString(L[11 : 11+(R.Data_len*2)])
	if e != nil {
		c.L.Printf("error: can't decode embedded data (line=%d): %v\n",
			c.lineno,
			e)
		return false
	}

	t := 11 + (R.Data_len * 2)

	// Checksum
	D, e = hex.DecodeString(L[t : t+2])
	if e != nil {
		c.L.Printf("error: can't decode checksum (line=%d): %v\n",
			c.lineno,
			e)
		return false
	}
	R.Checksum = uint8(D[0])

	// Check for trailing data ;)
	if len(L) != int(t+2) {
		c.L.Printf("warning: line length doesn't meet expectations: L=%v t+2=%v\n",
			len(L),
			t+2)
	}

	c.r = append(c.r, R)

	return true
}

func (c *Cyacd) parseHeader(x *Cya) bool {
	// Silicon ID is first; Cypress also abbreviates this Si Id
	L, e := hex.DecodeString(x.b.Text()[0:8])
	if e != nil {
		c.L.Printf("error: can't decode Silicon ID from header: %v\n", e)
		return false
	}
	c.h.silicon_id = binary.BigEndian.Uint32(L)

	// Silicon Revision
	L, e = hex.DecodeString(x.b.Text()[8:10])
	if e != nil {
		c.L.Printf("error: can't decode Silicon Revision from header: %v\n", e)
		return false
	}
	c.h.silicon_rev = uint8(L[0])

	// Checksum Type
	L, e = hex.DecodeString(x.b.Text()[10:12])
	if e != nil {
		c.L.Printf("error: can't decode Silicon Revision from header: %v\n", e)
		return false
	}
	c.h.checksum_type = uint8(L[0])

	return true
}

func (c *Cyacd) Header() Header {
	return c.h
}

func (c *Cyacd) FileType() CyaFileType {
	return FileTypeCyacd
}
