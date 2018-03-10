package transaction

import (
	"starchain/common"
	"io"
	"starchain/common/serialization"
	"bytes"
	"fmt"
)

type UTXOTxInput struct{
	ReferTxID	common.Uint256

	//the index of transcation
	ReferTxOutputIndex uint16
}

func (ui *UTXOTxInput) Serialize(w io.Writer) {
	ui.ReferTxID.Serialize(w)
	serialization.WriteUint16(w, ui.ReferTxOutputIndex)
}

func (ui *UTXOTxInput) Deserialize(r io.Reader) error {
	//referTxID
	err := ui.ReferTxID.Deserialize(r)
	if err != nil {
		return err
	}

	//Output Index
	temp, err := serialization.ReadUint16(r)
	ui.ReferTxOutputIndex = uint16(temp)
	if err != nil {
		return err
	}

	return nil
}

func (ui *UTXOTxInput) ToArray() ([]byte) {
	b := new(bytes.Buffer)
	ui.Serialize(b)
	return b.Bytes()
}

func (ui *UTXOTxInput) ToString() string {
	return fmt.Sprintf("%x%x", ui.ReferTxID.ToString(), ui.ReferTxOutputIndex)
}

func (ui *UTXOTxInput) Equals(other *UTXOTxInput) bool {
	if ui == other {
		return true
	}
	if other == nil {
		return false
	}
	if ui.ReferTxID == other.ReferTxID && ui.ReferTxOutputIndex == other.ReferTxOutputIndex {
		return true
	} else {
		return false
	}
}


