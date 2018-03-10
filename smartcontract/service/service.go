package service

import ."starchain/common"

type AccoutInfo struct {
	ProgramHash	string
	IsForze		bool
	Blanace		map[string]Fixed64
}

type AssetInfo struct {
	Name 		string
	Precision  byte
	AssetType  byte
	RecordType byte
}

func GetHeaderInfo(header *ledger.Header)