package common

import (
	"starchain/net/protocol"
	."starchain/common"
	"starchain/net/httprestful/ErrorCode"
	. "starchain/net/rpchttp"
	"starchain/core/ledger"
	"strconv"
	"bytes"
	"fmt"
	"starchain/smartcontract/states"
	tx"starchain/core/transaction"
	"starchain/errors"
	"starchain/util"
	"encoding/hex"
)

var node protocol.Noder

const TLSPORT = 443

const(
	CMD_ASSETID string = "Assetid"
)

type ApiServer interface {
	Start() error
	Stop()
}

type balanceRes struct{
	AssetId string
	Value string
}


func SetNode(n protocol.Noder){
	node = node
}

func GetConnectionCount(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(ErrorCode.SUCCESS)
	if node != nil {
		resp[ErrorCode.RESP_RESULT] =  node.GetConnectionCnt()
	}
	return resp
}

func GetBlockHeight(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(ErrorCode.SUCCESS)
	resp[ErrorCode.RESP_RESULT] = ledger.DefaultLedger.Blockchain.BlockHeight
	return resp
}

func GetBlockHash(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(ErrorCode.SUCCESS)
	param := cmd[ErrorCode.RESP_HEIGHT].(string)
	if len(param) == 0 {
		resp[ErrorCode.RESP_ERROR] = ErrorCode.INVALID_PARAMS
		return resp
	}
	height, err := strconv.ParseInt(param, 10, 64)
	if err != nil {
		resp[ErrorCode.RESP_ERROR] = ErrorCode.INVALID_PARAMS
		return resp
	}
	hash, err := ledger.DefaultLedger.Store.GetBlockHash(uint32(height))
	if err != nil {
		resp[ErrorCode.RESP_ERROR] = ErrorCode.INVALID_PARAMS
		return resp
	}
	resp[ErrorCode.RESP_RESULT] = BytesToHexString(hash.ToArrayReverse())
	return resp

}

func GetTotalIssued(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(ErrorCode.SUCCESS)
	assetid, ok := cmd[CMD_ASSETID].(string)
	if !ok {
		resp[ErrorCode.RESP_ERROR] = ErrorCode.INVALID_PARAMS
		return resp
	}
	var assetHash Uint256

	aid,err := HexStringToBytesReverse(assetid)
	if err != nil {
		resp[ErrorCode.RESP_ERROR] = ErrorCode.INVALID_PARAMS
		return resp
	}
	if err := assetHash.Deserialize(bytes.NewReader(aid)); err != nil {
		resp[ErrorCode.RESP_ERROR] = ErrorCode.INVALID_PARAMS
		return resp
	}
	amount, err := ledger.DefaultLedger.Store.GetQuantityIssued(assetHash)
	if err != nil {
		resp[ErrorCode.RESP_ERROR] = ErrorCode.INVALID_PARAMS
		return resp
	}
	resp[ErrorCode.RESP_RESULT] = amount.String()
	return resp

}

func GetBlockInfo(block *ledger.Block) BlockInfo {
	hash := block.Hash()
	blockHead := &BlockHead{
		Version:          block.Blockdata.Version,
		PrevBlockHash:    BytesToHexString(block.Blockdata.PrevBlockHash.ToArrayReverse()),
		TransactionsRoot: BytesToHexString(block.Blockdata.TransactionsRoot.ToArrayReverse()),
		Timestamp:        block.Blockdata.Timestamp,
		Height:           block.Blockdata.Height,
		ConsensusData:    block.Blockdata.ConsensusData,
		NextBookKeeper:   BytesToHexString(block.Blockdata.NextBookKeeper.ToArrayReverse()),
		Program: ProgramInfo{
			Code:      BytesToHexString(block.Blockdata.Program.Code),
			Parameter: BytesToHexString(block.Blockdata.Program.Parameter),
		},
		Hash: BytesToHexString(hash.ToArrayReverse()),
	}

	trans := make([]*Transactions, len(block.Transactions))
	for i := 0; i < len(block.Transactions); i++ {
		trans[i] = TransArryByteToHexString(block.Transactions[i])
	}

	b := BlockInfo{
		Hash:         BytesToHexString(hash.ToArrayReverse()),
		BlockData:    blockHead,
		Transactions: trans,
	}
	return b
}

func GetBlockTransactions(block *ledger.Block) interface{} {
	trans := make([]string, len(block.Transactions))
	for i := 0; i < len(block.Transactions); i++ {
		h := block.Transactions[i].Hash()
		trans[i] = BytesToHexString(h.ToArrayReverse())
	}
	hash := block.Hash()
	type BlockTransactions struct {
		Hash         string
		Height       uint32
		Transactions []string
	}
	b := BlockTransactions{
		Hash:         BytesToHexString(hash.ToArrayReverse()),
		Height:       block.Blockdata.Height,
		Transactions: trans,
	}
	return b
}

func getBlock(hash Uint256, getTxBytes bool) (interface{}, errors.ErrCode) {
	block, err := ledger.DefaultLedger.Store.GetBlock(hash)
	if err != nil {
		return "", ErrorCode.UNKNOWN_BLOCK
	}
	if getTxBytes {
		w := bytes.NewBuffer(nil)
		block.Serialize(w)
		return BytesToHexString(w.Bytes()), ErrorCode.SUCCESS
	}
	return GetBlockInfo(block), ErrorCode.SUCCESS
}

func GetBlockByHash(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(ErrorCode.SUCCESS)
	param := cmd[ErrorCode.RESP_HASH].(string)
	if len(param) == 0 {
		resp[ErrorCode.RESP_RESULT] = ErrorCode.INVALID_PARAMS
		return resp
	}
	var getTxBytes bool = false
	if raw, ok := cmd["Raw"].(string); ok && raw == "1" {
		getTxBytes = true
	}
	var hash Uint256
	hex, err := HexStringToBytesReverse(param)
	if err != nil {
		resp[ErrorCode.RESP_ERROR] = ErrorCode.INVALID_PARAMS
		return resp
	}
	if err := hash.Deserialize(bytes.NewReader(hex)); err != nil {
		resp[ErrorCode.RESP_ERROR] = ErrorCode.INVALID_TRANSACTION
		return resp
	}

	resp[ErrorCode.RESP_RESULT], resp[ErrorCode.RESP_ERROR] = getBlock(hash, getTxBytes)

	return resp
}


func GetBlockTxsByHeight(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(ErrorCode.SUCCESS)

	param := cmd[ErrorCode.RESP_HEIGHT].(string)
	if len(param) == 0 {
		resp[ErrorCode.RESP_ERROR] = ErrorCode.INVALID_PARAMS
		return resp
	}
	height, err := strconv.ParseInt(param, 10, 64)
	if err != nil {
		resp[ErrorCode.RESP_ERROR] = ErrorCode.INVALID_PARAMS
		return resp
	}
	index := uint32(height)
	hash, err := ledger.DefaultLedger.Store.GetBlockHash(index)
	if err != nil {
		resp[ErrorCode.RESP_ERROR] = ErrorCode.UNKNOWN_BLOCK
		return resp
	}
	block, err := ledger.DefaultLedger.Store.GetBlock(hash)
	if err != nil {
		resp[ErrorCode.RESP_ERROR] = ErrorCode.UNKNOWN_BLOCK
		return resp
	}
	resp[ErrorCode.RESP_RESULT] = GetBlockTransactions(block)
	return resp
}


func GetBlockByHeight(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(ErrorCode.SUCCESS)

	param := cmd[ErrorCode.RESP_HEIGHT].(string)
	if len(param) == 0 {
		resp[ErrorCode.RESP_ERROR] = ErrorCode.INVALID_PARAMS
		return resp
	}
	var getTxBytes bool = false
	if raw, ok := cmd["Raw"].(string); ok && raw == "1" {
		getTxBytes = true
	}
	height, err := strconv.ParseInt(param, 10, 64)
	if err != nil {
		resp[ErrorCode.RESP_ERROR] = ErrorCode.INVALID_PARAMS
		return resp
	}
	index := uint32(height)
	hash, err := ledger.DefaultLedger.Store.GetBlockHash(index)
	if err != nil {
		resp[ErrorCode.RESP_ERROR] = ErrorCode.UNKNOWN_BLOCK
		return resp
	}
	resp[ErrorCode.RESP_RESULT], resp[ErrorCode.RESP_ERROR] = getBlock(hash, getTxBytes)
	return resp
}


func GetAssetByHash(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(ErrorCode.SUCCESS)

	str := cmd[ErrorCode.RESP_HASH].(string)
	hex, err := HexStringToBytesReverse(str)
	if err != nil {
		resp[ErrorCode.RESP_ERROR] = ErrorCode.INVALID_PARAMS
		return resp
	}
	var hash Uint256
	err = hash.Deserialize(bytes.NewReader(hex))
	if err != nil {
		resp[ErrorCode.RESP_ERROR] = ErrorCode.INVALID_ASSET
		return resp
	}
	asset, err := ledger.DefaultLedger.Store.GetAsset(hash)
	if err != nil {
		resp[ErrorCode.RESP_ERROR] = ErrorCode.UNKNOWN_ASSET
		return resp
	}
	if raw, ok := cmd["Raw"].(string); ok && raw == "1" {
		w := bytes.NewBuffer(nil)
		asset.Serialize(w)
		resp[ErrorCode.RESP_RESULT] = BytesToHexString(w.Bytes())
		return resp
	}
	resp[ErrorCode.RESP_RESULT] = asset
	return resp
}


func GetBalanceByAddr(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(ErrorCode.SUCCESS)
	addr, ok := cmd["Addr"].(string)
	if !ok {
		resp[ErrorCode.RESP_ERROR] = ErrorCode.INVALID_PARAMS
		return resp
	}
	var programHash Uint160
	programHash, err := ToScriptHash(addr)
	if err != nil {
		resp[ErrorCode.RESP_ERROR] = ErrorCode.INVALID_PARAMS
		return resp
	}
	unspends, err := ledger.DefaultLedger.Store.GetUnspentsFromProgramHash(programHash)
	var balances = make(map[Uint256]Fixed64)
	for assetId, u := range unspends {
		var balance Fixed64 = 0
		for _, v := range u {
			balance = balance + v.Value
		}
		balances[assetId] = balance
	}
	var balRes []balanceRes
	for k,v := range balances{
		balRes = append(balRes,balanceRes{AssetId:hex.EncodeToString(k.ToArrayReverse()),Value:v.String()})
	}
	resp[ErrorCode.RESP_RESULT] = balRes
	return resp
}


func GetLockedAsset(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(ErrorCode.SUCCESS)
	addr, a := cmd["Addr"].(string)
	assetid, k := cmd["Assetid"].(string)
	if !a || !k {
		resp[ErrorCode.RESP_ERROR] = ErrorCode.INVALID_PARAMS
		return resp
	}
	var programHash Uint160
	programHash, err := ToScriptHash(addr)
	if err != nil {
		resp[ErrorCode.RESP_ERROR] = ErrorCode.INVALID_PARAMS
		return resp
	}
	tmpID, err := HexStringToBytesReverse(assetid)
	if err != nil {
		resp[ErrorCode.RESP_ERROR] = ErrorCode.INVALID_PARAMS
		return resp
	}
	asset, err := Uint256ParseFromBytes(tmpID)
	if err != nil {
		resp[ErrorCode.RESP_ERROR] = ErrorCode.INVALID_PARAMS
		return resp
	}
	type locked struct {
		Lock   uint32
		Unlock uint32
		Amount string
	}
	ret := []*locked{}
	lockedAsset, _ := ledger.DefaultLedger.Store.GetLockedFromProgramHash(programHash, asset)
	for _, v := range lockedAsset {
		a := &locked{
			Lock:   v.Lock,
			Unlock: v.Unlock,
			Amount: v.Amount.String(),
		}
		ret = append(ret, a)
	}
	resp[ErrorCode.RESP_RESULT] = ret

	return resp
}

func GetBalanceByAsset(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(ErrorCode.SUCCESS)
	addr, ok := cmd["Addr"].(string)
	assetid, k := cmd["Assetid"].(string)
	if !ok || !k {
		resp[ErrorCode.RESP_ERROR] = ErrorCode.INVALID_PARAMS
		return resp
	}
	var programHash Uint160
	programHash, err := ToScriptHash(addr)
	if err != nil {
		resp[ErrorCode.RESP_ERROR] = ErrorCode.INVALID_PARAMS
		return resp
	}
	unspends, err := ledger.DefaultLedger.Store.GetUnspentsFromProgramHash(programHash)
	var balance Fixed64 = 0
	for k, u := range unspends {
		assid := BytesToHexString(k.ToArrayReverse())
		for _, v := range u {
			if assetid == assid {
				balance = balance + v.Value
			}
		}
	}
	resp[ErrorCode.RESP_RESULT] = balance.String()
	return resp
}
func GetUnspends(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(ErrorCode.SUCCESS)
	addr, ok := cmd["Addr"].(string)
	if !ok {
		resp[ErrorCode.RESP_ERROR] = ErrorCode.INVALID_PARAMS
		return resp
	}
	var programHash Uint160

	programHash, err := ToScriptHash(addr)
	if err != nil {
		resp[ErrorCode.RESP_ERROR] = ErrorCode.INVALID_PARAMS
		return resp
	}
	type UTXOUnspentInfo struct {
		Txid  string
		Index uint32
		Value string
	}
	type Result struct {
		AssetId   string
		AssetName string
		Utxo      []UTXOUnspentInfo
	}
	var results []Result
	unspends, err := ledger.DefaultLedger.Store.GetUnspentsFromProgramHash(programHash)

	for k, u := range unspends {
		assetid := BytesToHexString(k.ToArrayReverse())
		asset, err := ledger.DefaultLedger.Store.GetAsset(k)
		if err != nil {
			resp[ErrorCode.RESP_ERROR] = ErrorCode.INTERNAL_ERROR
			return resp
		}
		var unspendsInfo []UTXOUnspentInfo
		for _, v := range u {
			unspendsInfo = append(unspendsInfo, UTXOUnspentInfo{BytesToHexString(v.Txid.ToArrayReverse()), v.Index, v.Value.String()})
		}
		results = append(results, Result{assetid, asset.Name, unspendsInfo})
	}
	resp[ErrorCode.RESP_RESULT] = results
	return resp
}
func GetUnspendOutput(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(ErrorCode.SUCCESS)
	addr, ok := cmd["Addr"].(string)
	assetid, k := cmd["Assetid"].(string)
	if !ok || !k {
		resp[ErrorCode.RESP_ERROR] = ErrorCode.INVALID_PARAMS
		return resp
	}

	var programHash Uint160
	var assetHash Uint256
	programHash, err := ToScriptHash(addr)
	if err != nil {
		resp[ErrorCode.RESP_ERROR] = ErrorCode.INVALID_PARAMS
		return resp
	}
	bys, err := HexStringToBytesReverse(assetid)
	if err != nil {
		resp[ErrorCode.RESP_ERROR] = ErrorCode.INVALID_PARAMS
		return resp
	}
	if err := assetHash.Deserialize(bytes.NewReader(bys)); err != nil {
		resp[ErrorCode.RESP_ERROR] = ErrorCode.INVALID_PARAMS
		return resp
	}
	type UTXOUnspentInfo struct {
		Txid  string
		Index uint32
		Value string
	}
	infos, err := ledger.DefaultLedger.Store.GetUnspentFromProgramHash(programHash, assetHash)
	if err != nil {
		resp[ErrorCode.RESP_ERROR] = ErrorCode.INVALID_PARAMS
		resp[ErrorCode.RESP_RESULT] = err
		return resp
	}
	var UTXOoutputs []UTXOUnspentInfo
	for _, v := range infos {
		UTXOoutputs = append(UTXOoutputs, UTXOUnspentInfo{Txid: BytesToHexString(v.Txid.ToArrayReverse()), Index: v.Index, Value: v.Value.String()})
	}
	resp[ErrorCode.RESP_RESULT] = UTXOoutputs
	return resp
}

//Transaction
func GetTransactionByHash(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(ErrorCode.SUCCESS)

	str := cmd["Hash"].(string)
	bys, err := HexStringToBytesReverse(str)
	if err != nil {
		resp[ErrorCode.RESP_ERROR] = ErrorCode.INVALID_PARAMS
		return resp
	}
	var hash Uint256
	err = hash.Deserialize(bytes.NewReader(bys))
	if err != nil {
		resp[ErrorCode.RESP_ERROR] = ErrorCode.INVALID_TRANSACTION
		return resp
	}
	tx, err := ledger.DefaultLedger.Store.GetTransaction(hash)
	if err != nil {
		resp[ErrorCode.RESP_ERROR] = ErrorCode.UNKNOWN_TRANSACTION
		return resp
	}
	if raw, ok := cmd["Raw"].(string); ok && raw == "1" {
		w := bytes.NewBuffer(nil)
		tx.Serialize(w)
		resp[ErrorCode.RESP_RESULT] = BytesToHexString(w.Bytes())
		return resp
	}
	tran := TransArryByteToHexString(tx)
	resp[ErrorCode.RESP_RESULT] = tran
	return resp
}
func SendRawTransaction(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(ErrorCode.SUCCESS)

	str, ok := cmd["Data"].(string)
	if !ok {
		resp[ErrorCode.RESP_ERROR] = ErrorCode.INVALID_PARAMS
		return resp
	}
	bys, err := HexStringToBytes(str)
	if err != nil {
		resp[ErrorCode.RESP_ERROR] = ErrorCode.INVALID_PARAMS
		return resp
	}
	var txn tx.Transaction
	if err := txn.Deserialize(bytes.NewReader(bys)); err != nil {
		resp[ErrorCode.RESP_ERROR] = ErrorCode.INVALID_TRANSACTION
		return resp
	}
	if txn.TxType != tx.TransferAsset {
		resp[ErrorCode.RESP_ERROR] = ErrorCode.INVALID_TRANSACTION
		return resp
	}
	var hash Uint256
	hash = txn.Hash()
	if errCode := VerifyAndSendTx(&txn); errCode != errors.ErrNoError {
		resp[ErrorCode.RESP_ERROR] = int32(errCode)
		return resp
	}
	resp[ErrorCode.RESP_RESULT] = BytesToHexString(hash.ToArrayReverse())
	//TODO 0xd1 -> tx.InvokeCode
	if txn.TxType == 0xd1 {
		if userid, ok := cmd["Userid"].(string); ok && len(userid) > 0 {
			resp["Userid"] = userid
		}
	}
	return resp
}

/**
send to address
 */

func SendToAddress(cmd map[string]interface{}) map[string]interface{}{
	resp := ResponsePack(ErrorCode.SUCCESS)
	var asset, address, value string
	asset = cmd["asset"].(string)
	address = cmd["to"].(string)
	value = cmd["value"].(string)
	if Wallet == nil {
		resp[ErrorCode.RESP_ERROR] = ErrorCode.INVALID_PARAMS
		return resp
	}

	batchOut := util.BatchOut{
		Address: address,
		Value:   value,
	}
	tmp, err := HexStringToBytesReverse(asset)
	if err != nil {
		resp[ErrorCode.RESP_ERROR] = ErrorCode.INVALID_PARAMS
		return resp
	}
	var assetID Uint256
	if err := assetID.Deserialize(bytes.NewReader(tmp)); err != nil {
		resp[ErrorCode.RESP_ERROR] = ErrorCode.INVALID_PARAMS
		return resp
	}
	txn, err := util.MakeTransferTransaction(Wallet, assetID, batchOut)
	if err != nil {
		resp[ErrorCode.RESP_ERROR] = ErrorCode.INVALID_TRANSACTION
		return resp
	}

	if errCode := VerifyAndSendTx(txn); errCode != errors.ErrNoError {
		resp[ErrorCode.RESP_ERROR] = errCode
		return resp
	}
	txHash := txn.Hash()
	resp[ErrorCode.RESP_RESULT] = BytesToHexString(txHash.ToArrayReverse())
	return resp
}


func GetNewAddress(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(ErrorCode.SUCCESS)
	acc,err :=Wallet.CreateAccount()
	if err != nil {
		resp[ErrorCode.RESP_ERROR] = ErrorCode.INTERNAL_ERROR
		return resp
	}
	if err := Wallet.CreateContract(acc); err != nil {
		Wallet.DeleteAccount(acc.ProgramHash)
		resp[ErrorCode.RESP_ERROR] = ErrorCode.INTERNAL_ERROR
		return resp
	}
	addr,err := acc.ProgramHash.ToAddress()
	if err != nil {
		resp[ErrorCode.RESP_ERROR] = ErrorCode.INTERNAL_ERROR
		return resp
	}
	resp[ErrorCode.RESP_RESULT] = addr
	return resp
}


//stateupdate
func GetStateUpdate(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(ErrorCode.SUCCESS)
	namespace, ok := cmd["Namespace"].(string)
	if !ok {
		resp[ErrorCode.RESP_ERROR] = ErrorCode.INVALID_PARAMS
		return resp
	}
	key, ok := cmd["Key"].(string)
	if !ok {
		resp[ErrorCode.RESP_ERROR] = ErrorCode.INVALID_PARAMS
		return resp
	}
	fmt.Println(cmd, namespace, key)
	//TODO get state from store
	return resp
}

func GetContract(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(ErrorCode.SUCCESS)
	str := cmd["Hash"].(string)
	bys, err := HexStringToBytesReverse(str)
	if err != nil {
		resp[ErrorCode.RESP_ERROR] = ErrorCode.INVALID_PARAMS
		return resp
	}
	var hash Uint160
	err = hash.Deserialize(bytes.NewReader(bys))
	if err != nil {
		resp[ErrorCode.RESP_ERROR] = ErrorCode.INVALID_PARAMS
		return resp
	}
	//TODO GetContract from store
	contract, err := ledger.DefaultLedger.Store.GetContract(hash)
	if err != nil {
		resp[ErrorCode.RESP_ERROR] = ErrorCode.INVALID_PARAMS
		return resp
	}
	c := new(states.ContractState)
	b := bytes.NewBuffer(contract)
	c.Deserialize(b)
	var params []int
	for _, v := range c.Code.ParameterTypes {
		params = append(params, int(v))
	}
	codehash := c.Code.CodeHash()
	funcCode := &FunctionCodeInfo{
		Code:           BytesToHexString(c.Code.Code),
		ParameterTypes: params,
		ReturnType:     int(c.Code.ReturnType),
		CodeHash:       BytesToHexString(codehash.ToArrayReverse()),
	}
	programHash := c.ProgramHash
	result := DeployCodeInfo{
		Name:        c.Name,
		Author:      c.Author,
		Email:       c.Email,
		Version:     c.Version,
		Description: c.Description,
		Language:    int(c.Language),
		Code:        new(FunctionCodeInfo),
		ProgramHash: BytesToHexString(programHash.ToArrayReverse()),
	}

	result.Code = funcCode
	resp[ErrorCode.RESP_RESULT] = result
	return resp
}




func ResponsePack(errCode errors.ErrCode) map[string]interface{} {
	resp := map[string]interface{}{
		ErrorCode.RESP_ACTION:  "",
		ErrorCode.RESP_RESULT:  "",
		ErrorCode.RESP_ERROR:   errCode,
		ErrorCode.RESP_DESC:    "",
		ErrorCode.RESP_VERSION: "1.0.0",
	}
	return resp
}