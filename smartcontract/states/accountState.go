package states

import (
	"starchain/common"
	"starchain/common/serialization"
	"bytes"
	"io"
)

type AccountState struct {
	StateBase
	ProgramHash common.Uint160
	IsFrozen bool
	Balances map[common.Uint256]common.Fixed64
}

func NewAccountState(hash common.Uint160,balance map[common.Uint256]common.Fixed64) *AccountState{
	var as AccountState
	as.ProgramHash = hash
	as.Balances = balance
	as.IsFrozen = false
	return &as
}

func(accountState *AccountState)Serialize(w io.Writer) error {
	accountState.StateBase.Serialize(w)
	accountState.ProgramHash.Serialize(w)
	serialization.WriteBool(w, accountState.IsFrozen)
	serialization.WriteUint64(w, uint64(len(accountState.Balances)))
	for k, v := range accountState.Balances {
		k.Serialize(w)
		v.Serialize(w)
	}
	return nil
}

func(accountState *AccountState)Deserialize(r io.Reader) error {
	stateBase := new(StateBase)
	err := stateBase.Deserialize(r)
	if err != nil {
		return err
	}
	accountState.StateBase = *stateBase
	accountState.ProgramHash.Deserialize(r)
	isFrozen, err := serialization.ReadBool(r)
	if err != nil {
		return err
	}
	accountState.IsFrozen = isFrozen
	l, err := serialization.ReadUint64(r)
	if err != nil {
		return err
	}
	balances := make(map[common.Uint256]common.Fixed64, 0)
	u := new(common.Uint256)
	f := new(common.Fixed64)
	for i:=0; i<int(l); i++ {
		err = u.Deserialize(r)
		if err != nil { return err }
		err = f.Deserialize(r)
		if err != nil { return err }
		balances[*u] = *f
	}
	accountState.Balances = balances
	return nil
}

func(accountState *AccountState) ToArray() []byte {
	b := new(bytes.Buffer)
	accountState.Serialize(b)
	return b.Bytes()
}
