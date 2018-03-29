package message

import (
	"starchain/crypto"
	"starchain/net/protocol"
	"starchain/common/config"
	"time"
	"starchain/core/ledger"
	"bytes"
	"encoding/binary"
	"crypto/sha256"
	"errors"
	"fmt"
	"starchain/common/log"
)

const (
	HTTPINFOFLAG = 0
)


type version struct {
	Hdr	msgHdr
	P 	struct{
			 Version      uint32
			 Services     uint64
			 TimeStamp    uint32
			 Port         uint16
			 HttpInfoPort uint16
			 Cap          [32]byte
			 Nonce        uint64
			 //remove tempory to get serilization function passed
			 UserAgent   uint8
			 StartHeight uint64
			 // FIXME check with the specify relay type length
			 Relay uint8
		 }
	pk *crypto.PubKey
}


func NewVersion(node protocol.Noder)([]byte,error){
	var msg version
	msg.P.Version = node.Version()
	msg.P.Services = node.Services()
	msg.P.HttpInfoPort = config.Parameters.HttpInfoPort
	if config.Parameters.HttpInfoStart {
		msg.P.Cap[HTTPINFOFLAG] = 0x01
	} else {
		msg.P.Cap[HTTPINFOFLAG] = 0x00
	}
	msg.P.TimeStamp = uint32(time.Now().UTC().UnixNano())
	msg.P.Port = node.GetPort()
	msg.P.Nonce = node.GetID()
	msg.P.UserAgent = 0x00
	msg.P.StartHeight = uint64(ledger.DefaultLedger.GetLocalBlockChainHeight())
	if node.GetRelay() {
		msg.P.Relay = 1
	}else {
		msg.P.Relay = 0
	}

	msg.pk = node.GetBookKeeperAddr()
	msg.Hdr.Magic = protocol.NETMAGIC
	copy(msg.Hdr.CMD[:7],"version")
	p := bytes.NewBuffer([]byte{})
	err := binary.Write(p,binary.LittleEndian,&(msg.P))
	msg.pk.Serialize(p)
	if err != nil {
		log.Error("net version serialize fail")
		return nil,err
	}
	s:= sha256.Sum256(p.Bytes())
	s2 := s[:]
	s = sha256.Sum256(s2)
	buf := bytes.NewBuffer(s[:4])
	binary.Read(buf,binary.LittleEndian,&(msg.Hdr.Checksum))
	msg.Hdr.Length = uint32(len(p.Bytes()))
	log.Debug("The message payload length is ", msg.Hdr.Length)
	m,err := msg.Serialization()
	if err!= nil {
		return nil,err
	}
	return m,nil
}

func (msg version)Verify(buf []byte) error{
	err := msg.Hdr.Verify(buf)
	return err
}

func (msg version)Handle(node protocol.Noder) error{
	localNode := node.LocalNode()
	if msg.P.Nonce == localNode.GetID(){
		log.Warn("the node handshark with itself")
		node.CloseConn()
		return errors.New("handsharek self")
	}
	s := node.GetState()
	if s != protocol.INIT && s != protocol.HAND{
		log.Error("node state is unknow")
		return errors.New("unknow node state")
	}
	n,ret := localNode.DelNbrNode(msg.P.Nonce)
	if ret {
		log.Info(fmt.Sprintf("Node reconnect 0x%x", msg.P.Nonce))
		n.SetState(protocol.INACTIVITY)
		n.CloseConn()
	}
	log.Debug("handle version msg.pk is ", msg.pk)

	if msg.P.Cap[HTTPINFOFLAG] == 0x01 {
		node.SetHttpInfoState(true)
	} else {
		node.SetHttpInfoState(false)
	}
	node.SetHttpInfoPort(msg.P.HttpInfoPort)
	node.SetBookKeeperAddr(msg.pk)
	node.UpdateInfo(time.Now(), msg.P.Version, msg.P.Services,
		msg.P.Port, msg.P.Nonce, msg.P.Relay, msg.P.StartHeight)
	localNode.AddNbrNode(node)

	var buf []byte
	if s == protocol.INIT {
		node.SetState(protocol.HANDSHAKE)
		buf, _ = NewVersion(localNode)
	} else if s == protocol.HAND {
		node.SetState(protocol.HANDSHAKED)
		buf, _ = NewVerack()
	}
	node.Tx(buf)

	return nil
}


func (msg version) Serialization() ([]byte, error) {
	hdrBuf, err := msg.Hdr.Serialization()
	if err != nil {
		return nil, err
	}
	buf := bytes.NewBuffer(hdrBuf)
	err = binary.Write(buf, binary.LittleEndian, msg.P)
	if err != nil {
		return nil, err
	}
	msg.pk.Serialize(buf)

	return buf.Bytes(), err
}

func (msg *version) Deserialization(p []byte) error {
	buf := bytes.NewBuffer(p)

	err := binary.Read(buf, binary.LittleEndian, &(msg.Hdr))
	if err != nil {
		log.Warn("Parse version message hdr error")
		return errors.New("Parse version message hdr error")
	}

	err = binary.Read(buf, binary.LittleEndian, &(msg.P))
	if err != nil {
		log.Warn("Parse version P message error")
		return errors.New("Parse version P message error")
	}

	pk := new(crypto.PubKey)
	err = pk.DeSerialize(buf)
	if err != nil {
		return errors.New("Parse pubkey Deserialize failed.")
	}
	msg.pk = pk
	return err
}
