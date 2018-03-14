package signature

import (
	"crypto"
)

type Signer interface {
	PrivKey() []byte
	PubKey() *crypto.PublicKey
}

