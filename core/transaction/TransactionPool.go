package transaction

type TransactionPool interface {

	//  add a transaction to the pool.
	Add(*Transaction) error

	//returns all transactions that were in the pool.
	Dump() ([]*Transaction, error)
}
