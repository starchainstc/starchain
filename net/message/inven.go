package message

import (
	"starchain/net/protocol"
	."starchain/common"
	"starchain/core/ledger"
	"bytes"
	"encoding/binary"
	"starchain/common/log"
	"crypto/sha256"
	"io"
	"starchain/common/serialization"
	"encoding/hex"
	"fmt"
)

type blocksReq struct {
	msgHdr
	p struct {
		  HeaderHashCount uint8
		  hashStart       [protocol.HASHLEN]byte
		  hashStop        [protocol.HASHLEN]byte
	  }
}

type InvPayload struct {
	InvType InventoryType
	Cnt     uint32
	Blk     []byte
}

type Inv struct {
	Hdr msgHdr
	P   InvPayload
}


func NewBlocksReq(n protocol.Noder) ([]byte, error) {
	var log = log.NewLog()
	var h blocksReq
	log.Debug("request block hash")
	// Fixme correct with the exactly request length
	h.p.HeaderHashCount = 1
	//Fixme! Should get the remote Node height.
	buf := ledger.DefaultLedger.Blockchain.CurrentBlockHash()

	copy(h.p.hashStart[:], reverse(buf[:]))

	p := new(bytes.Buffer)
	err := binary.Write(p, binary.LittleEndian, &(h.p))
	if err != nil {
		log.Error("Binary Write failed at new blocksReq")
		return nil, err
	}

	s := checkSum(p.Bytes())
	h.msgHdr.init("getblocks", s, uint32(len(p.Bytes())))

	m, err := h.Serialization()

	return m, err
}

func (msg blocksReq) Verify(buf []byte) error {

	// TODO verify the message Content
	err := msg.msgHdr.Verify(buf)
	return err
}

func (msg blocksReq) Handle(node protocol.Noder) error {
	var log = log.NewLog()
	log.Debug("handle blocks request")
	var starthash Uint256
	var stophash Uint256
	starthash = msg.p.hashStart
	stophash = msg.p.hashStop
	//FIXME if HeaderHashCount > 1
	inv, err := GetInvFromBlockHash(starthash, stophash)
	if err != nil {
		return err
	}
	buf, err := NewInv(inv)
	if err != nil {
		return err
	}
	go node.Tx(buf)
	return nil
}


func (msg blocksReq) Serialization() ([]byte, error) {
	var buf bytes.Buffer

	err := binary.Write(&buf, binary.LittleEndian, msg)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), err
}

func (msg *blocksReq) Deserialization(p []byte) error {
	buf := bytes.NewBuffer(p)
	err := binary.Read(buf, binary.LittleEndian, &msg)
	return err
}

func (msg Inv) Verify(buf []byte) error {
	// TODO verify the message Content
	err := msg.Hdr.Verify(buf)
	return err
}

func (msg Inv) Handle(node protocol.Noder) error {
	var log = log.NewLog()
	var id Uint256
	str := hex.EncodeToString(msg.P.Blk)
	log.Debug(fmt.Sprintf("The inv type: 0x%x block len: %d, %s\n",
		msg.P.InvType, len(msg.P.Blk), str))

	invType := InventoryType(msg.P.InvType)
	switch invType {
	case TRANSACTION:
		log.Debug("RX TRX message")
		// TODO check the ID queue
		id.Deserialize(bytes.NewReader(msg.P.Blk[:32]))
		if !node.ExistedID(id) {
			reqTxnData(node, id)
		}
	case BLOCK:
		log.Debug("RX block message")
		var i uint32
		count := msg.P.Cnt
		log.Debug("RX inv-block message, hash is ", msg.P.Blk)
		for i = 0; i < count; i++ {
			id.Deserialize(bytes.NewReader(msg.P.Blk[protocol.HASHLEN*i:]))
			// TODO check the ID queue
			if !ledger.DefaultLedger.Store.BlockInCache(id) &&
				!ledger.DefaultLedger.BlockInLedger(id) {
				node.CacheHash(id) //cached hash would not relayed
				if !node.LocalNode().ExistedID(id) {
					// send the block request
					//log.Infof("inv request block hash: %x", id)
					ReqBlkData(node, id)
				}

			}

		}
	case CONSENSUS:
		log.Debug("RX consensus message")
		id.Deserialize(bytes.NewReader(msg.P.Blk[:32]))
		reqConsensusData(node, id)
	default:
		log.Warn("RX unknown inventory message")
	}
	return nil
}

func (msg Inv) Serialization() ([]byte, error) {
	hdrBuf, err := msg.Hdr.Serialization()
	if err != nil {
		return nil, err
	}
	buf := bytes.NewBuffer(hdrBuf)
	msg.P.Serialization(buf)

	return buf.Bytes(), err
}

func (msg *Inv) Deserialization(p []byte) error {
	err := msg.Hdr.Deserialization(p)
	if err != nil {
		return err
	}

	buf := bytes.NewBuffer(p[protocol.MSGHDRLEN:])
	invType, err := serialization.ReadUint8(buf)
	if err != nil {
		return err
	}
	msg.P.InvType = InventoryType(invType)
	msg.P.Cnt, err = serialization.ReadUint32(buf)
	if err != nil {
		return err
	}

	msg.P.Blk = make([]byte, msg.P.Cnt*protocol.HASHLEN)
	err = binary.Read(buf, binary.LittleEndian, &(msg.P.Blk))

	return err
}

func (msg Inv) invType() InventoryType {
	return msg.P.InvType
}

func GetInvFromBlockHash(starthash Uint256, stophash Uint256) (*InvPayload, error) {
	var log = log.NewLog()
	var count uint32 = 0
	var i uint32
	var empty Uint256
	var startheight uint32
	var stopheight uint32
	curHeight := ledger.DefaultLedger.GetLocalBlockChainHeight()
	if starthash == empty {
		if stophash == empty {
			if curHeight > protocol.MAXBLKHDRCNT {
				count = protocol.MAXBLKHDRCNT
			} else {
				count = curHeight
			}
		} else {
			bkstop, err := ledger.DefaultLedger.Store.GetHeader(stophash)
			if err != nil {
				return nil, err
			}
			stopheight = bkstop.Blockdata.Height
			count = curHeight - stopheight
			if curHeight > protocol.MAXINVHDRCNT {
				count = protocol.MAXINVHDRCNT
			}
		}
	} else {
		bkstart, err := ledger.DefaultLedger.Store.GetHeader(starthash)
		if err != nil {
			return nil, err
		}
		startheight = bkstart.Blockdata.Height
		if stophash != empty {
			bkstop, err := ledger.DefaultLedger.Store.GetHeader(stophash)
			if err != nil {
				return nil, err
			}
			stopheight = bkstop.Blockdata.Height
			count = startheight - stopheight
			if count >= protocol.MAXINVHDRCNT {
				count = protocol.MAXINVHDRCNT
				stopheight = startheight + protocol.MAXINVHDRCNT
			}
		} else {

			if startheight > protocol.MAXINVHDRCNT {
				count = protocol.MAXINVHDRCNT
			} else {
				count = startheight
			}
		}
	}
	tmpBuffer := bytes.NewBuffer([]byte{})
	for i = 1; i <= count; i++ {
		//FIXME need add error handle for GetBlockWithHash
		hash, _ := ledger.DefaultLedger.Store.GetBlockHash(stopheight + i)
		log.Debug("GetInvFromBlockHash i is ", i, " , hash is ", hash)
		hash.Serialize(tmpBuffer)
	}
	log.Debug("GetInvFromBlockHash hash is ", tmpBuffer.Bytes())
	return NewInvPayload(BLOCK, count, tmpBuffer.Bytes()), nil
}

func NewInvPayload(invType InventoryType, count uint32, msg []byte) *InvPayload {
	return &InvPayload{
		InvType: invType,
		Cnt:     count,
		Blk:     msg,
	}
}

func NewInv(inv *InvPayload) ([]byte, error) {
	var log = log.NewLog()
	var msg Inv

	msg.P.Blk = inv.Blk
	msg.P.InvType = inv.InvType
	msg.P.Cnt = inv.Cnt
	msg.Hdr.Magic = protocol.NETMAGIC
	cmd := "inv"
	copy(msg.Hdr.CMD[0:len(cmd)], cmd)
	tmpBuffer := bytes.NewBuffer([]byte{})
	inv.Serialization(tmpBuffer)

	b := new(bytes.Buffer)
	err := binary.Write(b, binary.LittleEndian, tmpBuffer.Bytes())
	if err != nil {
		log.Error("Binary Write failed at new Msg", err.Error())
		return nil, err
	}
	s := sha256.Sum256(b.Bytes())
	s2 := s[:]
	s = sha256.Sum256(s2)
	buf := bytes.NewBuffer(s[:4])
	binary.Read(buf, binary.LittleEndian, &(msg.Hdr.Checksum))
	msg.Hdr.Length = uint32(len(b.Bytes()))

	m, err := msg.Serialization()
	if err != nil {
		log.Error("Error Convert net message ", err.Error())
		return nil, err
	}

	return m, nil
}

func (msg *InvPayload) Serialization(w io.Writer) {
	serialization.WriteUint8(w, uint8(msg.InvType))
	serialization.WriteUint32(w, msg.Cnt)

	binary.Write(w, binary.LittleEndian, msg.Blk)
}


