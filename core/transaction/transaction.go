package transaction

import (
	"io"
	"starchain/common"
	"starchain/core/contract/program"
	"starchain/common/serialization"
	"starchain/core/transaction/payload"
	."starchain/errors"
	"errors"
	"starchain/common/log"
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

