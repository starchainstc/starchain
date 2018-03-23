package ledger

import (
	. "starchain/common"
	"starchain/core/contract/program"
	"io"
	"starchain/common/serialization"
	."starchain/errors"
	sig"starchain/core/signature"
	"errors"
	"crypto/sha256"
	"bytes"
)

type Blockdata struct {
	Version		uint32
	PrevBlockHash	Uint256
	TransactionsRoot	Uint256
	Timestamp	uint32
	Height		uint32
	ConsensusData	uint64
	NextBookKeeper	Uint160
	hash		Uint256
	Program		*program.Program

}
func (bd *Blockdata) Serialize(w io.Writer) error{
	bd.SerializeUnsigned(w)
	w.Write([]byte{byte(1)})
	if bd.Program != nil{
		bd.Program.Serialize(w)
	}
	return nil
}



//don't write the hash and program to writer
func (bd *Blockdata) SerializeUnsigned(w io.Writer) error{
	serialization.WriteUint32(w,bd.Version)
	bd.PrevBlockHash.Serialize(w)
	bd.TransactionsRoot.Serialize(w)
	serialization.WriteUint32(w,bd.Timestamp)
	serialization.WriteUint32(w,bd.Height)
	serialization.WriteUint64(w,bd.ConsensusData)
	bd.NextBookKeeper.Serialize(w)
	return nil

}

func (bd *Blockdata) Deserialize(r io.Reader) error{
	bd.DeserializeUnsigned(r)
	p:=make([]byte,1)
	n,err := r.Read(p)
	if n > 0 {
		x := []byte(p[:])
		if x[0] != 1{
			return NewDetailErr(errors.New("BlockData Deserializ format error"),ErrNoCode,"")
		}
	}else{
		return NewDetailErr(errors.New("BlockData Deserializ format error"),ErrNoCode,"")
	}
	pg := new(program.Program)
	err = pg.Deserialize(r)
	if err != nil {
		return err
	}
	bd.Program = pg
	return nil

}


func (bd *Blockdata) DeserializeUnsigned(r io.Reader) error{
	v,err:= serialization.ReadUint32(r)
	if err != nil {
		return err
	}
	bd.Version = v
	preBlock := new(Uint256)
	err = preBlock.Deserialize(r)
	if err != nil{
		return nil
	}
	bd.PrevBlockHash = *preBlock
	root := new(Uint256)
	err = root.Deserialize(r)
	if err != nil {
		return nil
	}
	bd.TransactionsRoot = *root

	bd.Timestamp,err = serialization.ReadUint32(r)
	bd.Height,err = serialization.ReadUint32(r)
	bd.ConsensusData,err = serialization.ReadUint64(r)
	if err != nil {
		return nil
	}
	bd.NextBookKeeper.Deserialize(r)
	return nil

}

func (bd *Blockdata) GetProgramHashes() ([]Uint160,error){
	programHash := []Uint160{}
	temp := Uint256{}
	if bd.PrevBlockHash == temp {
		pg := *bd.Program
		outputhash,err := ToCodeHash(pg.Code)
		if err != nil {
			return nil,err
		}
		programHash = append(programHash,outputhash)
		return programHash,nil
	}else{
		prev_header, err := DefaultLedger.Store.GetHeader(bd.PrevBlockHash)
		if err != nil {
			return programHash, err
		}
		programHash = append(programHash, prev_header.Blockdata.NextBookKeeper)
		return programHash, nil
	}
}

func (bd *Blockdata) SetPrograms(programs []*program.Program) {
	if len(programs) != 1 {
		return
	}
	bd.Program = programs[0]
}

func (bd *Blockdata) GetPrograms() []*program.Program {
	return []*program.Program{bd.Program}
}

func (bd *Blockdata) Hash() Uint256 {

	d := sig.GetHashData(bd)
	temp := sha256.Sum256([]byte(d))
	f := sha256.Sum256(temp[:])
	hash := Uint256(f)
	return hash
}

func (bd *Blockdata) GetMessage() []byte {
	return sig.GetHashData(bd)
}

func (bd *Blockdata) ToArray() ([]byte) {
	b := new(bytes.Buffer)
	bd.Serialize(b)
	return b.Bytes()
}

