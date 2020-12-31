package ohcya_test // import "github.com/DefenseMouse/ohcya"

import (
	"fmt"
	"github.com/DefenseMouse/ohcya"
	"os"
	"testing"
)

func TestOpen(t *testing.T) {
	c := ohcya.New(nil)
	c.Open("./foo.cyacd")

	f, e := os.OpenFile("./output.test", os.O_CREATE|os.O_WRONLY, 0644)
	if e != nil {
		fmt.Printf("can't open file for write: %v\n", e)
		return
	}
	defer f.Close()

	b := c.GetData()
	if b != nil {
		f.Write(b)
	}
}
