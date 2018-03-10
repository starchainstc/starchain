package signature

import (
	"crypto"
	"starchain/vm/avm/interfaces"
)

type Signer interface {
	PriKey() []byte
	PubKey() *crypto.PublicKey
}


type SignableData interface {
	interfaces.ICodeContainer
}