package net

import (
	"starchain/common"
	"starchain/core/transaction"
	"starchain/events"
	"starchain/crypto"
	"starchain/core/ledger"
	"starchain/net/protocol"
	"starchain/net/node"
	."starchain/errors"
)

type Neter interface {
	GetTxnPool(byCount bool) map[common.Uint256]*transaction.Transaction
	Xmit(interface{}) error
	GetEvent(eventName string) *events.Event
	GetBookKeepersAddrs() ([]*crypto.PubKey, uint64)
	CleanSubmittedTransactions(block *ledger.Block) error
	GetNeighborNoder() []protocol.Noder
	Tx(buf []byte)
	AppendTxnPool(*transaction.Transaction, bool) ErrCode
}

func StartProtocol(pubKey *crypto.PubKey) protocol.Noder {
	net := node.InitNode(pubKey)
	net.ConnectSeeds()

	return net
}
