package transaction

import (
	"io"
	"starchain/common"
	"starchain/core/contract/program"
	"starchain/common/serialization"
	"starchain/core/transaction/payload"
	."starchain/errors"
	sig"starchain/core/signature"
	"errors"
	"starchain/common/log"
	"sort"
	"fmt"
	"starchain/core/contract"
	"bytes"
	"crypto/sha256"
)

type TransactionType byte

const(
	BookKeeping    TransactionType = 0x00
	IssueAsset     TransactionType = 0x01
	BookKeeper     TransactionType = 0x02
	LockAsset      TransactionType = 0x03
	PrivacyPayload TransactionType = 0x20
	RegisterAsset  TransactionType = 0x40
	TransferAsset  TransactionType = 0x80
	Record         TransactionType = 0x81
	DeployCode     TransactionType = 0xd0
	InvokeCode     TransactionType = 0xd1
	DataFile       TransactionType = 0x12
)


const (
	// encoded public key length 0x21 || encoded public key (33 bytes) || OP_CHECKSIG(0xac)
	PublickKeyScriptLen = 35

	// signature length(0x40) || 64 bytes signature
	SignatureScriptLen = 65

	// 1byte m || 3 encoded public keys with leading 0x40 (34 bytes * 3) ||
	// 1byte n + 1byte OP_CHECKMULTISIG
	// FIXME: if want to support 1/2 multisig
	MinMultisigCodeLen = 105
)

type TransactionResult map[common.Uint256]common.Fixed64


type Payload interface {
	Data(version byte) []byte
	Serialize(w io.Writer, version byte) error

	Deserialize(r io.Reader, version byte) error
}


var TxStore ILedgerStore

type Transaction struct {
	TxType 		TransactionType
	PayloadVersion	byte
	Payload 	Payload
	Attributes	[]*TxAttribute
	UTXOInputs	[]*UTXOTxInput
	BalanceInputs	[]*BalanceTxInput
	Outputs		[]*TxOutput
	Programs	[]*program.Program


	AssetOutputs 	map[common.Uint256][]*TxOutput
	AssetInputAmount 	map[common.Uint256]common.Fixed64
	AssetOutputAmount 	map[common.Uint256]common.Fixed64

	hash 		*common.Uint256
}


//Serialize the Transaction
func (tx *Transaction) Serialize(w io.Writer) error {

	err := tx.SerializeUnsigned(w)
	if err != nil {
		return NewDetailErr(err, ErrNoCode, "Transaction txSerializeUnsigned Serialize failed.")
	}
	//Serialize  Transaction's programs
	lens := uint64(len(tx.Programs))
	err = serialization.WriteVarUint(w, lens)
	if err != nil {
		return NewDetailErr(err, ErrNoCode, "Transaction WriteVarUint failed.")
	}
	if lens > 0 {
		for _, p := range tx.Programs {
			err = p.Serialize(w)
			if err != nil {
				return NewDetailErr(err, ErrNoCode, "Transaction Programs Serialize failed.")
			}
		}
	}
	return nil
}

//Serialize the Transaction data without contracts
func (tx *Transaction) SerializeUnsigned(w io.Writer) error {
	//txType
	w.Write([]byte{byte(tx.TxType)})
	//PayloadVersion
	w.Write([]byte{tx.PayloadVersion})
	//Payload
	if tx.Payload == nil {
		return errors.New("Transaction Payload is nil.")
	}
	tx.Payload.Serialize(w, tx.PayloadVersion)
	//[]*txAttribute
	err := serialization.WriteVarUint(w, uint64(len(tx.Attributes)))
	if err != nil {
		return NewDetailErr(err, ErrNoCode, "Transaction item txAttribute length serialization failed.")
	}
	if len(tx.Attributes) > 0 {
		for _, attr := range tx.Attributes {
			attr.Serialize(w)
		}
	}
	//[]*UTXOInputs
	err = serialization.WriteVarUint(w, uint64(len(tx.UTXOInputs)))
	if err != nil {
		return NewDetailErr(err, ErrNoCode, "Transaction item UTXOInputs length serialization failed.")
	}
	if len(tx.UTXOInputs) > 0 {
		for _, utxo := range tx.UTXOInputs {
			utxo.Serialize(w)
		}
	}
	// TODO BalanceInputs
	//[]*Outputs
	err = serialization.WriteVarUint(w, uint64(len(tx.Outputs)))
	if err != nil {
		return NewDetailErr(err, ErrNoCode, "Transaction item Outputs length serialization failed.")
	}
	if len(tx.Outputs) > 0 {
		for _, output := range tx.Outputs {
			output.Serialize(w)
		}
	}

	return nil
}

//deserialize the Transaction
func (tx *Transaction) Deserialize(r io.Reader) error {
	// tx deserialize
	err := tx.DeserializeUnsigned(r)
	if err != nil {
		log.Error("Deserialize DeserializeUnsigned:", err)
		return NewDetailErr(err, ErrNoCode, "transaction Deserialize error")
	}

	// tx program
	lens, err := serialization.ReadVarUint(r, 0)
	if err != nil {
		return NewDetailErr(err, ErrNoCode, "transaction tx program Deserialize error")
	}

	programHashes := []*program.Program{}
	if lens > 0 {
		for i := 0; i < int(lens); i++ {
			outputHashes := new(program.Program)
			outputHashes.Deserialize(r)
			programHashes = append(programHashes, outputHashes)
		}
		tx.Programs = programHashes
	}
	return nil
}

func (tx *Transaction) DeserializeUnsigned(r io.Reader) error {
	var txType [1]byte
	_, err := io.ReadFull(r, txType[:])
	if err != nil {
		log.Error("DeserializeUnsigned ReadFull:", err)
		return err
	}
	tx.TxType = TransactionType(txType[0])
	return tx.DeserializeUnsignedWithoutType(r)
}

func (tx *Transaction) DeserializeUnsignedWithoutType(r io.Reader) error {
	var payloadVersion [1]byte
	_, err := io.ReadFull(r, payloadVersion[:])
	tx.PayloadVersion = payloadVersion[0]
	if err != nil {
		log.Error("DeserializeUnsignedWithoutType:", err)
		return err
	}

	//payload
	//tx.Payload.Deserialize(r)
	switch tx.TxType {
	case RegisterAsset:
		tx.Payload = new(payload.RegisterAsset)
	case LockAsset:
		tx.Payload = new(payload.LockAsset)
	case IssueAsset:
		tx.Payload = new(payload.IssueAsset)
	case TransferAsset:
		tx.Payload = new(payload.TransferAsset)
	case BookKeeping:
		tx.Payload = new(payload.BookKeeping)
	case Record:
		tx.Payload = new(payload.Record)
	case BookKeeper:
		tx.Payload = new(payload.BookKeeper)
	case PrivacyPayload:
		tx.Payload = new(payload.PrivacyPayload)
	case DeployCode:
		tx.Payload = new(payload.DeployCode)
	case InvokeCode:
		tx.Payload = new(payload.InvokeCode)
	case DataFile:
		tx.Payload = new(payload.DataFile)
	default:
		return errors.New("[Transaction],invalide transaction type.")
	}
	err = tx.Payload.Deserialize(r, tx.PayloadVersion)
	if err != nil {
		log.Error("tx Payload Deserialize:", err)
		return NewDetailErr(err, ErrNoCode, "Payload Parse error")
	}
	//attributes
	Len, err := serialization.ReadVarUint(r, 0)
	if err != nil {
		log.Error("tx attributes Deserialize:", err)
		return err
	}
	if Len > uint64(0) {
		for i := uint64(0); i < Len; i++ {
			attr := new(TxAttribute)
			err = attr.Deserialize(r)
			if err != nil {
				return err
			}
			tx.Attributes = append(tx.Attributes, attr)
		}
	}
	//UTXOInputs
	Len, err = serialization.ReadVarUint(r, 0)
	if err != nil {
		log.Error("tx UTXOInputs Deserialize:", err)

		return err
	}
	if Len > uint64(0) {
		for i := uint64(0); i < Len; i++ {
			utxo := new(UTXOTxInput)
			err = utxo.Deserialize(r)
			if err != nil {
				return err
			}
			tx.UTXOInputs = append(tx.UTXOInputs, utxo)
		}
	}
	//TODO balanceInputs
	//Outputs
	Len, err = serialization.ReadVarUint(r, 0)
	if err != nil {
		return err
	}
	if Len > uint64(0) {
		for i := uint64(0); i < Len; i++ {
			output := new(TxOutput)
			output.Deserialize(r)

			tx.Outputs = append(tx.Outputs, output)
		}
	}
	return nil
}






func (tx *Transaction) GetMergedAssetIDValueFromReference() (TransactionResult, error) {
	reference, err := tx.GetReference()
	if err != nil {
		return nil, err
	}
	var result = make(map[common.Uint256]common.Fixed64)
	for _, v := range reference {
		amout, ok := result[v.AssetID]
		if ok {
			result[v.AssetID] = amout + v.Value
		} else {
			result[v.AssetID] = v.Value
		}
	}
	return result, nil
}


func (tx *Transaction) GetProgramHashes()([]common.Uint160,error){
	if tx == nil {
		return []common.Uint160{},errors.New("getprogramhash tx is nil")
	}
	hashs := []common.Uint160{}
	uniqHashs := []common.Uint160{}
	referenceWithUTXo_output,err := tx.GetReference()
	if err != nil{
		return nil,NewDetailErr(err,ErrNoCode,"tx getReference error")
	}
	for _,output := range referenceWithUTXo_output{
		programHash := output.ProgramHash
		hashs = append(hashs, programHash)
	}
	for _, attribute := range tx.Attributes {
		if attribute.Usage == Script {
			dataHash, err := common.Uint160ParseFromBytes(attribute.Data)
			if err != nil {
				return nil, NewDetailErr(errors.New("[Transaction], GetProgramHashes err."), ErrNoCode, "")
			}
			hashs = append(hashs, common.Uint160(dataHash))
		}
	}
	switch tx.TxType {
	case RegisterAsset:
		issuer := tx.Payload.(*payload.RegisterAsset).Issuer
		signatureRedeemScript, err := contract.CreateSignatureRedeemScript(issuer)
		if err != nil {
			return nil, NewDetailErr(err, ErrNoCode, "[Transaction], GetProgramHashes CreateSignatureRedeemScript failed.")
		}

		astHash, err := common.ToCodeHash(signatureRedeemScript)
		if err != nil {
			return nil, NewDetailErr(err, ErrNoCode, "[Transaction], GetProgramHashes ToCodeHash failed.")
		}
		hashs = append(hashs, astHash)
	case LockAsset:
		hashs = append(hashs, tx.Payload.(*payload.LockAsset).ProgramHash)
	case IssueAsset:
		result := tx.GetMergedAssetIDValueFromOutputs()
		if err != nil {
			return nil, NewDetailErr(err, ErrNoCode, "[Transaction], GetTransactionResults failed.")
		}
		for k := range result {
			tx, err := TxStore.GetTransaction(k)
			if err != nil {
				return nil, NewDetailErr(err, ErrNoCode, fmt.Sprintf("[Transaction], GetTransaction failed With AssetID:=%x", k))
			}
			if tx.TxType != RegisterAsset {
				return nil, NewDetailErr(errors.New("[Transaction] error"), ErrNoCode, fmt.Sprintf("[Transaction], Transaction Type ileage With AssetID:=%x", k))
			}

			switch v1 := tx.Payload.(type) {
			case *payload.RegisterAsset:
				hashs = append(hashs, v1.Controller)
			default:
				return nil, NewDetailErr(errors.New("[Transaction] error"), ErrNoCode, fmt.Sprintf("[Transaction], payload is illegal", k))
			}
		}
	case DataFile:
		issuer := tx.Payload.(*payload.DataFile).Issuer
		signatureRedeemScript, err := contract.CreateSignatureRedeemScript(issuer)
		if err != nil {
			return nil, NewDetailErr(err, ErrNoCode, "[Transaction], GetProgramHashes CreateSignatureRedeemScript failed.")
		}

		astHash, err := common.ToCodeHash(signatureRedeemScript)
		if err != nil {
			return nil, NewDetailErr(err, ErrNoCode, "[Transaction], GetProgramHashes ToCodeHash failed.")
		}
		hashs = append(hashs, astHash)
	case TransferAsset:
	case Record:
	case DeployCode:
	case InvokeCode:
		issuer := tx.Payload.(*payload.InvokeCode).ProgramHash
		hashs = append(hashs, issuer)
	case BookKeeper:
		issuer := tx.Payload.(*payload.BookKeeper).Issuer
		signatureRedeemScript, err := contract.CreateSignatureRedeemScript(issuer)
		if err != nil {
			return nil, NewDetailErr(err, ErrNoCode, "[Transaction - BookKeeper], GetProgramHashes CreateSignatureRedeemScript failed.")
		}

		astHash, err := common.ToCodeHash(signatureRedeemScript)
		if err != nil {
			return nil, NewDetailErr(err, ErrNoCode, "[Transaction - BookKeeper], GetProgramHashes ToCodeHash failed.")
		}
		hashs = append(hashs, astHash)
	case PrivacyPayload:
		issuer := tx.Payload.(*payload.PrivacyPayload).EncryptAttr.(*payload.EcdhAes256).FromPubkey
		signatureRedeemScript, err := contract.CreateSignatureRedeemScript(issuer)
		if err != nil {
			return nil, NewDetailErr(err, ErrNoCode, "[Transaction], GetProgramHashes CreateSignatureRedeemScript failed.")
		}

		astHash, err := common.ToCodeHash(signatureRedeemScript)
		if err != nil {
			return nil, NewDetailErr(err, ErrNoCode, "[Transaction], GetProgramHashes ToCodeHash failed.")
		}
		hashs = append(hashs, astHash)
	default:
	}
	//remove dupilicated hashes
	uniq := make(map[common.Uint160]bool)
	for _, v := range hashs {
		uniq[v] = true
	}
	for k := range uniq {
		uniqHashs = append(uniqHashs, k)
	}
	sort.Sort(byProgramHashes(uniqHashs))
	return uniqHashs, nil
}

func (tx *Transaction) SetPrograms(programs []*program.Program) {
	tx.Programs = programs
}

func (tx *Transaction) GetPrograms() []*program.Program {
	return tx.Programs
}

func (tx *Transaction) GetOutputHashes() ([]common.Uint160, error) {
	//TODO: implement Transaction.GetOutputHashes()

	return []common.Uint160{}, nil
}

func (tx *Transaction) GenerateAssetMaps() {
	//TODO: implement Transaction.GenerateAssetMaps()
}

func (tx *Transaction) GetMessage() []byte {
	return sig.GetHashData(tx)
}

func (tx *Transaction) ToArray() []byte {
	b := new(bytes.Buffer)
	tx.Serialize(b)
	return b.Bytes()
}

func (tx *Transaction) Hash() common.Uint256 {
	if tx.hash == nil {
		d := sig.GetHashData(tx)
		temp := sha256.Sum256([]byte(d))
		f := common.Uint256(sha256.Sum256(temp[:]))
		tx.hash = &f
	}
	return *tx.hash

}

func (tx *Transaction) SetHash(hash common.Uint256) {
	tx.hash = &hash
}

func (tx *Transaction) Type() common.InventoryType {
	return common.TRANSACTION
}
func (tx *Transaction) Verify() error {
	//TODO: Verify()
	return nil
}

func (tx *Transaction) GetReference() (map[*UTXOTxInput]*TxOutput, error) {
	if tx.TxType == RegisterAsset {
		return nil, nil
	}
	//UTXO input /  Outputs
	reference := make(map[*UTXOTxInput]*TxOutput)
	// Key index，v UTXOInput
	for _, utxo := range tx.UTXOInputs {
		transaction, err := TxStore.GetTransaction(utxo.ReferTxID)
		if err != nil {
			return nil, NewDetailErr(err, ErrNoCode, "[Transaction], GetReference failed.")
		}
		index := utxo.ReferTxOutputIndex
		reference[utxo] = transaction.Outputs[index]
	}
	return reference, nil
}
func (tx *Transaction) GetTransactionResults() (TransactionResult, error) {
	result := make(map[common.Uint256]common.Fixed64)
	outputResult := tx.GetMergedAssetIDValueFromOutputs()
	InputResult, err := tx.GetMergedAssetIDValueFromReference()
	if err != nil {
		return nil, err
	}
	//calc the balance of input vs output
	for outputAssetid, outputValue := range outputResult {
		if inputValue, ok := InputResult[outputAssetid]; ok {
			result[outputAssetid] = inputValue - outputValue
		} else {
			result[outputAssetid] -= outputValue
		}
	}
	for inputAssetid, inputValue := range InputResult {
		if _, exist := result[inputAssetid]; !exist {
			result[inputAssetid] += inputValue
		}
	}
	return result, nil
}

func (tx *Transaction) GetMergedAssetIDValueFromOutputs() TransactionResult {
	var result = make(map[common.Uint256]common.Fixed64)
	for _, v := range tx.Outputs {
		amout, ok := result[v.AssetID]
		if ok {
			result[v.AssetID] = amout + v.Value
		} else {
			result[v.AssetID] = v.Value
		}
	}
	return result
}




func ParseMultisigTransactionCode(code []byte) []common.Uint160{
	if len(code) < MinMultisigCodeLen{
		log.Error("code not enough in multisig transaction detected")
		return nil
	}
	//remove the last 2
	code = code[:len(code)-2]
	//remove the first
	code = code[1:]
	if len(code)%(PublickKeyScriptLen-1) != 0{
		log.Error("invalid code in multisig transaction detected")
		return nil
	}
	var programHash []common.Uint160
	for i:=0;i<len(code);i+=PublickKeyScriptLen-1{
		script := make([]byte,PublickKeyScriptLen-1)
		copy(script,code[i:i+PublickKeyScriptLen-1])
		script = append(script,0xac)
		hash,_ := common.ToCodeHash(script)
		programHash = append(programHash,hash)
	}
	return programHash
}

func (tx *Transaction) ParseTransactionCode() []common.Uint160{
	code := make([]byte,len(tx.Programs[0].Code))
	copy(code,tx.Programs[0].Code)
	return ParseMultisigTransactionCode(code)
}

func (tx *Transaction) ParseTransactionSig()(havesig,needsig int,err error){
	if len(tx.Programs) < 1 {
		return -1,-1,errors.New("missing transaction program")
	}
	x := len(tx.Programs[0].Parameter)/SignatureScriptLen
	y := len(tx.Programs[0].Parameter)%SignatureScriptLen
	return x,y,nil
}


func (tx *Transaction) AppendNewSignature(sig []byte) error {
	if len(tx.Programs) < 0{
		return errors.New("missing transaction program")
	}
	newsig := []byte{}
	newsig = append(newsig,byte(len(sig)))
	newsig = append(newsig,sig...)
	havesig,_,err := tx.ParseTransactionSig()
	if err != nil {
		return err
	}
	existedsigs := tx.Programs[0].Parameter[0 : havesig*SignatureScriptLen]
	leftsigs := tx.Programs[0].Parameter[havesig*SignatureScriptLen+1:]

	tx.Programs[0].Parameter = nil
	tx.Programs[0].Parameter = append(tx.Programs[0].Parameter, existedsigs...)
	tx.Programs[0].Parameter = append(tx.Programs[0].Parameter, newsig...)
	tx.Programs[0].Parameter = append(tx.Programs[0].Parameter, leftsigs...)

	return nil
}




type byProgramHashes []common.Uint160

func (a byProgramHashes) Len() int      { return len(a) }
func (a byProgramHashes) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a byProgramHashes) Less(i, j int) bool {
	if a[i].CompareTo(a[j]) > 0 {
		return false
	} else {
		return true
	}
}

