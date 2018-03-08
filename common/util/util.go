package util

import (
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"crypto/hmac"
)

const (
	HASHLEN       = 32
	PRIVATEKEYLEN = 32
	PUBLICKEYLEN  = 32
	SIGNRLEN      = 32
	SIGNSLEN      = 32
	SIGNATURELEN  = 64
	NEGBIGNUMLEN  = 33
)

type CryptoAlgSet struct {
	EccParams elliptic.CurveParams
	Curve     elliptic.Curve
}

func RandomNum(n int)([]byte,error){
	b := make([]byte,n)
	_,err := rand.Read(b)
	if err != nil {
		return nil,err
	}
	return b,nil
}

func Hash(data []byte) [HASHLEN]byte{
	return sha256.Sum256(data)
}

func CheckMAC(msg,mmac,key []byte) bool {
	mac:= hmac.New(sha256.New,key)
	mac.Write(msg)
	tarMac := mac.Sum(nil)
	return hmac.Equal(tarMac,mmac)
}
