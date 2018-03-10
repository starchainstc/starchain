package transaction

import (
	"starchain/core/asset"
	"starchain/common"
	"starchain/crypto"
)

func NewRegisterAssetTransaction(asset *asset.Asset,amount common.Fixed64,issuer *crypto.PubKey,controller common.Uint160) (*Transaction,error){

}
