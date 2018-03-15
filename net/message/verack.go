package message

import (
	"encoding/hex"
	"starchain/common/log"
	"strconv"
	."starchain/net/protocol"
	"errors"
)

type verACK struct {
	msgHdr
	// No payload
}


func NewVerack() ([]byte, error) {
	var msg verACK
	// Fixme the check is the []byte{0} instead of 0
	var sum []byte
	sum = []byte{0x5d, 0xf6, 0xe0, 0xe2}
	msg.msgHdr.init("verack", sum, 0)

	buf, err := msg.Serialization()
	if err != nil {
		return nil, err
	}

	str := hex.EncodeToString(buf)
	log.Debug("The message tx verack length is ", len(buf), ", ", str)

	return buf, err
}

func (msg verACK) Handle(node Noder) error {
	log.Debug()

	s := node.GetState()
	if s != HANDSHAKE && s != HANDSHAKED {
		log.Warn("Unknow status to received verack")
		return errors.New("Unknow status to received verack")
	}

	node.SetState(ESTABLISH)

	if s == HANDSHAKE {
		buf, _ := NewVerack()
		node.Tx(buf)
	}

	node.DumpInfo()
	// Fixme, there is a race condition here,
	// but it doesn't matter to access the invalid
	// node which will trigger a warning
	node.ReqNeighborList()
	addr := node.GetAddr()
	port := node.GetPort()
	nodeAddr := addr + ":" + strconv.Itoa(int(port))
	node.LocalNode().RemoveAddrInConnectingList(nodeAddr)
	return nil
}
