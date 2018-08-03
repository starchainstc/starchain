package common

import (
	"testing"
	"fmt"
)

func TestFixed64_String(t *testing.T) {
	value := "10000000000"
	f,e:=StringToFixed64(value)
	if e != nil {
		fmt.Println(e)
	}
	fmt.Printf(f.String())
}
