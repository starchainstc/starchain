package asset

import (
	"io"
	"starchain/common/serialization"
	"bytes"
)

type AssetType byte
const (
	Currency AssetType = 0x00
	Share 	 AssetType = 0x01
	Invoice  AssetType = 0x10
	Token	 AssetType = 0x11
)


const (
	MaxPrecision = 8
	MinPrecision = 0
)

type AssetRecordType byte

//starchin STC is planed to support UTXO and Balance
const (
	UTXO    AssetRecordType = 0x00
	Balance AssetRecordType = 0x01
)

type Asset struct {
	Name string
	Description string
	Precision byte
	AssetType AssetType
	RecordType AssetRecordType
}

func (asset *Asset) Serialize(w io.Writer) error{
	err := serialization.WriteVarString(w,asset.Name)
	if err != nil {
		return err
	}
	err = serialization.WriteVarString(w,asset.Description)
	if err != nil {
		return err
	}
	err = serialization.WriteBool(w,asset.Precision)
	if err != nil {
		return err
	}
	err = serialization.WriteBool(w,asset.AssetType)
	if err != nil {
		return err
	}
	err = serialization.WriteBool(w,asset.RecordType)
	if err != nil {
		return err
	}
	return nil
}

func (asset *Asset) Deserialize(r io.Reader) error{
	name,err := serialization.ReadVarString(r)
	if err != nil{
		return err
	}
	asset.Name = name
	desc,err := serialization.ReadVarString(r)
	if err != nil {
		return err
	}
	asset.Description = desc
	asset.Precision,err = serialization.ReadByte(r)
	if err != nil {
		return err
	}
	asset.AssetType ,err = serialization.ReadByte(r)
	if err != nil {
		return err
	}
	asset.RecordType,err = serialization.ReadByte(r)
	if err != nil {
		return err
	}
	return err
}

func (asset *Asset) ToArray() []byte {
	b := new(bytes.Buffer)
	asset.Serialize(b)
	return b.Bytes()
}