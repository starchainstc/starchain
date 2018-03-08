package asset

import (
	"io"
	"starchain/common/serialization"
)

type LockAsset struct{
	Lock	uint32
	Unlock	uint32
	Amount uint32
}

func (la *LockAsset)Serialize(w io.Writer) error{
	if err := serialization.WriteUint32(w,la.Lock);err != nil{
		return err
	}
	if err := serialization.WriteUint32(w,la.Unlock);err != nil{
		return err
	}
	if err := serialization.WriteUint32(w,la.Amount);err != nil{
		return err
	}
	return nil
}

func (la *LockAsset)Deserialize(r io.Reader) error{
	lock,err := serialization.ReadUint32(r)
	if err != nil {
		return err
	}
	la.Lock = lock
	la.Unlock,err = serialization.ReadUint32(r)
	if err != nil {
		return err
	}
	la.Amount,err = serialization.ReadUint32(r)
	if err != nil {
		return err
	}
	return nil
}
