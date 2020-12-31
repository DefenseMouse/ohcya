package ohcya // import "github.com/DefenseMouse/ohcya"

import (
	"bufio"
	"log"
	"os"
)

func New(L *log.Logger) *Cya {
	c := new(Cya)

	if L != nil {
		c.L = L
	} else {
		c.L = log.New(os.Stdout, "ohcya: ", log.LstdFlags)
	}

	return c
}

func (c *Cya) Open(p string) bool {
	var e error
	c.f, e = os.Open(p)
	if e != nil {
		c.L.Printf("error: can't open file %v: %v\n", p, e)
		return false
	}

	c.b = bufio.NewScanner(c.f)
	c.b.Scan()

	// Check if it's a cyacd file by looking for the header newline
	if len(c.b.Text()) == 12 {
		c.t = FileTypeCyacd
		x := new(Cyacd)
		x.L = c.L
		c.Ops = x
	} else {
		c.t = FileTypeCyacd2
		x := new(Cyacd2)
		x.L = c.L
		c.Ops = x
	}

	return c.Ops.Parse(c)
}

func (c *Cya) GetData() []byte {
	return c.Ops.GetData()
}
