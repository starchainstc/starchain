package ChainStore

import (
	"errors"
	."starchain/core/ledger"
	."starchain/common"
	."starchain/core/store"
	."starchain/core/store/LevelDBStore"

	"sync"
)

const (
	HeaderHashListCount = 2000
	CleanCacheThreshold = 2
	TaskChanCap         = 4
	DEPLOY_TRANSACTION  = "DeployTransaction"
	INVOKE_TRANSACTION  = "InvokeTransaction"
)

var (
	ErrDBNotFound = errors.New("leveldb: not found")
)

type persistTask interface{}
type persistHeaderTask struct {
	header *Header
}
type persistBlockTask struct {
	block  *Block
	ledger *Ledger
}

type ChainStore struct {
	st IStore

	taskCh chan persistTask
	quit   chan chan bool

	mu          sync.RWMutex // guard the following var
	headerIndex map[uint32]Uint256
	blockCache  map[Uint256]*Block
	headerCache map[Uint256]*Header

	currentBlockHeight uint32
	storedHeaderCount  uint32
}

func NewStore(file string) (IStore, error) {
	ldbs, err := NewLevelDBStore(file)

	return ldbs, err
}
