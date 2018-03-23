package code

import (
	."starchain/core/contract"
	."starchain/common"
)

type ICode interface {
	GetCode() []byte
	GetParameterTypes() []ContractParameterType
	GetReturnTypes() []ContractParameterType

	CodeHash() Uint160
}
