package signature

import "starchain/crypto"

type Signer interface {
	PrivKey() []byte
	PubKey() *crypto.PubKey
}

