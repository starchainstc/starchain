package node

import (
	"net"
	"time"
	"starchain/net/protocol"
	msg"starchain/net/message"
	."starchain/common/config"
	"starchain/common/log"
	"os"
	"crypto/tls"
	"strconv"
	"io/ioutil"
	"crypto/x509"
	"strings"
	"fmt"
	"errors"
)

type link struct {
	addr 		string
	conn		net.Conn
	port 		uint16
	httpInfoPort 	uint16
	time		time.Time
	rxBuf 		struct{
		p 	[]byte //msg remind data
		len 	int //msg lost length
			     }
	connCnt 	uint64
}

func unpackNodeBuf(node *node,buf []byte){
	var log = log.NewLog()
	var msgLen int
	var msgBuf []byte
	if len(buf) == 0 {
		return
	}
	//a total msg be parsed
	if node.rxBuf.len == 0{
		length := protocol.MSGHDRLEN -len(node.rxBuf.p)
		//data not enough for msg header
		if length > len(buf){
			length = len(buf)
			node.rxBuf.p = append(node.rxBuf.p,buf[0:length]...)
			return
		}
		//fill the msg header
		node.rxBuf.p = append(node.rxBuf.p,buf[0:length]...)
		if msg.ValidMsgHdr(node.rxBuf.p) == false{
			//the invaild msg
			node.rxBuf.p = nil
			node.rxBuf.len = 0
			log.Warn("got error message header")
			return
		}
		node.rxBuf.len = msg.PayloadLen(node.rxBuf.p)
		buf = buf[length:]
	}
	//
	msgLen = node.rxBuf.len
	if len(buf) == msgLen{
		msgBuf = append(node.rxBuf.p,buf[:]...)
		go msg.HandleNodeMsg(node,msgBuf,len(msgBuf))
		node.rxBuf.len = 0
		node.rxBuf.p = nil
	}else if len(buf) < msgLen {
		node.rxBuf.p = append(node.rxBuf.p, buf[:]...)
		node.rxBuf.len = msgLen - len(buf)
	} else {
		msgBuf = append(node.rxBuf.p, buf[0:msgLen]...)
		go msg.HandleNodeMsg(node, msgBuf, len(msgBuf))
		node.rxBuf.p = nil
		node.rxBuf.len = 0

		unpackNodeBuf(node, buf[msgLen:])
	}
}



func printIPAddr() {
	var log = log.NewLog()
	host, _ := os.Hostname()
	addrs, _ := net.LookupIP(host)
	for _, addr := range addrs {
		if ipv4 := addr.To4(); ipv4 != nil {
			log.Info("IPv4: ", ipv4)
		}
	}
}

func (link *link) CloseConn() {
	link.conn.Close()
}

func (n *node) initConnection() {
	var log = log.NewLog()
	isTls := Parameters.IsTLS
	var listener net.Listener
	var err error
	if isTls {
		listener, err = initTlsListen()
		if err != nil {
			log.Error("TLS listen failed")
			return
		}
	} else {
		listener, err = initNonTlsListen()
		if err != nil {
			log.Error("non TLS listen failed")
			return
		}
	}
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Error("Error accepting ", err.Error())
			return
		}
		log.Info("Remote node connect with ", conn.RemoteAddr(), conn.LocalAddr())

		n.link.connCnt++

		node := NewNode()
		node.addr, err = parseIPaddr(conn.RemoteAddr().String())
		node.local = n
		node.conn = conn
		go node.rx()
	}
	//TODO Release the net listen resouce
}

func initNonTlsListen() (net.Listener, error) {
	var log = log.NewLog()
	listener, err := net.Listen("tcp", ":"+strconv.Itoa(Parameters.NodePort))
	if err != nil {
		log.Error("Error listening\n", err.Error())
		return nil, err
	}
	return listener, nil
}

func initTlsListen() (net.Listener, error) {
	var log = log.NewLog()
	CertPath := Parameters.CertPath
	KeyPath := Parameters.KeyPath
	CAPath := Parameters.CAPath

	// load cert
	cert, err := tls.LoadX509KeyPair(CertPath, KeyPath)
	if err != nil {
		log.Error("load keys fail", err)
		return nil, err
	}
	// load root ca
	caData, err := ioutil.ReadFile(CAPath)
	if err != nil {
		log.Error("read ca fail", err)
		return nil, err
	}
	pool := x509.NewCertPool()
	ret := pool.AppendCertsFromPEM(caData)
	if !ret {
		return nil, errors.New("failed to parse root certificate")
	}

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
		RootCAs:      pool,
		ClientAuth:   tls.RequireAndVerifyClientCert,
		ClientCAs:    pool,
	}

	log.Info("TLS listen port is ", strconv.Itoa(Parameters.NodePort))
	listener, err := tls.Listen("tcp", ":"+strconv.Itoa(Parameters.NodePort), tlsConfig)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	return listener, nil
}

func parseIPaddr(s string) (string, error) {
	var log = log.NewLog()
	i := strings.Index(s, ":")
	if i < 0 {
		log.Warn("Split IP address&port error")
		return s, errors.New("Split IP address&port error")
	}
	return s[:i], nil
}

func (node *node) Connect(nodeAddr string) error {
	var log = log.NewLog()
	if node.IsAddrInNbrList(nodeAddr) == true {
		return nil
	}
	if added := node.SetAddrInConnectingList(nodeAddr); added == false {
		return errors.New("node exist in connecting list, cancel")
	}

	isTls := Parameters.IsTLS
	var conn net.Conn
	var err error

	if isTls {
		conn, err = TLSDial(nodeAddr)
		if err != nil {
			node.RemoveAddrInConnectingList(nodeAddr)
			log.Error("TLS connect failed: ", err)
			return err
		}
	} else {
		conn, err = NonTLSDial(nodeAddr)
		if err != nil {
			node.RemoveAddrInConnectingList(nodeAddr)
			log.Error("non TLS connect failed: ", err)
			return err
		}
	}
	node.link.connCnt++
	n := NewNode()
	n.conn = conn
	n.addr, err = parseIPaddr(conn.RemoteAddr().String())
	n.local = node

	log.Info(fmt.Sprintf("Connect node %s connect with %s with %s",
		conn.LocalAddr().String(), conn.RemoteAddr().String(),
		conn.RemoteAddr().Network()))
	go n.rx()

	n.SetState(protocol.HAND)
	buf, _ := msg.NewVersion(node)
	n.Tx(buf)

	return nil
}

func NonTLSDial(nodeAddr string) (net.Conn, error) {
	conn, err := net.DialTimeout("tcp", nodeAddr, time.Second*protocol.DIALTIMEOUT)
	if err != nil {
		return nil, err
	}
	return conn, nil
}

func TLSDial(nodeAddr string) (net.Conn, error) {
	CertPath := Parameters.CertPath
	KeyPath := Parameters.KeyPath
	CAPath := Parameters.CAPath

	clientCertPool := x509.NewCertPool()

	cacert, err := ioutil.ReadFile(CAPath)
	cert, err := tls.LoadX509KeyPair(CertPath, KeyPath)
	if err != nil {
		return nil, err
	}

	ret := clientCertPool.AppendCertsFromPEM(cacert)
	if !ret {
		return nil, errors.New("failed to parse root certificate")
	}

	conf := &tls.Config{
		RootCAs:      clientCertPool,
		Certificates: []tls.Certificate{cert},
	}

	var dialer net.Dialer
	dialer.Timeout = time.Second * protocol.DIALTIMEOUT
	conn, err := tls.DialWithDialer(&dialer, "tcp", nodeAddr, conf)
	if err != nil {
		return nil, err
	}
	return conn, nil
}


