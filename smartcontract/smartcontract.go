package smartcontract

import (
	. "starchain/common"
	"starchain/core/contract"
	"starchain/smartcontract/types"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
)

type Engine interface {
	Create(caller Uint160,code []byte) ([]byte,error)
	Call(caller Uint160,codeHash Uint160,input []byte) ([]byte,error)
}

type SmartContract struct {
	Engine 	Engine
	Code 	[]byte
	Input 	[]byte
	ParameterType []contract.ContractParameterType
	Caller	Uint160
	CodeHash Uint160
	VMType	types.VmType
	ReturnType contract.ContractParameterType
}

type Context struct {
	Language	types.LangType
	Caller 		Uint160
	StateMachine	*serivce.StateMachine
}
