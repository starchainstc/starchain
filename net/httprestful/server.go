package httprestful

import (
	"starchain/net/protocol"
	"starchain/net/httprestful/common"
	"starchain/core/ledger"
	"strconv"
	."starchain/common/config"
	"starchain/events"
	"starchain/net/httprestful/restful"
)

func StartServer(n protocol.Noder){
	common.SetNode(n)
	ledger.DefaultLedger.Blockchain.BCEvents.Subscribe(events.EventBlockPersistCompleted, SendBlock2NoticeServer)
	func() {
		rest := restful.InitRestServer(common.CheckAccessToken)
		go rest.Start()
	}()
}


func SendBlock2NoticeServer(v interface{}) {

	if len(Parameters.NoticeServerUrl) == 0 || !common.CheckPushBlock() {
		return
	}
	go func() {
		req := make(map[string]interface{})
		req["Height"] = strconv.FormatInt(int64(ledger.DefaultLedger.Blockchain.BlockHeight), 10)
		req = common.GetBlockByHeight(req)

		repMsg, _ := common.PostRequest(req, Parameters.NoticeServerUrl)
		if repMsg[""] == nil {
			//TODO
		}
	}()
}
