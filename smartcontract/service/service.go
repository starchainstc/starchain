package service

import (
	. "starchain/common"
)

type AccoutInfo struct {
	ProgramHash	string
	IsForze		bool
	Blanace		map[string]Fixed64
}

