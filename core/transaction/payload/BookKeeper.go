package payload

import (
	"starchain/crypto"
	"bytes"
	"starchain/common/serialization"
	"io"
	."starchain/errors"
)

const BookKeeperPayloadVersion byte = 0x00

type BookKeeperAction byte
const (
	BookKeeperAction_ADD BookKeeperAction = 0
	BookKeeperAction_SUB BookKeeperAction = 1
)

type BookKeeper struct {
	PubKey *crypto.PubKey
	Action BookKeeperAction
	Cert 	[]byte
	Issuer *crypto.PubKey
}

func (bk *BookKeeper) Data(version byte) []byte{
	buf := new(bytes.Buffer)

	bk.PubKey.Serialize(buf)
	buf.WriteByte(byte(bk.Action))
	serialization.WriteVarBytes(buf,bk.Cert)
	bk.Issuer.Serialize(buf)

	return buf.Bytes()
}

func (bk *BookKeeper) Serialize(w io.Writer,version byte) error{
	_,err := w.Write(bk.Data(version))
	return err
}

func (self *BookKeeper) Deserialize(r io.Reader, version byte) error {
	self.PubKey = new(crypto.PubKey)
	err := self.PubKey.DeSerialize(r)
	if err != nil {
		return NewDetailErr(err, ErrNoCode, "[BookKeeper], PubKey Deserialize failed.")
	}
	var p [1]byte
	n, err := r.Read(p[:])
	if n == 0 {
		return NewDetailErr(err, ErrNoCode, "[BookKeeper], Action Deserialize failed.")
	}
	self.Action = BookKeeperAction(p[0])
	self.Cert, err = serialization.ReadVarBytes(r)
	if err != nil {
		return NewDetailErr(err, ErrNoCode, "[BookKeeper], Cert Deserialize failed.")
	}
	self.Issuer = new(crypto.PubKey)
	err = self.Issuer.DeSerialize(r)
	if err != nil {
		return NewDetailErr(err, ErrNoCode, "[BookKeeper], Issuer Deserialize failed.")
	}

	return nil
}