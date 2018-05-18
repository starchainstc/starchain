package dbft

import (
	"io"
	ser"starchain/common/serialization"
)

type PrepareResponse struct {
	msgData   ConsensusMessageData
	Signature []byte
}


func (pres *PrepareResponse) Serialize(w io.Writer) error {
	pres.msgData.Serialize(w)
	w.Write(pres.Signature)
	return nil
}

func (pres *PrepareResponse) Deserialize(r io.Reader) error {
	err := pres.msgData.Deserialize(r)
	if err != nil {
		return err
	}
	// Fixme the 64 should be defined as a unified const
	pres.Signature, err = ser.ReadBytes(r, 64)
	if err != nil {
		return err
	}
	return nil
}

func (pres *PrepareResponse) Type() ConsensusMessageType {
	return pres.ConsensusMessageData().Type
}

func (pres *PrepareResponse) ViewNumber() byte {
	return pres.msgData.ViewNumber
}

func (pres *PrepareResponse) ConsensusMessageData() *ConsensusMessageData {
	return &(pres.msgData)
}