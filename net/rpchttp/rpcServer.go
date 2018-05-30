package rpchttp

import (
	"net/http"
	"strconv"
	"starchain/common/config"
	"starchain/common/log"
)

const (
	LocalHost = "127.0.0.1"
)

func StartRPCServer() {
	var log = log.NewLog()
	http.HandleFunc("/",Handle)
	HandleFunc("/getbestblockhash",getBestBlockHash)
	HandleFunc("getblock", getBlock)
	HandleFunc("getblockcount", getBlockCount)
	HandleFunc("getblockhash", getBlockHash)
	HandleFunc("getconnectioncount", getConnectionCount)
	HandleFunc("getrawmempool", getRawMemPool)
	HandleFunc("getrawtransaction", getRawTransaction)
	HandleFunc("sendrawtransaction", sendRawTransaction)
	HandleFunc("getversion", getVersion)
	HandleFunc("getneighbor", getNeighbor)
	HandleFunc("getnodestate", getNodeState)

	HandleFunc("setdebuginfo", setDebugInfo)
	HandleFunc("sendtoaddress", sendToAddress)
	HandleFunc("lockasset", lockAsset)
	HandleFunc("createmultisigtransaction", createMultisigTransaction)
	HandleFunc("signmultisigtransaction", signMultisigTransaction)

	err := http.ListenAndServe(LocalHost+":"+strconv.Itoa(config.Parameters.HttpJsonPort), nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err.Error())
	}
}
