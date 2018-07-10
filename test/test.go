package main

import (
	"bytes"
	"strings"
	"strconv"
	"errors"
)

func main(){

}



func StringToFixed64(s string) (Fixed64, error) {
	var buffer bytes.Buffer
	//TODO: check invalid string
	di := strings.Index(s, ".")
	if len(s)-di > 9 {
		return Fixed64(0), errors.New("unsupported precision")
	}
	if di == -1 {
		buffer.WriteString(s)
		for i := 0; i < 8; i++ {
			buffer.WriteByte('0')
		}
	} else {
		buffer.WriteString(s[:di])
		buffer.WriteString(s[di+1:])
		n := 8 - (len(s) - di - 1)
		for i := 0; i < n; i++ {
			buffer.WriteByte('0')
		}
	}
	r, err := strconv.ParseInt(buffer.String(), 10, 64)
	if err != nil {
		return Fixed64(0), err
	}

	return Fixed64(r), nil
}
