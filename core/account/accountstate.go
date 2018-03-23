package account

import (
	"starchain/common"
	"io"
	"starchain/common/serialization"
	"bytes"
)

type AccountState struct{
	ProgramHash common.Uint160
	IsFrozen bool
	Balances map[common.Uint256]common.Fixed64

}

func NewAccountState(programhash common.Uint160,balance map[common.Uint256]common.Fixed64) *AccountState{
	var accountState AccountState
	accountState.ProgramHash = programhash
	accountState.Balances = balance
	accountState.IsFrozen = false
	return &accountState
}

func (acc *AccountState) Serialize(w io.Writer) (error){
	acc.ProgramHash.Serialize(w)
	serialization.WriteBool(w,acc.IsFrozen)
	serialization.WriteUint64(w,uint64(len(acc.Balances)))
	for k,v := range acc.Balances{
		k.Serialize(w)
		v.Serialize(w)
	}
	return nil
}

func (acc *AccountState) Deserialize(r io.Reader) error{
	acc.ProgramHash.Deserialize(r)
	isFrozen,err := serialization.ReadBool(r)
	if err != nil {
		return err
	}
	acc.IsFrozen = isFrozen
	len,err := serialization.ReadUint64(r)
	if err != nil {
		return err
	}
	balance := make(map[common.Uint256]common.Fixed64)
	u := new(common.Uint256)
	f := new(common.Fixed64)
	for i:=0;i<int(len);i++{
		err = u.Deserialize(r)
		if err != nil {
			return err
		}
		err = f.Deserialize(r)
		if err != nil {
			return err
		}
		balance[*u] = *f
	}
	acc.Balances = balance
	return nil

}

func (acc *AccountState) ToArray() []byte{
	b:=new(bytes.Buffer)
	acc.Serialize(b)
	return b.Bytes()
}