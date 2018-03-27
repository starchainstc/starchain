package consensus

import "starchain/common"

type PolicyLevel byte

const (
	AllowAll PolicyLevel = 0x00
	DenyAll PolicyLevel = 0x01
	AllowList PolicyLevel = 0x02
	DenyList PolicyLevel = 0x03
)


type Policy struct {
	PolicyLevel PolicyLevel
	List []common.Uint160
}

func NewPolicy()  *Policy{
	return &Policy{}
}

func (p *Policy) Refresh(){
	//TODO: Refresh
}

var DefaultPolicy *Policy

func InitPolicy(){
	DefaultPolicy := NewPolicy()
	DefaultPolicy.Refresh()
}
