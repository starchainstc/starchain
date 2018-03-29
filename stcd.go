package main

import (
	"starchain/common/log"
	"starchain/account"
	"starchain/crypto"
	"os"
	"starchain/core/transaction"
	"starchain/common/config"
	"starchain/core/store/ChainStore"
	"starchain/net"
	"starchain/net/rpchttp"
	"time"
	"starchain/consensus/dbft"
	"starchain/net/httprestful"
	"starchain/core/ledger"
	"starchain/net/protocol"
)

func main(){
	var err error
	ledger.DefaultLedger = new(ledger.Ledger)
	ledger.DefaultLedger.Store,err = ChainStore.NewLedgerStore()
	defer ledger.DefaultLedger.Store.Close()
	if err != nil {
		log.Fatal("open LedgerStore err:", err)
		os.Exit(1)
	}
	ledger.DefaultLedger.Store.InitLedgerStore(ledger.DefaultLedger)
	transaction.TxStore = ledger.DefaultLedger.Store
	crypto.SetAlg(config.Parameters.EncryptAlg)
	ledger.StandbyBookKeepers = account.GetBookKeepers()
	chain, err := ledger.GenesisBlock(ledger.StandbyBookKeepers)
	checkErr(err,"generate blockchain failed")
	ledger.DefaultLedger.Blockchain = chain
	log.Info("get client")
	cli := account.GetClient()
	if cli == nil {
		log.Fatal("Can't get local account.")
		os.Exit(1)
	}
	acc, err := cli.GetDefaultAccount()
	checkErr(err,"can't get main-account")
	rpchttp.Wallet = cli
	node := net.StartProtocol(acc.PublicKey)
	rpchttp.RegistRpcNode(node)
	time.Sleep(6 * time.Second)
	log.Info("start sync block")
	node.SyncNodeHeight()
	log.Info("sync block finish")
	node.WaitForFourPeersStart()
	node.WaitForSyncBlkFinish()
	log.Info("--Start the RPC interface")
	go rpchttp.StartRPCServer()
	log.Info("start http server")
	go httprestful.StartServer(node)
	if protocol.VERIFYNODENAME == config.Parameters.NodeType {
		dbftServices := dbft.NewDbftService(cli, "logcon", node)
		rpchttp.RegistDbftService(dbftServices)
		go dbftServices.Start()
		time.Sleep(8 * time.Second)
	}
	for {
		time.Sleep(dbft.GenBlockTime)
		log.Info("BlockHeight = ", ledger.DefaultLedger.Blockchain.BlockHeight)
		isNeedNewFile := log.CheckIfNeedNewFile()
		if isNeedNewFile == true {
			log.ClosePrintLog()
			log.Init(log.Path, os.Stdout)
		}
	}
}


func checkErr(err error,msg string){
	if err != nil {
		if msg == ""{
			log.Error(err)
		}
		log.Error(msg)
		os.Exit(1)
	}
}