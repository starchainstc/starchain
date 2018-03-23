package asset

import (
	"io"
	"starchain/common/serialization"
	"bytes"
	"errors"
	."starchain/errors"
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
	w.Write([]byte{byte(asset.Precision)})
	//err = serialization.WriteByte(w,asset.Precision)
	//if err != nil {
	//	return err
	//}
	w.Write([]byte{byte(asset.AssetType)})
	//err = serialization.WriteByte(w,asset.AssetType)
	//if err != nil {
	//	return err
	//}
	w.Write([]byte{byte(asset.RecordType)})
	//err = serialization.WriteByte(w,asset.RecordType)
	//if err != nil {
	//	return err
	//}
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
	p := make([]byte,1)
	n,err :=r.Read(p)
	if n > 0 {
		asset.Precision = p[0]
	}else{
		return NewDetailErr(errors.New("deserical asset precision error"),ErrNoCode,"")
	}
	n,err =r.Read(p)
	if n > 0 {
		asset.AssetType = AssetType(p[0])
	}else{
		return NewDetailErr(errors.New("deserical asset assetType error"),ErrNoCode,"")
	}
	n,err =r.Read(p)
	if n > 0 {
		asset.RecordType = AssetRecordType(p[0])
	}else{
		return NewDetailErr(errors.New("deserical asset recordtype error"),ErrNoCode,"")
	}
	return err
}

func (asset *Asset) ToArray() []byte {
	b := new(bytes.Buffer)
	asset.Serialize(b)
	return b.Bytes()
}