package message

import (
	"starchain/net/protocol"
	"starchain/core/ledger"
	"bytes"
	"starchain/common/serialization"
	"encoding/binary"
	"crypto/sha256"
	"starchain/common/log"
)

type ping struct {
	msgHdr
	height uint64
}

func NewPingMsg() ([]byte,error){
	var log = log.NewLog()
	var msg ping
	msg.msgHdr.Magic = protocol.NETMAGIC
	copy(msg.msgHdr.CMD[:7],"ping")
	msg.height = uint64(ledger.DefaultLedger.Store.GetHeaderHeight())
	tmpBuffer := bytes.NewBuffer([]byte{})
	serialization.WriteUint64(tmpBuffer, msg.height)
	b := new(bytes.Buffer)
	err := binary.Write(b, binary.LittleEndian, tmpBuffer.Bytes())
	if err != nil {
		log.Error("Binary Write failed at new Msg")
		return nil, err
	}
	s := sha256.Sum256(b.Bytes())
	s2 := s[:]
	s = sha256.Sum256(s2)
	buf := bytes.NewBuffer(s[:4])
	binary.Read(buf, binary.LittleEndian, &(msg.msgHdr.Checksum))
	msg.msgHdr.Length = uint32(len(b.Bytes()))

	m, err := msg.Serialization()
	if err != nil {
		log.Error("Error Convert net message ", err.Error())
		return nil, err
	}
	return m, nil
}

func (msg ping) Verify(buf []byte) error {
	err := msg.msgHdr.Verify(buf)
	// TODO verify the message Content
	return err
}

func (msg ping) Handle(node protocol.Noder) error {
	var log = log.NewLog()
	node.SetHeight(msg.height)
	buf, err := NewPongMsg()
	if err != nil {
		log.Error("failed build a new ping message")
	} else {
		go node.Tx(buf)
	}
	return err
}

func (msg ping) Serialization() ([]byte, error) {
	hdrBuf, err := msg.msgHdr.Serialization()
	if err != nil {
		return nil, err
	}
	buf := bytes.NewBuffer(hdrBuf)
	err = serialization.WriteUint64(buf, msg.height)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), err

}

func (msg *ping) Deserialization(p []byte) error {
	buf := bytes.NewBuffer(p)
	err := binary.Read(buf, binary.LittleEndian, &(msg.msgHdr))
	if err != nil {
		return err
	}

	msg.height, err = serialization.ReadUint64(buf)
	return err
}
