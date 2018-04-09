package transaction

import (
	"io"
	"starchain/common/serialization"
	"bytes"
	."starchain/errors"
	"errors"
)

type TransactionAttributeUsage byte

const (
	Nonce          TransactionAttributeUsage = 0x00
	Script         TransactionAttributeUsage = 0x20
	DescriptionUrl TransactionAttributeUsage = 0x81
	Description    TransactionAttributeUsage = 0x90
)

type TxAttribute struct {
	Usage TransactionAttributeUsage
	Data []byte
	Size uint32
}

func NewTxAttribute(u TransactionAttributeUsage,data []byte) TxAttribute{
	tx := TxAttribute{}
	tx.Usage = u
	tx.Data = data
	tx.Size = uint32(len(data))
	return tx
}

func (tx *TxAttribute)GetSize() uint32{
	if tx.Usage == DescriptionUrl {
		return uint32(2 + len(tx.Data))
	}
	return 0
}

func IsValidAttributeType(usage TransactionAttributeUsage) bool {
	return usage == Nonce || usage == Script ||
		usage == DescriptionUrl || usage == Description
}

func (tx *TxAttribute) Serialize(w io.Writer) error {
	if err := serialization.WriteUint8(w, byte(tx.Usage)); err != nil {
		return NewDetailErr(err, ErrNoCode, "Transaction attribute Usage serialization error.")
	}
	if !IsValidAttributeType(tx.Usage) {
		return NewDetailErr(errors.New("[TxAttribute] error"), ErrNoCode, "Unsupported attribute Description.")
	}
	if err := serialization.WriteVarBytes(w, tx.Data); err != nil {
		return NewDetailErr(err, ErrNoCode, "Transaction attribute Data serialization error.")
	}
	return nil
}

func (tx *TxAttribute) Deserialize(r io.Reader) error {
	val, err := serialization.ReadBytes(r, 1)
	if err != nil {
		return NewDetailErr(err, ErrNoCode, "Transaction attribute Usage deserialization error.")
	}
	tx.Usage = TransactionAttributeUsage(val[0])
	if !IsValidAttributeType(tx.Usage) {
		return NewDetailErr(errors.New("[TxAttribute] error"), ErrNoCode, "Unsupported attribute Description.")
	}
	tx.Data, err = serialization.ReadVarBytes(r)
	if err != nil {
		return NewDetailErr(err, ErrNoCode, "Transaction attribute Data deserialization error.")
	}
	tx.Size = uint32(len(tx.Data))
	return nil

}


func (tx *TxAttribute) ToArray() ([]byte) {
	b := new(bytes.Buffer)
	tx.Serialize(b)
	return b.Bytes()
}


