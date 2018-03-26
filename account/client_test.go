package account

import (
	"testing"
	"fmt"
	"Elastos.ELA/common"
)

func TestClient(t *testing.T){
	client := NewClient("./wallet.dat",[]byte("kpcloud"),false)
	en,err:=client.EncryptPrivateKey([]byte("save"))
	if err != nil {
		fmt.Println("error:",err)
	}
	fmt.Println(common.BytesToHexString(en))
}
