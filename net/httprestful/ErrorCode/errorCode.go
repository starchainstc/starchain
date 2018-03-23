package ErrorCode

import ."starchain/errors"

const(
	SUCCESS		ErrCode = 0
	SESSION_EXPIRED	ErrCode = 41001
	SERVICE_CEILING    ErrCode = 41002
	ILLEGAL_DATAFORMAT ErrCode = 41003
	OAUTH_TIMEOUT      ErrCode = 41004

	INVALID_METHOD ErrCode = 42001
	INVALID_PARAMS ErrCode = 42002
	INVALID_TOKEN  ErrCode = 42003

	INVALID_TRANSACTION ErrCode = 43001
	INVALID_ASSET       ErrCode = 43002
	INVALID_BLOCK       ErrCode = 43003

	UNKNOWN_TRANSACTION ErrCode = 44001
	UNKNOWN_ASSET       ErrCode = 44002
	UNKNOWN_BLOCK       ErrCode = 44003
	UNKNOWN_RECORD      ErrCode = 44004

	INVALID_VERSION ErrCode = 45001
	INTERNAL_ERROR  ErrCode = 45002

	OAUTH_INVALID_APPID    ErrCode = 46001
	OAUTH_INVALID_CHECKVAL ErrCode = 46002
	SMARTCODE_ERROR        ErrCode = 47001
)

const (
	RESP_ERROR string = "Error"
	RESP_RESULT string = "Result"
	RESP_HEIGHT string = "Height"
	RESP_ACTION string = "Action"
	RESP_HASH string = "Hash"
	RESP_DESC string = "Desc"
	RESP_VERSION string = "Version"

)


var ErrMap = map[ErrCode]string{
	SUCCESS:            "SUCCESS",
	SESSION_EXPIRED:    "SESSION EXPIRED",
	SERVICE_CEILING:    "SERVICE CEILING",
	ILLEGAL_DATAFORMAT: "ILLEGAL DATAFORMAT",
	OAUTH_TIMEOUT:      "CONNECT TO OAUTH TIMEOUT",

	INVALID_METHOD: "INVALID METHOD",
	INVALID_PARAMS: "INVALID PARAMS",
	INVALID_TOKEN:  "VERIFY TOKEN ERROR",

	INVALID_TRANSACTION: "INVALID TRANSACTION",
	INVALID_ASSET:       "INVALID ASSET",
	INVALID_BLOCK:       "INVALID BLOCK",

	UNKNOWN_TRANSACTION: "UNKNOWN TRANSACTION",
	UNKNOWN_ASSET:       "UNKNOWN ASSET",
	UNKNOWN_BLOCK:       "UNKNOWN BLOCK",
	UNKNOWN_RECORD:      "UNKNOWN RECORD",

	INVALID_VERSION:                "INVALID VERSION",
	INTERNAL_ERROR:                 "INTERNAL ERROR",
	SMARTCODE_ERROR:                "SMARTCODE EXEC ERROR",
	ErrDuplicateInput:       "INTERNAL ERROR, ErrDuplicateInput",
	ErrAssetPrecision:       "INTERNAL ERROR, ErrAssetPrecision",
	ErrTransactionBalance:   "INTERNAL ERROR, ErrTransactionBalance",
	ErrAttributeProgram:     "INTERNAL ERROR, ErrAttributeProgram",
	ErrTransactionContracts: "INTERNAL ERROR, ErrTransactionContracts",
	ErrTransactionPayload:   "INTERNAL ERROR, ErrTransactionPayload",
	ErrDoubleSpend:          "INTERNAL ERROR, ErrDoubleSpend",
	ErrTxHashDuplicate:      "INTERNAL ERROR, ErrTxHashDuplicate",
	ErrStateUpdaterVaild:    "INTERNAL ERROR, ErrStateUpdaterVaild",
	ErrSummaryAsset:         "INTERNAL ERROR, ErrSummaryAsset",
	ErrLockedAsset:          "INTERNAL ERROR, ErrLockedAsset",
	ErrDuplicateLockAsset:   "INTERNAL ERROR, ErrDuplicateLockAsset",
	ErrXmitFail:             "INTERNAL ERROR, ErrXmitFail",
}