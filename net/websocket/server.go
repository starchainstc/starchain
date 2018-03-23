package websocket

import (
	"starchain/net/websocket/websocket"
	"bytes"
	"starchain/net/httprestful/common"
	"starchain/core/ledger"
	"starchain/net/httprestful/ErrorCode"
	."starchain/common/config"
	"starchain/events"
	"starchain/net/protocol"
	."starchain/common"
	"starchain/errors"
)

var ws *websocket.WsServer
var (
	pushBlockFlag    bool = false
	pushRawBlockFlag bool = false
	pushBlockTxsFlag bool = false
)


func StartServer(n protocol.Noder) {
	common.SetNode(n)
	ledger.DefaultLedger.Blockchain.BCEvents.Subscribe(events.EventBlockPersistCompleted, SendBlock2WSclient)
	go func() {
		ws = websocket.InitWsServer(common.CheckAccessToken)
		ws.Start()
	}()
}
func SendBlock2WSclient(v interface{}) {
	if Parameters.HttpWsPort != 0 && pushBlockFlag {
		go func() {
			PushBlock(v)
		}()
	}
	if Parameters.HttpWsPort != 0 && pushBlockTxsFlag {
		go func() {
			PushBlockTransactions(v)
		}()
	}
}
func Stop() {
	if ws == nil {
		return
	}
	ws.Stop()
}
func ReStartServer() {
	if ws == nil {
		ws = websocket.InitWsServer(common.CheckAccessToken)
		ws.Start()
		return
	}
	ws.Restart()
}
func GetWsPushBlockFlag() bool {
	return pushBlockFlag
}
func SetWsPushBlockFlag(b bool) {
	pushBlockFlag = b
}
func GetPushRawBlockFlag() bool {
	return pushRawBlockFlag
}
func SetPushRawBlockFlag(b bool) {
	pushRawBlockFlag = b
}
func GetPushBlockTxsFlag() bool {
	return pushBlockTxsFlag
}
func SetPushBlockTxsFlag(b bool) {
	pushBlockTxsFlag = b
}
func SetTxHashMap(txhash string, sessionid string) {
	if ws == nil {
		return
	}
	ws.SetTxHashMap(txhash, sessionid)
}

func PushResult(txHash Uint256, errcode int64, action string, result interface{}) {
	if ws != nil {
		resp := common.ResponsePack(ErrorCode.SUCCESS)
		resp["Result"] = result
		resp["Error"] = errcode
		resp["Action"] = action
		resp["Desc"] = ErrorCode.ErrMap[resp["Error"].(errors.ErrCode)]
		ws.PushTxResult(BytesToHexString(txHash.ToArrayReverse()), resp)
	}
}

func PushSmartCodeInvokeResult(txHash Uint256, errcode errors.ErrCode, result interface{}) {
	if ws == nil {
		return
	}
	resp := common.ResponsePack(ErrorCode.SUCCESS)
	var Result = make(map[string]interface{})
	txHashStr := BytesToHexString(txHash.ToArray())
	Result["TxHash"] = txHashStr
	Result["ExecResult"] = result

	resp["Result"] = Result
	resp["Action"] = "sendsmartcodeinvoke"
	resp["Error"] = errcode
	resp["Desc"] = ErrorCode.ErrMap[errcode]
	ws.PushTxResult(txHashStr, resp)
}
func PushBlock(v interface{}) {
	if ws == nil {
		return
	}
	resp := common.ResponsePack(ErrorCode.SUCCESS)
	if block, ok := v.(*ledger.Block); ok {
		if pushRawBlockFlag {
			w := bytes.NewBuffer(nil)
			block.Serialize(w)
			resp["Result"] = BytesToHexString(w.Bytes())
		} else {
			resp["Result"] = common.GetBlockInfo(block)
		}
		resp["Action"] = "sendrawblock"
		ws.PushResult(resp)
	}
}
func PushBlockTransactions(v interface{}) {
	if ws == nil {
		return
	}
	resp := common.ResponsePack(ErrorCode.SUCCESS)
	if block, ok := v.(*ledger.Block); ok {
		if pushBlockTxsFlag {
			resp["Result"] = common.GetBlockTransactions(block)
		}
		resp["Action"] = "sendblocktransactions"
		ws.PushResult(resp)
	}
}
