package states

import (
	"starchain/common"
	"io"
	"starchain/common/serialization"
	."starchain/errors"
)

type StorageKey struct {
	CodeHash common.Uint160
	Key []byte
}

func NewStorageKey(codehash *common.Uint160,key []byte) *StorageKey{
	var storagekey StorageKey
	storagekey.CodeHash = *codehash
	storagekey.Key = key
	return &storagekey
}



func (storageKey *StorageKey) Serialize(w io.Writer) (int, error) {
	storageKey.CodeHash.Serialize(w)
	serialization.WriteVarBytes(w, storageKey.Key)
	return 0, nil
}

func (storageKey *StorageKey) Deserialize(r io.Reader) error {
	u := new(common.Uint160)
	err := u.Deserialize(r)
	if err != nil {
		return NewDetailErr(err, ErrNoCode, "StorageKey CodeHash Deserialize fail.")
	}
	storageKey.CodeHash = *u
	key, err := serialization.ReadVarBytes(r)
	if err != nil {
		return NewDetailErr(err, ErrNoCode, "StorageKey Key Deserialize fail.")
	}
	storageKey.Key = key
	return nil
}