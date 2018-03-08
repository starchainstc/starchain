package signature

import (
	"crypto"
	"starchain/vm/avm/interfaces"
)

type Signer interface {
	Prikey() []byte
	Pubkey() *crypto.PublicKey
}


type SignableData interface {
	interfaces.ICodeContainer
}