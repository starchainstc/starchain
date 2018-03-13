package ledger

import (
	"starchain/events"
	"sync"
	"starchain/crypto"
	."starchain/errors"
)

type Blockchain struct {
	BlockHeight uint32
	BCEvents *events.Event
	mutex sync.Mutex
}

func NewBlockchain(height uint32) *Blockchain{
	return &Blockchain{
		BlockHeight:height,
		BCEvents:events.NewEvent(),
	}
}

func NewBlockchainWithGenesisBlock(defBookKeeper []*crypto.PubKey) (*Blockchain,error){
	genesisBlock,err := GenesisBlockInit(defBookKeeper)
	if err != nil {
		return nil, NewDetailErr(err, ErrNoCode, "[Blockchain], NewBlockchainWithGenesisBlock failed.")
	}
	genesisBlock.RebuildMerkleRoot()
	hashx := genesisBlock.Hash()
	genesisBlock.hash = &hashx
}
