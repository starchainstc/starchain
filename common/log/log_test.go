package log

import (
	"testing"
	"fmt"
	"runtime"
)


func TestGetGid(t *testing.T){
	gid := GetGID()
	fmt.Println(gid)
	var buf [64]byte
	res:=runtime.Stack(buf[:],false)
	fmt.Println(res)
	fmt.Println("finish")
}