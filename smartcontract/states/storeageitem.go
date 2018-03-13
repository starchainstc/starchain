package states

import (
	"io"
	"starchain/common/serialization"
	"bytes"
	."starchain/errors"
)

type StorageItem struct {
	StateBase
	Value []byte
}


func NewStoreageItem(value []byte) *StorageKey{
	var item StorageItem
	item.Value = value
	return &item
}

func(storageItem *StorageItem)Serialize(w io.Writer) error {
	storageItem.StateBase.Serialize(w)
	serialization.WriteVarBytes(w, storageItem.Value)
	return nil
}

func(storageItem *StorageItem)Deserialize(r io.Reader) error {
	stateBase := new(StateBase)
	err := stateBase.Deserialize(r)
	if err != nil {
		return err
	}
	storageItem.StateBase = *stateBase
	value, err := serialization.ReadVarBytes(r)
	if err != nil {
		return NewDetailErr(err, ErrNoCode, "ContractState Code Deserialize fail.")
	}
	storageItem.Value = value
	return nil
}

func(storageItem *StorageItem) ToArray() []byte {
	b := new(bytes.Buffer)
	storageItem.Serialize(b)
	return b.Bytes()
}