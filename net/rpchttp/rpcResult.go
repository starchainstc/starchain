package rpchttp

var (
	STCRpcInvalidHash        = responsePacking("invalid hash")
	STCRpcInvalidBlock       = responsePacking("invalid block")
	STCRpcInvalidTransaction = responsePacking("invalid transaction")
	STCRpcInvalidParameter   = responsePacking("invalid parameter")

	STCRpcUnknownBlock       = responsePacking("unknown block")
	STCRpcUnknownTransaction = responsePacking("unknown transaction")

	STCRpcNil           = responsePacking(nil)
	STCRpcUnsupported   = responsePacking("Unsupported")
	STCRpcInternalError = responsePacking("internal error")
	STCRpcIOError       = responsePacking("internal IO error")
	STCRpcAPIError      = responsePacking("internal API error")
	STCRpcSuccess       = responsePacking(true)
	STCRpcFailed        = responsePacking(false)

	// error code for wallet
	STCRpcWalletAlreadyExists = responsePacking("wallet already exist")
	STCRpcWalletNotExists     = responsePacking("wallet doesn't exist")

	STCRpc = responsePacking
)
