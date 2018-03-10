package signature

import (
	"starchain/vm/avm/interfaces"
	."starchain/common"
	"starchain/core/contract/program"
	"io"
)

type SignbaleData interface {
	interfaces.ICodeContainer

	GetProgramHashes()([]Uint160)

	SetPrograms([]*program.Program)
	GetPrograms() [] *program.Program
	SerializeUnsigned(io.Writer) error
}