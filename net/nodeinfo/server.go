package nodeinfo

import (
	"starchain/net/protocol"
	"fmt"
	"starchain/common/config"
	"net/http"
	"starchain/core/ledger"
	"strconv"
	"sort"
	"html/template"
)

type Info struct {
	NodeVersion	string
	BlockHeight	uint32
	NeighborCnt	int
	Neighbors	[]NgbNodeInfo
	HttpRestPort	int
	HttpWsPort 	int
	HttpJsonPort 	int
	NodePort	int
	NodeId		string
	NodeType 	string
}

const (
	verifyNode  = "Verify Node"
	serviceNode = "Service Node"
)

var node protocol.Noder

func newNgbNodeInfo(ngbId string, ngbType string, ngbAddr string, httpInfoAddr string, httpInfoPort int, httpInfoStart bool) *NgbNodeInfo {
	return &NgbNodeInfo{NgbId: ngbId, NgbType: ngbType, NgbAddr: ngbAddr, HttpInfoAddr: httpInfoAddr,
		HttpInfoPort: httpInfoPort, HttpInfoStart: httpInfoStart}
}

func initPageInfo(blockHeight uint32, curNodeType string, ngbrCnt int, ngbrsInfo []NgbNodeInfo) (*Info, error) {
	id := fmt.Sprintf("0x%x", node.GetID())
	return &Info{NodeVersion: config.Version, BlockHeight: blockHeight,
		NeighborCnt: ngbrCnt, Neighbors: ngbrsInfo,
		HttpRestPort: config.Parameters.HttpRestPort,
		HttpWsPort:   config.Parameters.HttpWsPort,
		HttpJsonPort: config.Parameters.HttpJsonPort,
		NodePort:     config.Parameters.NodePort,
		NodeId:       id, NodeType: curNodeType}, nil
}


var templates = template.Must(template.New("info").Parse(page))

func viewHander(w http.ResponseWriter,r *http.Request){
	var ngbrNodersInfo []NgbNodeInfo
	var ngbId string
	var ngbAddr string
	var ngbType string
	var ngbInfoPort int
	var ngbInfoState bool
	var ngbHttpInfoAddr string

	curNodeType := serviceNode
	bookKeepers,_,_ := ledger.DefaultLedger.Store.GetBookKeeperList()
	bookKeeperLen := len(bookKeepers)
	///if the node is verify node then
	for i := 0; i < bookKeeperLen;i++{
		if node.GetPubKey().X.Cmp(bookKeepers[i].X) == 0{
			curNodeType = verifyNode
			break
		}
	}
	ngbNodes := node.GetNeighborNoder()
	ngbNodeLen := len(ngbNodes)
	for i:=0;i<ngbNodeLen;i++{
		ngbType = serviceNode
		//find the verify node in neighbor node
		for j:=0;j<bookKeeperLen;j++{
			if ngbNodes[i].GetPubKey().X.Cmp(bookKeepers[j].X) == 0{
				ngbType = verifyNode
				break;
			}
		}
		ngbAddr = ngbNodes[i].GetAddr()
		ngbInfoPort = ngbNodes[i].GetHttpInfoPort()
		ngbInfoState = ngbNodes[i].GetHttpInfoState()
		ngbHttpInfoAddr = ngbAddr+":"+strconv.Itoa(int(ngbInfoPort))
		ngbId = fmt.Sprintf("0x%x",ngbNodes[i].GetID())
		ngbInfo := newNgbNodeInfo(ngbId,ngbType,ngbAddr,ngbHttpInfoAddr,ngbInfoPort,ngbInfoState)
		ngbrNodersInfo = append(ngbrNodersInfo,*ngbInfo)
	}
	sort.Sort(NgbNodeInfoSlice(ngbrNodersInfo))
	blockHeight := ledger.DefaultLedger.Blockchain.BlockHeight
	pageInfo,err := initPageInfo(blockHeight,curNodeType,ngbNodeLen,ngbrNodersInfo)
	if err != nil {
		http.Redirect(w,r,"/info",http.StatusFound)
		return
	}
	err = templates.ExecuteTemplate(w, "info", pageInfo)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func StartServer(n protocol.Noder){
	node = n
	port := int(config.Parameters.HttpInfoPort)
	http.HandleFunc("/info",viewHander)
	http.ListenAndServe(":"+strconv.Itoa(port),nil)
}