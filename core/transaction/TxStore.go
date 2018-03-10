package transaction

import ."starchain/common"

type ILedgerStore interface {
	GetTransaction(hash Uint256) (*Transaction, error)
	GetQuantityIssued(AssetId Uint256) (Fixed64, error)
}

