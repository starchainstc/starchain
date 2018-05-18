package dbft

import (
	"starchain/common"
	"io"
	tx"starchain/core/transaction"
	ser"starchain/common/serialization"
	."starchain/errors"
)

type PrepareRequest struct {
	msgData        ConsensusMessageData
	Nonce          uint64
	NextBookKeeper common.Uint160
	Transactions   []*tx.Transaction
	Signature      []byte
}


func (pr *PrepareRequest) Serialize(w io.Writer) error {

	pr.msgData.Serialize(w)
	if err := ser.WriteVarUint(w, pr.Nonce); err != nil {
		return NewDetailErr(err, ErrNoCode, "[PrepareRequest] nonce serialization failed")
	}
	if _, err := pr.NextBookKeeper.Serialize(w); err != nil {
		return NewDetailErr(err, ErrNoCode, "[PrepareRequest] nextbookKeeper serialization failed")
	}
	if err := ser.WriteVarUint(w, uint64(len(pr.Transactions))); err != nil {
		return NewDetailErr(err, ErrNoCode, "[PrepareRequest] length serialization failed")
	}
	for _, t := range pr.Transactions {
		if err := t.Serialize(w); err != nil {
			return NewDetailErr(err, ErrNoCode, "[PrepareRequest] transactions serialization failed")
		}
	}
	if err := ser.WriteVarBytes(w, pr.Signature); err != nil {
		return NewDetailErr(err, ErrNoCode, "[PrepareRequest] signature serialization failed")
	}
	return nil
}

func (pr *PrepareRequest) Deserialize(r io.Reader) error {
	pr.msgData = ConsensusMessageData{}
	pr.msgData.Deserialize(r)
	pr.Nonce, _ = ser.ReadVarUint(r, 0)

	pr.NextBookKeeper = common.Uint160{}
	if err := pr.NextBookKeeper.Deserialize(r); err != nil {
		return NewDetailErr(err, ErrNoCode, "[PrepareRequest] nextbookKeeper deserialization failed")
	}

	length, err := ser.ReadVarUint(r, 0)
	if err != nil {
		return NewDetailErr(err, ErrNoCode, "[PrepareRequest] length deserialization failed")
	}

	pr.Transactions = make([]*tx.Transaction, length)
	for i := 0; i < len(pr.Transactions); i++ {
		var t tx.Transaction
		if err := t.Deserialize(r); err != nil {
			return NewDetailErr(err, ErrNoCode, "[PrepareRequest] transactions deserialization failed")
		}
		pr.Transactions[i] = &t
	}

	pr.Signature, err = ser.ReadVarBytes(r)
	if err != nil {
		return NewDetailErr(err, ErrNoCode, "[PrepareRequest] signature deserialization failed")
	}

	return nil
}

func (pr *PrepareRequest) Type() ConsensusMessageType {
	return pr.ConsensusMessageData().Type
}

func (pr *PrepareRequest) ViewNumber() byte {
	return pr.msgData.ViewNumber
}

func (pr *PrepareRequest) ConsensusMessageData() *ConsensusMessageData {
	return &(pr.msgData)
}